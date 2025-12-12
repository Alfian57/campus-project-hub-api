package services

import (
	"errors"
	"fmt"

	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/google/uuid"
)

// RegisterInput for user registration
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginInput for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register creates a new user account
func Register(input *RegisterInput) (*AuthResult, error) {
	db := database.GetDB()

	// Check if email exists
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses password: %w", err)
	}

	// Create user
	user := models.User{
		Email:        input.Email,
		PasswordHash: &hashedPassword,
		Name:         input.Name,
		Role:         models.RoleUser,
		Status:       models.StatusActive,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("gagal membuat akun: %w", err)
	}

	// Generate tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token: %w", err)
	}

	return &AuthResult{
		User:      &user,
		TokenPair: tokenPair,
	}, nil
}

// Login authenticates a user
func Login(input *LoginInput) (*AuthResult, error) {
	db := database.GetDB()

	// Find user by email
	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return nil, errors.New("email atau password salah")
	}

	// Check if user is blocked
	if user.Status == models.StatusBlocked {
		return nil, errors.New("akun Anda telah diblokir")
	}

	// Check if user has password (might be OAuth only)
	if user.PasswordHash == nil {
		return nil, errors.New("silakan login menggunakan akun sosial")
	}

	// Verify password
	if !utils.CheckPassword(*user.PasswordHash, input.Password) {
		return nil, errors.New("email atau password salah")
	}

	// Generate tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token: %w", err)
	}

	return &AuthResult{
		User:      &user,
		TokenPair: tokenPair,
	}, nil
}

// RefreshTokens generates new token pair from refresh token
func RefreshTokens(refreshToken string) (*utils.TokenPair, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	db := database.GetDB()
	var user models.User
	if err := db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if user.Status == models.StatusBlocked {
		return nil, errors.New("akun Anda telah diblokir")
	}

	return utils.GenerateTokenPair(user.ID, user.Email, string(user.Role))
}

// GetUserByID retrieves a user by ID
func GetUserByID(id uuid.UUID) (*models.User, error) {
	db := database.GetDB()
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	return &user, nil
}

// UpdateUserInput for profile updates
type UpdateUserInput struct {
	Name       *string `json:"name" validate:"omitempty,min=2,max=100"`
	University *string `json:"university" validate:"omitempty,max=255"`
	Major      *string `json:"major" validate:"omitempty,max=255"`
	Bio        *string `json:"bio" validate:"omitempty,max=500"`
	Phone      *string `json:"phone" validate:"omitempty,max=20"`
	AvatarURL  *string `json:"avatarUrl" validate:"omitempty,url"`
}

// UpdateUser updates user profile
func UpdateUser(userID uuid.UUID, input *UpdateUserInput) (*models.User, error) {
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.University != nil {
		user.University = input.University
	}
	if input.Major != nil {
		user.Major = input.Major
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.Phone != nil {
		user.Phone = input.Phone
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("gagal memperbarui profil: %w", err)
	}

	return &user, nil
}

// AddUserExp adds EXP to user
func AddUserExp(userID uuid.UUID, exp int) error {
	db := database.GetDB()
	return db.Model(&models.User{}).Where("id = ?", userID).
		Update("total_exp", database.GetDB().Raw("total_exp + ?", exp)).Error
}
