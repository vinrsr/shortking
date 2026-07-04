package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortking-api/internal/handler"
	"shortking-api/internal/mailer"
	"shortking-api/internal/service"
	"shortking-api/internal/testsupport"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// testAPI wires the real router + handlers + services on top of in-memory
// fakes and a miniredis-backed cache, so these tests exercise routing,
// middleware (auth, CORS, rate limiting) and handler wiring exactly as
// cmd/server/main.go does, without a live Postgres.
type testAPI struct {
	handler   http.Handler
	linkRepo  *testsupport.FakeLinkRepository
	userRepo  *testsupport.FakeUserRepository
	clickRepo *testsupport.FakeClickRepository
	// auth is the same service instance wired into the router. Tests use it
	// to fetch a password-reset token directly, standing in for the
	// out-of-band email a real user would receive.
	auth *service.AuthService
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()

	userRepo := testsupport.NewFakeUserRepository()
	linkRepo := testsupport.NewFakeLinkRepository()
	clickRepo := testsupport.NewFakeClickRepository()
	statsRepo := testsupport.NewFakeStatsRepository()
	c := testsupport.NewTestCache(t)

	authService := service.NewAuthService(userRepo, c, "access-secret", "refresh-secret", 15*time.Minute, 30*24*time.Hour)
	linkService := service.NewLinkService(linkRepo, c, "http://sk.io")
	clickRecorder := service.NewClickRecorder(clickRepo, linkRepo, "pepper")
	statsService := service.NewStatsService(statsRepo)

	engine, err := New(Deps{
		Redis:          c.Client(),
		AllowedOrigins: []string{"http://localhost:3000"},
		Auth:           handler.NewAuthHandler(authService, mailer.New(mailer.Config{}), "http://localhost:3000"),
		Link:           handler.NewLinkHandler(linkService, clickRepo, statsService),
		Redirect:       handler.NewRedirectHandler(linkService, clickRecorder, "http://localhost:3000"),
		Stats:          handler.NewStatsHandler(linkService, authService, clickRepo, statsService),
		AuthSvc:        authService,
	})
	require.NoError(t, err)

	return &testAPI{handler: engine, linkRepo: linkRepo, userRepo: userRepo, clickRepo: clickRepo, auth: authService}
}

func (a *testAPI) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec
}

