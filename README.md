# Chirpy

REST API developed using Go and PostgreSQL, allows users to register and perform CRUD operations on chirps.

## Installation

To get started, you need to have [Go](https://go.dev/doc/install) and [PostgreSQL](https://www.postgresql.org/download/) installed.

1. Clone the repository

```bash
git clone https://github.com/syeero7/boot-chirpy
cd boot-chirpy
```

2. Create a file named .env in the root directory and configure environment variables

```.env
DB_URL=database_url?sslmode=disable
PLATFORM=dev
JWT_SECRET=jwt_secret
POLKA_KEY=polka_api_key
```

3. Run migrations and generate queries

```bash
# install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# run migrations
cd sql/schema
goose postgres <database_url> up
cd ../..

# install sqlc 
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# generate queries
sqlc generate 
```

4. Install chirpy

```bash
go install
```

## API Reference

The API will be available at `http://localhost:8080`

### Authentication & User Management

| Endpoint | Description |
| :-------------| :-----------------------|
| `POST /api/users` | Create a new user account|
| `PUT /api/users` | Update the authenticated user's information |
| `POST /api/login` |  Authenticate and receive access token and refresh token |
| `POST /api/refresh` | Generate a new access token using a valid refresh token |
| `POST /api/revoke` | Revoke user's refresh tokens |

### Chirps CRUD

| Endpoint | Description |
| :-------------| :-----------------------|
| `POST /api/chirps` |  Create a new chirp |
| `GET /api/chirps` | Retrieve a list of all chirps, supporting optional query parameters for `author_id` and `sort` |
| `GET /api/chirps/{chirp_id}` | Retrieve a specific chirp by id |
| `DELETE /api/chirps/{chirp_id}` | Delete a chirp by id |
