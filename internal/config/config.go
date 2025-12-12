package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Midtrans MidtransConfig
	Upload   UploadConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  int
	Debug bool
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret             string
	ExpiryHours        int
	RefreshExpiryHours int
}

type OAuthConfig struct {
	Google OAuthProviderConfig
	GitHub OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type MidtransConfig struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
}

type UploadConfig struct {
	Dir     string
	MaxSize int64
}

type CORSConfig struct {
	FrontendURL string
}

var AppConfig_ *Config

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Enable environment variable override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	bindEnvVariables()

	config := &Config{
		App: AppConfig{
			Name:  viper.GetString("app.name"),
			Env:   viper.GetString("app.env"),
			Port:  viper.GetInt("app.port"),
			Debug: viper.GetBool("app.debug"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			Name:     viper.GetString("database.name"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("jwt.secret"),
			ExpiryHours:        viper.GetInt("jwt.expiry_hours"),
			RefreshExpiryHours: viper.GetInt("jwt.refresh_expiry_hours"),
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     viper.GetString("oauth.google.client_id"),
				ClientSecret: viper.GetString("oauth.google.client_secret"),
				RedirectURL:  viper.GetString("oauth.google.redirect_url"),
			},
			GitHub: OAuthProviderConfig{
				ClientID:     viper.GetString("oauth.github.client_id"),
				ClientSecret: viper.GetString("oauth.github.client_secret"),
				RedirectURL:  viper.GetString("oauth.github.redirect_url"),
			},
		},
		Midtrans: MidtransConfig{
			ServerKey:    viper.GetString("midtrans.server_key"),
			ClientKey:    viper.GetString("midtrans.client_key"),
			IsProduction: viper.GetBool("midtrans.is_production"),
		},
		Upload: UploadConfig{
			Dir:     viper.GetString("upload.dir"),
			MaxSize: viper.GetInt64("upload.max_size"),
		},
		CORS: CORSConfig{
			FrontendURL: viper.GetString("cors.frontend_url"),
		},
	}

	// Set defaults
	if config.App.Port == 0 {
		config.App.Port = 8000
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.JWT.ExpiryHours == 0 {
		config.JWT.ExpiryHours = 24
	}
	if config.JWT.RefreshExpiryHours == 0 {
		config.JWT.RefreshExpiryHours = 168
	}
	if config.Upload.Dir == "" {
		config.Upload.Dir = "./uploads"
	}
	if config.Upload.MaxSize == 0 {
		config.Upload.MaxSize = 10 * 1024 * 1024 // 10MB
	}

	AppConfig_ = config
	return config, nil
}

func bindEnvVariables() {
	// App
	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("app.debug", "APP_DEBUG")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiry_hours", "JWT_EXPIRY_HOURS")
	viper.BindEnv("jwt.refresh_expiry_hours", "JWT_REFRESH_EXPIRY_HOURS")

	// OAuth - Google
	viper.BindEnv("oauth.google.client_id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.google.client_secret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.google.redirect_url", "GOOGLE_REDIRECT_URL")

	// OAuth - GitHub
	viper.BindEnv("oauth.github.client_id", "GITHUB_CLIENT_ID")
	viper.BindEnv("oauth.github.client_secret", "GITHUB_CLIENT_SECRET")
	viper.BindEnv("oauth.github.redirect_url", "GITHUB_REDIRECT_URL")

	// Midtrans
	viper.BindEnv("midtrans.server_key", "MIDTRANS_SERVER_KEY")
	viper.BindEnv("midtrans.client_key", "MIDTRANS_CLIENT_KEY")
	viper.BindEnv("midtrans.is_production", "MIDTRANS_IS_PRODUCTION")

	// Upload
	viper.BindEnv("upload.dir", "UPLOAD_DIR")
	viper.BindEnv("upload.max_size", "MAX_UPLOAD_SIZE")

	// CORS
	viper.BindEnv("cors.frontend_url", "FRONTEND_URL")
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func GetConfig() *Config {
	if AppConfig_ == nil {
		cfg, err := Load()
		if err != nil {
			panic(err)
		}
		return cfg
	}
	return AppConfig_
}

// EnsureUploadDir creates the upload directory if it doesn't exist
func EnsureUploadDir() error {
	cfg := GetConfig()
	return os.MkdirAll(cfg.Upload.Dir, 0755)
}
