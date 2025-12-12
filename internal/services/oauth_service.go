package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuth provider types
const (
	OAuthProviderGoogle = "google"
	OAuthProviderGithub = "github"
)

// OAuthUserInfo represents user info from OAuth providers
type OAuthUserInfo struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Provider  string
}

// GoogleUserInfo from Google API
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GitHubUserInfo from GitHub API
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmail from GitHub API
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// GetGoogleOAuthConfig returns Google OAuth configuration
func GetGoogleOAuthConfig() *oauth2.Config {
	cfg := config.GetConfig()
	return &oauth2.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GetGitHubOAuthConfig returns GitHub OAuth configuration
func GetGitHubOAuthConfig() *oauth2.Config {
	cfg := config.GetConfig()
	return &oauth2.Config{
		ClientID:     cfg.OAuth.GitHub.ClientID,
		ClientSecret: cfg.OAuth.GitHub.ClientSecret,
		RedirectURL:  cfg.OAuth.GitHub.RedirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}
}

// GetGoogleUserInfo fetches user info from Google
func GetGoogleUserInfo(code string) (*OAuthUserInfo, error) {
	oauthConfig := GetGoogleOAuthConfig()

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  OAuthProviderGoogle,
	}, nil
}

// GetGitHubUserInfo fetches user info from GitHub
func GetGitHubUserInfo(code string) (*OAuthUserInfo, error) {
	oauthConfig := GetGitHubOAuthConfig()

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := oauthConfig.Client(context.Background(), token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var githubUser GitHubUserInfo
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// Get primary email if not public
	email := githubUser.Email
	if email == "" {
		email, err = getGitHubPrimaryEmail(client)
		if err != nil {
			return nil, err
		}
	}

	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     email,
		Name:      name,
		AvatarURL: githubUser.AvatarURL,
		Provider:  OAuthProviderGithub,
	}, nil
}

func getGitHubPrimaryEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("failed to get emails: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("failed to parse emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", errors.New("no email found")
}

// FindOrCreateOAuthUser finds existing user or creates new one
func FindOrCreateOAuthUser(info *OAuthUserInfo) (*models.User, error) {
	db := database.GetDB()

	// Try to find by OAuth ID
	var user models.User
	err := db.Where("oauth_provider = ? AND oauth_id = ?", info.Provider, info.ID).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// Try to find by email
	err = db.Where("email = ?", info.Email).First(&user).Error
	if err == nil {
		// Link OAuth to existing account
		user.OAuthProvider = &info.Provider
		user.OAuthID = &info.ID
		if user.AvatarURL == nil || *user.AvatarURL == "" {
			user.AvatarURL = &info.AvatarURL
		}
		db.Save(&user)
		return &user, nil
	}

	// Create new user
	newUser := models.User{
		Email:         info.Email,
		Name:          info.Name,
		AvatarURL:     &info.AvatarURL,
		Role:          models.RoleUser,
		Status:        models.StatusActive,
		OAuthProvider: &info.Provider,
		OAuthID:       &info.ID,
	}

	if err := db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}

// AuthResult for login/register
type AuthResult struct {
	User      *models.User
	TokenPair *utils.TokenPair
}

// AuthenticateOAuthUser authenticates or creates OAuth user and returns tokens
func AuthenticateOAuthUser(info *OAuthUserInfo) (*AuthResult, error) {
	user, err := FindOrCreateOAuthUser(info)
	if err != nil {
		return nil, err
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}
