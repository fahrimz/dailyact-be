package handlers

import (
	"dailyact/models"
	"dailyact/types"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	oauth  *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	oauth := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		db:    db,
		oauth: oauth,
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.oauth.AuthCodeURL("state")
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Login URL generated",
		gin.H{"url": url},
		nil,
	))
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := h.oauth.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"AUTH_ERROR",
			"Failed to exchange code",
			err.Error(),
		))
		return
	}

	// Get user info from Google
	client := h.oauth.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"AUTH_ERROR",
			"Failed to get user info",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"AUTH_ERROR",
			"Failed to decode user info",
			err.Error(),
		))
		return
	}

	// Find or create user
	var user models.User
	result := h.db.Where("google_id = ?", googleUser.ID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:    googleUser.Email,
			Name:     googleUser.Name,
			Picture:  googleUser.Picture,
			GoogleID: googleUser.ID,
			Role:     models.RoleUser, // Default role
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
				"DB_ERROR",
				"Failed to create user",
				err.Error(),
			))
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to fetch user",
			result.Error.Error(),
		))
		return
	}

	// Update last login
	user.LastLoginAt = time.Now()
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to update last login",
			err.Error(),
		))
		return
	}

	// Generate JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"AUTH_ERROR",
			"Failed to generate token",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Login successful",
		gin.H{
			"token": tokenString,
			"user":  user,
		},
		nil,
	))
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"UNAUTHORIZED",
			"User not found in context",
			"Please login again",
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"User retrieved successfully",
		user,
		nil,
	))
}