func (a *testAPI) signup(t *testing.T, email, password, displayName string) (userID, accessToken string) {
	t.Helper()
	rec := a.do(t, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": email, "password": password, "displayName": displayName,
	}, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var resp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		AccessToken string `json:"accessToken"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.User.ID, resp.AccessToken
}

// verifyEmail marks the given user's email verified, standing in for the
// user clicking the link in their verification email. Link creation
// requires a verified email, so tests exercising link CRUD call this right
// after signup.
func (a *testAPI) verifyEmail(t *testing.T, email string) {
	t.Helper()
	token, err := a.auth.RequestEmailVerification(t.Context(), email)
	require.NoError(t, err)
	require.NoError(t, a.auth.VerifyEmail(t.Context(), token))
}

func TestSignupLoginMe_FullFlow(t *testing.T) {
	api := newTestAPI(t)

	_, accessToken := api.signup(t, "ann@example.com", "hunter2pass", "Ann")
	require.NotEmpty(t, accessToken)

	loginRec := api.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "ann@example.com", "password": "hunter2pass",
	}, "")
	assert.Equal(t, http.StatusOK, loginRec.Code, loginRec.Body.String())

	meRec := api.do(t, http.MethodGet, "/api/v1/me", nil, accessToken)
	require.Equal(t, http.StatusOK, meRec.Code, meRec.Body.String())
	var me struct {
		DisplayName   string `json:"displayName"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"emailVerified"`
	}
	require.NoError(t, json.Unmarshal(meRec.Body.Bytes(), &me))
	assert.Equal(t, "Ann", me.DisplayName)
	assert.Equal(t, "ann@example.com", me.Email)
	assert.False(t, me.EmailVerified, "a freshly signed-up user must start unverified")
}

func TestSignup_DuplicateEmailReturnsConflict(t *testing.T) {
	api := newTestAPI(t)
	api.signup(t, "dup@example.com", "hunter2pass", "First")

	rec := api.do(t, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": "dup@example.com", "password": "otherpass1", "displayName": "Second",
	}, "")

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestLogin_WrongPasswordReturnsUnauthorized(t *testing.T) {
	api := newTestAPI(t)
	api.signup(t, "bob@example.com", "hunter2pass", "Bob")

	rec := api.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "bob@example.com", "password": "wrong-password",
	}, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLinks_RequireAuth(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodGet, "/api/v1/links", nil, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLinks_CreateListDeleteFullFlow(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "carl@example.com", "hunter2pass", "Carl")
	api.verifyEmail(t, "carl@example.com")

	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, accessToken)
	require.Equal(t, http.StatusCreated, createRec.Code, createRec.Body.String())
	var created struct {
		ID        string `json:"id"`
		ShortCode string `json:"shortCode"`
	}
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	assert.NotEmpty(t, created.ShortCode)

	listRec := api.do(t, http.MethodGet, "/api/v1/links", nil, accessToken)
	require.Equal(t, http.StatusOK, listRec.Code)
	var links []struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &links))
	require.Len(t, links, 1)
	assert.Equal(t, created.ID, links[0].ID)

	deleteRec := api.do(t, http.MethodDelete, "/api/v1/links/"+created.ID, nil, accessToken)
	assert.Equal(t, http.StatusNoContent, deleteRec.Code)

	listAfterRec := api.do(t, http.MethodGet, "/api/v1/links", nil, accessToken)
	require.Equal(t, http.StatusOK, listAfterRec.Code)
	var afterDelete []json.RawMessage
	require.NoError(t, json.Unmarshal(listAfterRec.Body.Bytes(), &afterDelete))
	assert.Empty(t, afterDelete)
}

func TestLinks_CreateBlockedUntilEmailVerified(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "gina@example.com", "hunter2pass", "Gina")

	blockedRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, accessToken)
	assert.Equal(t, http.StatusForbidden, blockedRec.Code)

	api.verifyEmail(t, "gina@example.com")

	allowedRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, accessToken)
	assert.Equal(t, http.StatusCreated, allowedRec.Code, allowedRec.Body.String())
}

func TestLinks_CannotDeleteAnotherUsersLink(t *testing.T) {
	api := newTestAPI(t)
	_, ownerToken := api.signup(t, "owner@example.com", "hunter2pass", "Owner")
	api.verifyEmail(t, "owner@example.com")
	_, otherToken := api.signup(t, "other@example.com", "hunter2pass", "Other")

	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, createRec.Code)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteRec := api.do(t, http.MethodDelete, "/api/v1/links/"+created.ID, nil, otherToken)
	assert.Equal(t, http.StatusForbidden, deleteRec.Code)
}

func TestRedirect_UnknownCodeRedirectsToNotFoundPage(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodGet, "/does-not-exist", nil, "")

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "http://localhost:3000/link-not-found", rec.Header().Get("Location"))
}

func TestRedirect_ValidCodeRedirectsToDestination(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "dana@example.com", "hunter2pass", "Dana")
	api.verifyEmail(t, "dana@example.com")
	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com/target",
		"customAlias": "my-alias",
	}, accessToken)
	require.Equal(t, http.StatusCreated, createRec.Code, createRec.Body.String())

	redirectRec := api.do(t, http.MethodGet, "/my-alias", nil, "")

	assert.Equal(t, http.StatusFound, redirectRec.Code)
	assert.Equal(t, "https://example.com/target", redirectRec.Header().Get("Location"))
}

func TestRedirect_ExpiredCodeRedirectsToExpiredPage(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "faye@example.com", "hunter2pass", "Faye")
	api.verifyEmail(t, "faye@example.com")

	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com/target",
		"customAlias": "already-expired",
		"expiresAt":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	}, accessToken)
	require.Equal(t, http.StatusCreated, createRec.Code, createRec.Body.String())

	redirectRec := api.do(t, http.MethodGet, "/already-expired", nil, "")

	assert.Equal(t, http.StatusFound, redirectRec.Code)
	assert.Equal(t, "http://localhost:3000/link-expired", redirectRec.Header().Get("Location"))
}

func TestPublicShorten_DoesNotRequireAuth(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodPost, "/api/v1/shorten", map[string]string{
		"destination": "https://example.com",
	}, "")

	assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
}

func TestPublicShorten_DailyLimitBlocksTheSixthAttempt(t *testing.T) {
	api := newTestAPI(t)

	for i := 0; i < 5; i++ {
		rec := api.do(t, http.MethodPost, "/api/v1/shorten", map[string]string{
			"destination": "https://example.com",
		}, "")
		require.Equal(t, http.StatusCreated, rec.Code, "attempt %d should be within the daily limit: %s", i+1, rec.Body.String())
	}

	rec := api.do(t, http.MethodPost, "/api/v1/shorten", map[string]string{
		"destination": "https://example.com",
	}, "")
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Contains(t, rec.Body.String(), "sign up")
}

func TestRateLimits_OnDifferentRoutesDoNotShareOneCounter(t *testing.T) {
	// Regression: every rate limiter used to share the same Redis key per
	// IP (no prefix), so hitting one IP-keyed limit silently ate into every
	// other IP-keyed limit's budget. Exhausting the login rate limit (10-M)
	// must not affect the separate anonymous-shorten limits.
	api := newTestAPI(t)

	for i := 0; i < 10; i++ {
		api.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"email": "nobody@example.com", "password": "wrong-password",
		}, "")
	}
	loginRec := api.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "nobody@example.com", "password": "wrong-password",
	}, "")
	require.Equal(t, http.StatusTooManyRequests, loginRec.Code, "the login limiter itself should now be exhausted")

	shortenRec := api.do(t, http.MethodPost, "/api/v1/shorten", map[string]string{
		"destination": "https://example.com",
	}, "")
	assert.Equal(t, http.StatusCreated, shortenRec.Code, "shorten must have its own independent budget")
}

func TestPublicStats_ReturnsAggregateCounts(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "eve@example.com", "hunter2pass", "Eve")
	api.verifyEmail(t, "eve@example.com")
	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, accessToken)
	require.Equal(t, http.StatusCreated, createRec.Code)

	rec := api.do(t, http.MethodGet, "/api/v1/stats", nil, "")

	require.Equal(t, http.StatusOK, rec.Code)
	var stats struct {
		TotalLinks  int64 `json:"totalLinks"`
		ActiveLinks int64 `json:"activeLinks"`
		TotalUsers  int64 `json:"totalUsers"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &stats))
	assert.EqualValues(t, 1, stats.TotalLinks)
	assert.EqualValues(t, 1, stats.ActiveLinks)
	assert.EqualValues(t, 1, stats.TotalUsers)
}

