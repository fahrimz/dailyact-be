package handlers

import (
	"dailyact/models"
	"dailyact/types"
	"dailyact/utils"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type MobileAuthHandler struct {
	db     *gorm.DB
	config *oauth2.Config
}

func NewMobileAuthHandler(db *gorm.DB) *MobileAuthHandler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	return &MobileAuthHandler{
		db:     db,
		config: config,
	}
}

type GoogleTokenRequest struct {
	IdToken string `json:"id_token" binding:"required"`
}

// VerifyGoogleToken handles token verification from mobile clients
func (h *MobileAuthHandler) VerifyGoogleToken(c *gin.Context) {
	var req GoogleTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Verify token with Google
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + req.IdToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"TOKEN_VERIFICATION_FAILED",
			"Failed to verify Google token",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"TOKEN_READ_ERROR",
			"Failed to read token verification response",
			err.Error(),
		))
		return
	}

	var tokenInfo struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		EmailVerified string `json:"email_verified"`
		Sub           string `json:"sub"` // Google ID
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"TOKEN_PARSE_ERROR",
			"Failed to parse token info",
			err.Error(),
		))
		return
	}

	// Check if email is verified
	if tokenInfo.EmailVerified != "true" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"EMAIL_NOT_VERIFIED",
			"Email not verified with Google",
			"User's email must be verified",
		))
		return
	}

	// Find or create user
	var user models.User
	if err := h.db.Where("google_id = ?", tokenInfo.Sub).First(&user).Error; err != nil {
		// Create new user
		user = models.User{
			Email:    tokenInfo.Email,
			Name:     tokenInfo.Name,
			Picture:  tokenInfo.Picture,
			GoogleID: tokenInfo.Sub,
			Role:     "user",
		}

		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
				"USER_CREATE_ERROR",
				"Failed to create user",
				err.Error(),
			))
			return
		}
	} else {
		// Update existing user
		user.LastLoginAt = time.Now()
		if err := h.db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
				"USER_UPDATE_ERROR",
				"Failed to update user",
				err.Error(),
			))
			return
		}
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"TOKEN_GENERATION_ERROR",
			"Failed to generate JWT token",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Login successful",
		gin.H{"token": token, "user": user},
		nil,
	))
}
