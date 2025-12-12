# Campus Project Hub API

Backend API untuk platform Campus Project Hub menggunakan Golang.

## Tech Stack

- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT + OAuth (Google, GitHub)
- **Payment**: Midtrans
- **Configuration**: Viper

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Make (optional)

### Installation

1. Clone repository
2. Copy environment file:
   ```bash
   cp .env.example .env
   ```
3. Update `.env` with your configuration
4. Install dependencies:
   ```bash
   go mod tidy
   ```
5. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

### Docker

```bash
docker-compose up -d
```

## API Documentation

Base URL: `http://localhost:8000/api/v1`

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login with email/password |
| POST | `/auth/refresh` | Refresh JWT token |
| GET | `/auth/google` | Google OAuth |
| GET | `/auth/github` | GitHub OAuth |
| GET | `/auth/me` | Get current user |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/leaderboard` | Get EXP leaderboard |
| GET | `/users/:id` | Get user by ID |
| PUT | `/users/:id` | Update user profile |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/:id` | Get project |
| PUT | `/projects/:id` | Update project |
| DELETE | `/projects/:id` | Delete project |
| POST | `/projects/:id/like` | Like/unlike project |

### Articles

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/articles` | List articles |
| POST | `/articles` | Create article |
| GET | `/articles/:id` | Get article |
| PUT | `/articles/:id` | Update article |
| DELETE | `/articles/:id` | Delete article |

### Transactions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/transactions` | Create payment |
| GET | `/transactions` | List transactions |
| POST | `/transactions/callback` | Midtrans webhook |

## License

MIT