func TestLinks_UpdateFullFlow(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "hank@example.com", "hunter2pass", "Hank")
	api.verifyEmail(t, "hank@example.com")
	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://old.example.com",
	}, accessToken)
	require.Equal(t, http.StatusCreated, createRec.Code)
	var created struct {
		ID        string `json:"id"`
		ShortCode string `json:"shortCode"`
	}
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateRec := api.do(t, http.MethodPatch, "/api/v1/links/"+created.ID, map[string]any{
		"destination": "https://new.example.com",
		"isActive":    false,
	}, accessToken)
	require.Equal(t, http.StatusOK, updateRec.Code, updateRec.Body.String())
	var updated struct {
		Destination string `json:"destination"`
		ShortCode   string `json:"shortCode"`
		IsActive    bool   `json:"isActive"`
	}
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &updated))
	assert.Equal(t, "https://new.example.com", updated.Destination)
	assert.Equal(t, created.ShortCode, updated.ShortCode)
	assert.False(t, updated.IsActive)

	// Deactivated links must stop redirecting.
	redirectRec := api.do(t, http.MethodGet, "/"+created.ShortCode, nil, "")
	assert.Equal(t, http.StatusFound, redirectRec.Code)
	assert.Equal(t, "http://localhost:3000/link-not-found", redirectRec.Header().Get("Location"))
}

