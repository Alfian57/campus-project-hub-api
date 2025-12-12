package handlers

import (
	"net/http"

	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/campus-project-hub/api/internal/services"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body services.RegisterInput true "Registration data"
// @Success      201 {object} map[string]interface{} "Successfully registered"
// @Failure      400 {object} map[string]interface{} "Invalid input or email already exists"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := utils.Validate(&input); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errors":  errors,
		})
		return
	}

	result, err := services.Register(&input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, gin.H{
		"user":  result.User.ToResponse(),
		"token": result.TokenPair,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body services.LoginInput true "Login credentials"
// @Success      200 {object} map[string]interface{} "Successfully logged in"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Invalid credentials"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := utils.Validate(&input); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errors":  errors,
		})
		return
	}

	result, err := services.Login(&input)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"user":  result.User.ToResponse(),
		"token": result.TokenPair,
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Generate new token pair using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{refreshToken=string} true "Refresh token"
// @Success      200 {object} map[string]interface{} "New token pair"
// @Failure      400 {object} map[string]interface{} "Refresh token required"
// @Failure      401 {object} map[string]interface{} "Invalid refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Refresh token diperlukan")
		return
	}

	tokenPair, err := services.RefreshTokens(input.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"token": tokenPair,
	})
}

// GetMe godoc
// @Summary      Get current user
// @Description  Get authenticated user profile with gamification stats
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "User profile with gamification"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.Unauthorized(c, "Tidak terautentikasi")
		return
	}

	stats := services.GetUserGamificationStats(user.TotalExp)

	utils.Success(c, gin.H{
		"user":         user.ToResponse(),
		"gamification": stats,
	})
}

// GoogleAuth godoc
// @Summary      Google OAuth
// @Description  Initiate Google OAuth flow
// @Tags         auth
// @Produce      json
// @Success      307 {string} string "Redirect to Google"
// @Router       /auth/google [get]
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	oauthConfig := services.GetGoogleOAuthConfig()
	url := oauthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary      Google OAuth callback
// @Description  Handle Google OAuth callback
// @Tags         auth
// @Produce      json
// @Param        code query string true "Authorization code"
// @Success      307 {string} string "Redirect to frontend with token"
// @Failure      400 {object} map[string]interface{} "Missing code"
// @Failure      500 {object} map[string]interface{} "OAuth error"
// @Router       /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.BadRequest(c, "Kode otorisasi tidak ditemukan")
		return
	}

	userInfo, err := services.GetGoogleUserInfo(code)
	if err != nil {
		utils.InternalServerError(c, "Gagal mendapatkan info user dari Google")
		return
	}

	result, err := services.AuthenticateOAuthUser(userInfo)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	// Redirect to frontend with token
	frontendURL := "http://localhost:3000/auth/callback"
	c.Redirect(http.StatusTemporaryRedirect,
		frontendURL+"?token="+result.TokenPair.AccessToken+"&refresh="+result.TokenPair.RefreshToken)
}

// GitHubAuth godoc
// @Summary      GitHub OAuth
// @Description  Initiate GitHub OAuth flow
// @Tags         auth
// @Produce      json
// @Success      307 {string} string "Redirect to GitHub"
// @Router       /auth/github [get]
func (h *AuthHandler) GitHubAuth(c *gin.Context) {
	oauthConfig := services.GetGitHubOAuthConfig()
	url := oauthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback godoc
// @Summary      GitHub OAuth callback
// @Description  Handle GitHub OAuth callback
// @Tags         auth
// @Produce      json
// @Param        code query string true "Authorization code"
// @Success      307 {string} string "Redirect to frontend with token"
// @Failure      400 {object} map[string]interface{} "Missing code"
// @Failure      500 {object} map[string]interface{} "OAuth error"
// @Router       /auth/github/callback [get]
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.BadRequest(c, "Kode otorisasi tidak ditemukan")
		return
	}

	userInfo, err := services.GetGitHubUserInfo(code)
	if err != nil {
		utils.InternalServerError(c, "Gagal mendapatkan info user dari GitHub")
		return
	}

	result, err := services.AuthenticateOAuthUser(userInfo)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	// Redirect to frontend with token
	frontendURL := "http://localhost:3000/auth/callback"
	c.Redirect(http.StatusTemporaryRedirect,
		frontendURL+"?token="+result.TokenPair.AccessToken+"&refresh="+result.TokenPair.RefreshToken)
}

// Logout godoc
// @Summary      Logout user
// @Description  Logout current user (client should clear token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Logout successful"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SuccessWithMessage(c, "Logout berhasil", nil)
}
