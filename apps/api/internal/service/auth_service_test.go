package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"shortking-api/internal/testsupport"
)

func newTestAuthService(t *testing.T) (*AuthService, *testsupport.FakeUserRepository) {
	t.Helper()
	users := testsupport.NewFakeUserRepository()
	c := testsupport.NewTestCache(t)
	svc := NewAuthService(users, c, "access-secret", "refresh-secret", 15*time.Minute, 30*24*time.Hour)
	return svc, users
}

func TestSignup_HashesPasswordAndIssuesTokens(t *testing.T) {
	ctx := context.Background()
	svc, users := newTestAuthService(t)

	user, tokens, err := svc.Signup(ctx, "ann@example.com", "hunter2", "Ann")

	require.NoError(t, err)
	assert.NotEqual(t, "hunter2", user.PasswordHash, "password must not be stored in plaintext")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("hunter2")))
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	stored, err := users.FindByEmail(ctx, "ann@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, stored.ID)
}

func TestSignup_DuplicateEmailIsRejected(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "dup@example.com", "password1", "First")
	require.NoError(t, err)

	_, _, err = svc.Signup(ctx, "dup@example.com", "password2", "Second")

	assert.ErrorIs(t, err, ErrEmailTaken)
}

func TestLogin_WrongPasswordIsRejected(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "bob@example.com", "correct-password", "Bob")
	require.NoError(t, err)

	_, _, err = svc.Login(ctx, "bob@example.com", "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_UnknownEmailIsRejected(t *testing.T) {
	svc, _ := newTestAuthService(t)

	_, _, err := svc.Login(context.Background(), "nobody@example.com", "whatever")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_SuccessIssuesAccessTokenForTheRightUser(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	user, _, err := svc.Signup(ctx, "carl@example.com", "correct-password", "Carl")
	require.NoError(t, err)

	_, tokens, err := svc.Login(ctx, "carl@example.com", "correct-password")
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
}

func TestRefresh_RotatesTokenAndRejectsTheOldOne(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, tokens, err := svc.Signup(ctx, "dana@example.com", "password1", "Dana")
	require.NoError(t, err)

	rotated, err := svc.Refresh(ctx, tokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, tokens.RefreshToken, rotated.RefreshToken)

	_, err = svc.Refresh(ctx, tokens.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidToken, "a rotated-out refresh token must be rejected")

	_, err = svc.Refresh(ctx, rotated.RefreshToken)
	assert.NoError(t, err, "the newly issued refresh token should still work")
}

func TestParseAccessToken_RejectsGarbage(t *testing.T) {
	svc, _ := newTestAuthService(t)

	_, err := svc.ParseAccessToken("not-a-real-token")

	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestRequestPasswordReset_UnknownEmailReturnsNoTokenAndNoError(t *testing.T) {
	// So the handler can respond identically whether or not the email is
	// registered, instead of leaking that via an error.
	svc, _ := newTestAuthService(t)

	token, err := svc.RequestPasswordReset(context.Background(), "nobody@example.com")

	assert.NoError(t, err)
	assert.Empty(t, token)
}

func TestResetPassword_FullFlow(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "fay@example.com", "old-password", "Fay")
	require.NoError(t, err)

	token, err := svc.RequestPasswordReset(ctx, "fay@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.NoError(t, svc.ResetPassword(ctx, token, "new-password"))

	_, _, err = svc.Login(ctx, "fay@example.com", "old-password")
	assert.ErrorIs(t, err, ErrInvalidCredentials, "the old password must stop working")

	_, _, err = svc.Login(ctx, "fay@example.com", "new-password")
	assert.NoError(t, err, "the new password must work")
}

func TestResetPassword_TokenIsSingleUse(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "gus@example.com", "old-password", "Gus")
	require.NoError(t, err)
	token, err := svc.RequestPasswordReset(ctx, "gus@example.com")
	require.NoError(t, err)

	require.NoError(t, svc.ResetPassword(ctx, token, "new-password"))

	err = svc.ResetPassword(ctx, token, "another-password")
	assert.ErrorIs(t, err, ErrInvalidToken, "a used reset token must not be replayable")
}

func TestResetPassword_RejectsUnknownToken(t *testing.T) {
	svc, _ := newTestAuthService(t)

	err := svc.ResetPassword(context.Background(), "not-a-real-token", "new-password")

	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestRequestEmailVerification_UnknownEmailReturnsNoTokenAndNoError(t *testing.T) {
	svc, _ := newTestAuthService(t)

	token, err := svc.RequestEmailVerification(context.Background(), "nobody@example.com")

	assert.NoError(t, err)
	assert.Empty(t, token)
}

func TestVerifyEmail_FullFlow(t *testing.T) {
	ctx := context.Background()
	svc, users := newTestAuthService(t)
	user, _, err := svc.Signup(ctx, "hana@example.com", "password1", "Hana")
	require.NoError(t, err)
	assert.Nil(t, user.EmailVerifiedAt, "a freshly signed-up user must start unverified")

	token, err := svc.RequestEmailVerification(ctx, "hana@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.NoError(t, svc.VerifyEmail(ctx, token))

	stored, err := users.FindByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.EmailVerifiedAt, "the user must now be marked verified")
}

func TestVerifyEmail_TokenIsSingleUse(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "ivan@example.com", "password1", "Ivan")
	require.NoError(t, err)
	token, err := svc.RequestEmailVerification(ctx, "ivan@example.com")
	require.NoError(t, err)

	require.NoError(t, svc.VerifyEmail(ctx, token))

	err = svc.VerifyEmail(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidToken, "a used verification token must not be replayable")
}

func TestVerifyEmail_RejectsUnknownToken(t *testing.T) {
	svc, _ := newTestAuthService(t)

	err := svc.VerifyEmail(context.Background(), "not-a-real-token")

	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestRequestEmailVerification_AlreadyVerifiedReturnsNoToken(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestAuthService(t)
	_, _, err := svc.Signup(ctx, "jill@example.com", "password1", "Jill")
	require.NoError(t, err)
	firstToken, err := svc.RequestEmailVerification(ctx, "jill@example.com")
	require.NoError(t, err)
	require.NoError(t, svc.VerifyEmail(ctx, firstToken))

	token, err := svc.RequestEmailVerification(ctx, "jill@example.com")

	assert.NoError(t, err)
	assert.Empty(t, token, "an already-verified user should not get a new token")
}