func TestLinks_UpdateForbidsAnotherUsersLink(t *testing.T) {
	api := newTestAPI(t)
	_, ownerToken := api.signup(t, "iris@example.com", "hunter2pass", "Iris")
	api.verifyEmail(t, "iris@example.com")
	_, otherToken := api.signup(t, "jack@example.com", "hunter2pass", "Jack")
	createRec := api.do(t, http.MethodPost, "/api/v1/links", map[string]any{
		"destination": "https://example.com",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, createRec.Code)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateRec := api.do(t, http.MethodPatch, "/api/v1/links/"+created.ID, map[string]any{
		"destination": "https://evil.example.com",
		"isActive":    true,
	}, otherToken)

	assert.Equal(t, http.StatusForbidden, updateRec.Code)
}

func TestPasswordReset_FullFlowThroughHTTP(t *testing.T) {
	api := newTestAPI(t)
	api.signup(t, "kim@example.com", "old-password", "Kim")

	forgotRec := api.do(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "kim@example.com",
	}, "")
	assert.Equal(t, http.StatusOK, forgotRec.Code)

	// The token would normally arrive by email; fetch it directly from the
	// same service instance the handler used, standing in for that email.
	token, err := api.auth.RequestPasswordReset(t.Context(), "kim@example.com")
	require.NoError(t, err)

	resetRec := api.do(t, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token": token, "newPassword": "new-password",
	}, "")
	assert.Equal(t, http.StatusNoContent, resetRec.Code)

	loginRec := api.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "kim@example.com", "password": "new-password",
	}, "")
	assert.Equal(t, http.StatusOK, loginRec.Code)
}

func TestPasswordReset_UnknownEmailStillReturns200(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "nobody@example.com",
	}, "")

	assert.Equal(t, http.StatusOK, rec.Code, "must not reveal whether the email is registered")
}

func TestPasswordReset_RejectsBogusToken(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token": "not-a-real-token", "newPassword": "new-password",
	}, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestEmailVerification_FullFlowThroughHTTP(t *testing.T) {
	api := newTestAPI(t)
	_, accessToken := api.signup(t, "leo@example.com", "hunter2pass", "Leo")

	// Signup itself already issued a verification token (and logged the
	// dev-stub email); resend-verification would issue another one, so
	// fetch the one signup made instead of double-issuing.
	token, err := api.auth.RequestEmailVerification(t.Context(), "leo@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token, "signup must not have already verified the email")

	verifyRec := api.do(t, http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": token,
	}, "")
	assert.Equal(t, http.StatusNoContent, verifyRec.Code)

	meRec := api.do(t, http.MethodGet, "/api/v1/me", nil, accessToken)
	require.Equal(t, http.StatusOK, meRec.Code)
	var me struct {
		EmailVerified bool `json:"emailVerified"`
	}
	require.NoError(t, json.Unmarshal(meRec.Body.Bytes(), &me))
	assert.True(t, me.EmailVerified)
}

func TestEmailVerification_ResendEndpointIssuesAWorkingToken(t *testing.T) {
	api := newTestAPI(t)
	api.signup(t, "mia@example.com", "hunter2pass", "Mia")

	resendRec := api.do(t, http.MethodPost, "/api/v1/auth/resend-verification", map[string]string{
		"email": "mia@example.com",
	}, "")
	assert.Equal(t, http.StatusOK, resendRec.Code)

	token, err := api.auth.RequestEmailVerification(t.Context(), "mia@example.com")
	require.NoError(t, err)

	verifyRec := api.do(t, http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": token,
	}, "")
	assert.Equal(t, http.StatusNoContent, verifyRec.Code)
}

func TestEmailVerification_ResendForUnknownEmailStillReturns200(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodPost, "/api/v1/auth/resend-verification", map[string]string{
		"email": "nobody@example.com",
	}, "")

	assert.Equal(t, http.StatusOK, rec.Code, "must not reveal whether the email is registered")
}

func TestEmailVerification_RejectsBogusToken(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(t, http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": "not-a-real-token",
	}, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
