package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortking-api/internal/mailer"
	"shortking-api/internal/service"
)

type AuthHandler struct {
	auth   *service.AuthService
	mailer *mailer.Mailer
	// webBaseURL points at the frontend (not this API), used to build the
	// link in password-reset and verification emails.
	webBaseURL string
}

func NewAuthHandler(auth *service.AuthService, m *mailer.Mailer, webBaseURL string) *AuthHandler {
	return &AuthHandler{auth: auth, mailer: m, webBaseURL: webBaseURL}
}

type signupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type authResponse struct {
	User struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.auth.Signup(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.sendVerificationEmail(c, user.Email)

	var resp authResponse
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.DisplayName = user.DisplayName
	resp.AccessToken = tokens.AccessToken
	resp.RefreshToken = tokens.RefreshToken
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	var resp authResponse
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.DisplayName = user.DisplayName
	resp.AccessToken = tokens.AccessToken
	resp.RefreshToken = tokens.RefreshToken
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Me returns the authenticated user's profile, used by the dashboard topbar.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	user, err := h.auth.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID.String(),
		"email":         user.Email,
		"displayName":   user.DisplayName,
		"emailVerified": user.EmailVerifiedAt != nil,
	})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// ForgotPassword always responds 200 with the same message, whether or not
// the email is registered, so this endpoint can't be used to enumerate
// accounts.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.auth.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if token != "" {
		link := fmt.Sprintf("%s/reset-password?token=%s", h.webBaseURL, token)
		body := fmt.Sprintf("Someone requested a password reset for your ShortKing account.\n\n"+
			"Reset your password: %s\n\nThis link expires in 1 hour. If you didn't request this, you can ignore this email.", link)
		if err := h.mailer.Send(req.Email, "Reset your ShortKing password", body); err != nil {
			// Falls back to logging the link so local dev works without SMTP
			// credentials configured; any other failure is logged, not
			// returned, so this endpoint can't be used to probe delivery.
			log.Printf("password reset email: failed to send to %s: %v (link: %s)", req.Email, err, link)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "if that email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auth.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		writeAuthError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// ResendVerification always responds 200 with the same message, whether or
// not the email is registered or already verified — same enumeration
// concern as ForgotPassword.
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.sendVerificationEmail(c, req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "if that email exists and isn't verified yet, a verification link has been sent",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auth.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		writeAuthError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// sendVerificationEmail issues a token and emails the verification link.
// Errors are logged, not returned: a failure here must never fail the
// caller (signup or resend).
func (h *AuthHandler) sendVerificationEmail(c *gin.Context, email string) {
	token, err := h.auth.RequestEmailVerification(c.Request.Context(), email)
	if err != nil {
		log.Printf("email verification: failed to issue token for %s: %v", email, err)
		return
	}
	if token == "" {
		return
	}

	link := fmt.Sprintf("%s/verify-email?token=%s", h.webBaseURL, token)
	body := fmt.Sprintf("Welcome to ShortKing! Verify your email to finish setting up your account.\n\n"+
		"Verify your email: %s\n\nThis link expires in 24 hours.", link)
	if err := h.mailer.Send(email, "Verify your ShortKing email", body); err != nil {
		// Falls back to logging the link so local dev works without SMTP
		// credentials configured.
		log.Printf("email verification: failed to send to %s: %v (link: %s)", email, err, link)
	}
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials), errors.Is(err, service.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
