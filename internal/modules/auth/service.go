package auth

import (
	"context"
	"errors"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/password"
	"github.com/AshishDevashi/GIDP/internal/platform/token"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// DefaultRole is assigned to every self-registered user.
const DefaultRole = "developer"

var (
	ErrEmailTaken         = errors.New("email is already registered")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Service contains the auth module's business logic.
type Service struct {
	repo      *Repository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewService(repo *Repository, jwtSecret string, jwtTTL time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

// Register creates a new user account with the default role and returns an auth token.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	if _, err := s.repo.GetUserByEmail(ctx, req.Email); err == nil {
		return AuthResponse{}, ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return AuthResponse{}, err
	}

	if _, err := s.repo.GetUserByUsername(ctx, req.Username); err == nil {
		return AuthResponse{}, ErrUsernameTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return AuthResponse{}, err
	}

	role, err := s.repo.GetRoleByName(ctx, DefaultRole)
	if err != nil {
		return AuthResponse{}, err
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.CreateUser(ctx, store.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     pgtype.Text{String: req.FullName, Valid: req.FullName != ""},
		RoleID:       role.ID,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return s.buildAuthResponse(user)
}

// Login verifies credentials and returns a fresh auth token.
func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}

	if !password.Verify(user.PasswordHash, req.Password) {
		return AuthResponse{}, ErrInvalidCredentials
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return AuthResponse{}, err
	}

	return s.buildAuthResponse(user)
}

// Me returns the profile of the user identified by a validated JWT subject.
func (s *Service) Me(ctx context.Context, userID string) (UserResponse, error) {
	id, err := uuidutil.Parse(userID)
	if err != nil {
		return UserResponse{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}

	return toUserResponse(user), nil
}

func (s *Service) buildAuthResponse(user store.User) (AuthResponse, error) {
	tok, err := token.Generate(s.jwtSecret, s.jwtTTL, uuidutil.String(user.ID), user.Username, uuidutil.String(user.RoleID))
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: tok, User: toUserResponse(user)}, nil
}

func toUserResponse(user store.User) UserResponse {
	return UserResponse{
		ID:        uuidutil.String(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName.String,
		RoleID:    uuidutil.String(user.RoleID),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Time,
	}
}
