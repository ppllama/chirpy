# Chirpy

Boot.dev HTTP Servers Course Project – Chirpy: A social network similar to Twitter  

Chirpy is a back-end server that can receive, store, and serve text. It exposes RESTful API endpoints for managing users, authentication, and posts (“chirps”). Postgresql is used as the database

---

## API Endpoints

### Health & Metrics

| Method | Endpoint            | Description                          |
|--------|-------------------|--------------------------------------|
| GET    | `/api/healthz`     | Check if the server is running (readiness) |
| GET    | `/admin/metrics`   | Get server metrics                    |
| POST   | `/admin/reset`     | Reset the server data                 |

### Chirps

| Method | Endpoint                   | Description                          |
|--------|----------------------------|--------------------------------------|
| POST   | `/api/chirps`              | Create a new chirp                   |
| GET    | `/api/chirps`              | List all chirps                      |
| GET    | `/api/chirps/{chirp_id}`   | Get a single chirp by ID             |
| DELETE | `/api/chirps/{chirp_id}`   | Delete a chirp by ID                 |

### Users & Authentication

| Method | Endpoint                   | Description                          |
|--------|----------------------------|--------------------------------------|
| POST   | `/api/users`               | Create a new user                     |
| PUT    | `/api/users`               | Update user details                   |
| POST   | `/api/login`               | Log in and receive access token       |
| POST   | `/api/refresh`             | Refresh access token                  |
| POST   | `/api/revoke`              | Revoke access token                   |
| POST   | `/api/polka/webhooks`      | Upgrade user (Fictional payments processor Polka integration)     |

### Static Files

| Method | Endpoint      | Description                      |
|--------|---------------|----------------------------------|
| GET    | `/app/*`      | Serve static files from the app directory |

---

## Installation

```bash
# Clone the repository
git clone <repository_url>
cd chirpy

# Build the project
go build -o chirpy

# Run the server
./chirpy
```

---

## Database

Chirpy uses PostgreSQL (v18). Database setup is required before running the server.

```env
Environment Variables (.env)
DB_URL=postgres://chirpy_user:secret@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev          # Allows access to /admin/reset endpoint
JWT_SECRET=your_jwt_secret
POLKA_KEY=your_polka_key
```

### Setup

Create the database:
```bash
createdb chirpy
```

Run migrations using Goose:
```bash
# Install goose if not installed

go install github.com/pressly/goose/v3/cmd/goose@latest

# Apply migrations
goose -dir ./sql/schema postgres "$DB_URL" up
```

Start the server:
```bash
./chirpy
```