package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"shortking-api/internal/cache"
	"shortking-api/internal/models"
	"shortking-api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("service: invalid credentials")
	ErrEmailTaken         = errors.New("service: email already registered")
	ErrInvalidToken       = errors.New("service: invalid or expired token")
)

const (
	bcryptCost           = 12
	passwordResetTTL     = 1 * time.Hour
	emailVerificationTTL = 24 * time.Hour
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	users         repository.UserRepository
	cache         *cache.Cache
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	c *cache.Cache,
	accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:         users,
		cache:         c,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *AuthService) CountUsers(ctx context.Context) (int64, error) {
	return s.users.CountAll(ctx)
}

func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *AuthService) Signup(ctx context.Context, email, password, displayName string) (*models.User, TokenPair, error) {
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, TokenPair{}, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, TokenPair{}, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, TokenPair{}, ErrEmailTaken
		}
		return nil, TokenPair{}, err
	}

	tokens, err := s.issueTokenPair(ctx, user.ID)
	return user, tokens, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, TokenPair{}, ErrInvalidCredentials
	}
	if err != nil {
		return nil, TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, TokenPair{}, ErrInvalidCredentials
	}

	tokens, err := s.issueTokenPair(ctx, user.ID)
	return user, tokens, err
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := parseToken(s.refreshSecret, refreshToken)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}

	allowed, err := s.cache.IsRefreshTokenAllowed(ctx, claims.UserID.String(), claims.JTI)
	if err != nil {
		return TokenPair{}, err
	}
	if !allowed {
		return TokenPair{}, ErrInvalidToken
	}

	// Rotate: revoke the used refresh token and issue a fresh pair.
	if err := s.cache.RevokeRefreshToken(ctx, claims.UserID.String(), claims.JTI); err != nil {
		return TokenPair{}, err
	}

	return s.issueTokenPair(ctx, claims.UserID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := parseToken(s.refreshSecret, refreshToken)
	if err != nil {
		return nil // already invalid/expired, nothing to revoke
	}
	return s.cache.RevokeRefreshToken(ctx, claims.UserID.String(), claims.JTI)
}

// RequestPasswordReset issues a one-time reset token for the given email, if
// an account with that email exists. It returns an empty token (and no
// error) when the email is unknown, so callers can always respond the same
// way and avoid leaking which emails are registered.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	token, err := randomOpaqueToken()
	if err != nil {
		return "", err
	}
	if err := s.cache.SetPasswordResetToken(ctx, hashOpaqueToken(token), user.ID.String(), passwordResetTTL); err != nil {
		return "", err
	}
	return token, nil
}

// ResetPassword consumes a token issued by RequestPasswordReset and sets a
// new password. The token is single-use: it's burned as soon as it's
// consumed, whether or not the rest of the call succeeds, so a partial
// failure can't be retried with the same token indefinitely.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tokenHash := hashOpaqueToken(token)
	userIDStr, err := s.cache.GetPasswordResetUserID(ctx, tokenHash)
	if errors.Is(err, cache.ErrCacheMiss) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	defer func() { _ = s.cache.DeletePasswordResetToken(ctx, tokenHash) }()

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return s.users.Update(ctx, user)
}

// RequestEmailVerification issues a one-time verification token for the
// given email, if an account with that email exists and isn't already
// verified. Like RequestPasswordReset, it returns an empty token (and no
// error) otherwise, so callers can respond identically regardless of
// whether the email is registered or already verified.
func (s *AuthService) RequestEmailVerification(ctx context.Context, email string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if user.EmailVerifiedAt != nil {
		return "", nil
	}

	token, err := randomOpaqueToken()
	if err != nil {
		return "", err
	}
	if err := s.cache.SetEmailVerificationToken(ctx, hashOpaqueToken(token), user.ID.String(), emailVerificationTTL); err != nil {
		return "", err
	}
	return token, nil
}

// VerifyEmail consumes a token issued by RequestEmailVerification and marks
// the owning user's email verified. Single-use, like ResetPassword.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := hashOpaqueToken(token)
	userIDStr, err := s.cache.GetEmailVerificationUserID(ctx, tokenHash)
	if errors.Is(err, cache.ErrCacheMiss) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	defer func() { _ = s.cache.DeleteEmailVerificationToken(ctx, tokenHash) }()

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	return s.users.Update(ctx, user)
}

func (s *AuthService) ParseAccessToken(token string) (*TokenClaims, error) {
	claims, err := parseToken(s.accessSecret, token)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (TokenPair, error) {
	accessJTI := uuid.NewString()
	accessToken, err := issueToken(s.accessSecret, userID, accessJTI, s.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}

	refreshJTI := uuid.NewString()
	refreshToken, err := issueToken(s.refreshSecret, userID, refreshJTI, s.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.cache.AllowRefreshToken(ctx, userID.String(), refreshJTI, s.refreshTTL); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
