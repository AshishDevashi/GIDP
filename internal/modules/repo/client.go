package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const githubAPIURL = "https://api.github.com"

type githubClient struct {
	token      string
	httpClient *http.Client
}

type githubError struct {
	StatusCode       int
	Message          string
	DocumentationURL string
	RequestID        string
}

type githubRepository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

func (e *githubError) Error() string {
	return fmt.Sprintf("github returned status %d: %s", e.StatusCode, e.Message)
}

func newGitHubClient(token string) *githubClient {
	return &githubClient{
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *githubClient) createRepositoryFromTemplate(ctx context.Context, templateOwner, templateRepo string, req CreateRequest) (githubRepository, error) {
	body, err := json.Marshal(struct {
		Owner       string `json:"owner,omitempty"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Private     bool   `json:"private"`
	}{
		Owner:       req.Organization,
		Name:        req.Name,
		Description: req.Description,
		Private:     req.Private,
	})
	if err != nil {
		return githubRepository{}, err
	}

	endpoint := githubAPIURL + "/repos/" + url.PathEscape(templateOwner) + "/" + url.PathEscape(templateRepo) + "/generate"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return githubRepository{}, err
	}
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	slog.InfoContext(ctx, "creating github repository from template",
		"endpoint", endpoint,
		"template_owner", templateOwner,
		"template_repo", templateRepo,
		"name", req.Name,
		"organization", req.Organization,
		"private", req.Private,
		"token_configured", c.token != "",
	)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.ErrorContext(ctx, "github repository request failed", "endpoint", endpoint, "error", err)
		return githubRepository{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		var errorBody struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
		}
		if err := json.NewDecoder(io.LimitReader(httpResp.Body, 1<<20)).Decode(&errorBody); err != nil {
			errorBody.Message = "unable to decode github error response"
		}

		githubErr := &githubError{
			StatusCode:       httpResp.StatusCode,
			Message:          errorBody.Message,
			DocumentationURL: errorBody.DocumentationURL,
			RequestID:        httpResp.Header.Get("X-GitHub-Request-Id"),
		}
		slog.ErrorContext(ctx, "github rejected repository creation",
			"endpoint", endpoint,
			"status", githubErr.StatusCode,
			"message", githubErr.Message,
			"documentation_url", githubErr.DocumentationURL,
			"github_request_id", githubErr.RequestID,
		)
		return githubRepository{}, githubErr
	}

	var repository githubRepository
	if err := json.NewDecoder(io.LimitReader(httpResp.Body, 1<<20)).Decode(&repository); err != nil {
		return githubRepository{}, err
	}
	slog.InfoContext(ctx, "github repository created from template",
		"full_name", repository.FullName,
		"html_url", repository.HTMLURL,
		"github_request_id", httpResp.Header.Get("X-GitHub-Request-Id"),
	)

	return repository, nil
}

func (c *githubClient) deleteRepository(ctx context.Context, owner, name string) error {
	endpoint := githubAPIURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	slog.InfoContext(ctx, "deleting github repository", "owner", owner, "name", name)
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.ErrorContext(ctx, "github repository delete request failed", "owner", owner, "name", name, "error", err)
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNoContent || httpResp.StatusCode == http.StatusNotFound {
		slog.InfoContext(ctx, "github repository deleted",
			"owner", owner,
			"name", name,
			"status", httpResp.StatusCode,
			"github_request_id", httpResp.Header.Get("X-GitHub-Request-Id"),
		)
		return nil
	}

	var errorBody struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
	}
	if err := json.NewDecoder(io.LimitReader(httpResp.Body, 1<<20)).Decode(&errorBody); err != nil {
		errorBody.Message = "unable to decode github error response"
	}
	githubErr := &githubError{
		StatusCode:       httpResp.StatusCode,
		Message:          errorBody.Message,
		DocumentationURL: errorBody.DocumentationURL,
		RequestID:        httpResp.Header.Get("X-GitHub-Request-Id"),
	}
	slog.ErrorContext(ctx, "github rejected repository deletion",
		"owner", owner,
		"name", name,
		"status", githubErr.StatusCode,
		"message", githubErr.Message,
		"documentation_url", githubErr.DocumentationURL,
		"github_request_id", githubErr.RequestID,
	)
	return githubErr
}
