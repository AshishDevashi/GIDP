package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	dockerHubAPIURL = "https://hub.docker.com"
	// Docker Hub JWTs are valid for ~30 minutes; refresh earlier to avoid edge expiry.
	tokenTTL = 25 * time.Minute
)

type dockerHubClient struct {
	username   string
	secret     string
	httpClient *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

type dockerHubError struct {
	StatusCode int
	Message    string
}

func (e *dockerHubError) Error() string {
	return fmt.Sprintf("docker hub returned status %d: %s", e.StatusCode, e.Message)
}

type dockerHubRepository struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

func newDockerHubClient(username, secret string) *dockerHubClient {
	return &dockerHubClient{
		username:   username,
		secret:     secret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *dockerHubClient) configured() bool {
	return c.username != "" && c.secret != ""
}

// accessToken returns a cached Docker Hub JWT, logging in when missing or expired.
func (c *dockerHubClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	body, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: c.username, Password: c.secret})
	if err != nil {
		return "", err
	}

	endpoint := dockerHubAPIURL + "/v2/users/login"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.ErrorContext(ctx, "docker hub login request failed", "error", err)
		return "", err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		hubErr := decodeError(httpResp)
		slog.ErrorContext(ctx, "docker hub rejected login", "status", hubErr.StatusCode, "message", hubErr.Message)
		return "", hubErr
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(io.LimitReader(httpResp.Body, 1<<20)).Decode(&loginResp); err != nil {
		return "", err
	}
	if loginResp.Token == "" {
		return "", &dockerHubError{StatusCode: httpResp.StatusCode, Message: "docker hub login returned an empty token"}
	}

	c.token = loginResp.Token
	c.tokenExpiry = time.Now().Add(tokenTTL)
	return c.token, nil
}

func (c *dockerHubClient) createRepository(ctx context.Context, namespace string, req CreateRequest) (dockerHubRepository, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return dockerHubRepository{}, err
	}

	body, err := json.Marshal(struct {
		Namespace       string `json:"namespace"`
		Name            string `json:"name"`
		Description     string `json:"description,omitempty"`
		FullDescription string `json:"full_description,omitempty"`
		IsPrivate       bool   `json:"is_private"`
		Registry        string `json:"registry"`
	}{
		Namespace:       namespace,
		Name:            req.Name,
		Description:     req.Description,
		FullDescription: req.FullDescription,
		IsPrivate:       req.Private,
		Registry:        "docker",
	})
	if err != nil {
		return dockerHubRepository{}, err
	}

	endpoint := dockerHubAPIURL + "/v2/repositories/"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return dockerHubRepository{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "JWT "+token)

	slog.InfoContext(ctx, "creating docker hub repository",
		"namespace", namespace,
		"name", req.Name,
		"private", req.Private,
	)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.ErrorContext(ctx, "docker hub repository request failed", "namespace", namespace, "name", req.Name, "error", err)
		return dockerHubRepository{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		hubErr := decodeError(httpResp)
		slog.ErrorContext(ctx, "docker hub rejected repository creation",
			"namespace", namespace,
			"name", req.Name,
			"status", hubErr.StatusCode,
			"message", hubErr.Message,
		)
		return dockerHubRepository{}, hubErr
	}

	var repository dockerHubRepository
	if err := json.NewDecoder(io.LimitReader(httpResp.Body, 1<<20)).Decode(&repository); err != nil {
		return dockerHubRepository{}, err
	}
	if repository.Name == "" {
		repository.Name = req.Name
	}
	if repository.Namespace == "" {
		repository.Namespace = namespace
	}

	slog.InfoContext(ctx, "docker hub repository created", "namespace", repository.Namespace, "name", repository.Name)
	return repository, nil
}

func (c *dockerHubClient) deleteRepository(ctx context.Context, namespace, name string) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	endpoint := dockerHubAPIURL + "/v2/repositories/" + url.PathEscape(namespace) + "/" + url.PathEscape(name) + "/"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "JWT "+token)

	slog.InfoContext(ctx, "deleting docker hub repository", "namespace", namespace, "name", name)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.ErrorContext(ctx, "docker hub repository delete request failed", "namespace", namespace, "name", name, "error", err)
		return err
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK, http.StatusAccepted, http.StatusNoContent, http.StatusNotFound:
		return nil
	default:
		hubErr := decodeError(httpResp)
		slog.ErrorContext(ctx, "docker hub rejected repository deletion",
			"namespace", namespace,
			"name", name,
			"status", hubErr.StatusCode,
			"message", hubErr.Message,
		)
		return hubErr
	}
}

func decodeError(resp *http.Response) *dockerHubError {
	var errorBody struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return &dockerHubError{StatusCode: resp.StatusCode, Message: "unable to read docker hub error response"}
	}
	_ = json.Unmarshal(raw, &errorBody)

	message := errorBody.Detail
	if message == "" {
		message = errorBody.Message
	}
	if message == "" && len(errorBody.Errors) > 0 {
		message = errorBody.Errors[0].Message
	}
	if message == "" {
		message = string(bytes.TrimSpace(raw))
	}
	if message == "" {
		message = "docker hub request failed"
	}
	return &dockerHubError{StatusCode: resp.StatusCode, Message: message}
}
