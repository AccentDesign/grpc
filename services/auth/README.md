# Auth Service

An authentication service that implements:

* BearerToken
* RevokeBearerToken
* Register
* ResetPassword
* ResetPasswordToken
* User
* UpdateUser
* VerifyUser
* VerifyUserToken
* UserTypes
* Scopes

Example user:

    {
      "id": "3f4b2b87-d7b1-4b9f-b207-ae00b112382f",
      "email": "someone@example.com",
      "first_name": "Some",
      "last_name": "One",
      "user_type": {
        "name": "user",
        "scopes": [
          "write",
          "read"
        ]
      },
      "is_active": true,
      "is_verified": true
    }

## Arguments

Command line arguments the service accepts:

| Argument                                | Description                                                           |
|-----------------------------------------|-----------------------------------------------------------------------|
| `-h`, `--help`                          | Show help message and exit                                            |
| `-reflection`, `--reflection`           | Used to allow gRPC Web UI tools to connect                            |
| `-port`, `--port`                       | Port to bind to                                                       |
| `-bearer-duration`, `--bearer-duration` | Duration of bearer tokens (e.g. 8h)                                   |
| `-reset-duration`, `--reset-duration`   | Duration of reset tokens (e.g. 1h)                                    |
| `-verify-duration`, `--verify-duration` | Duration of verify tokens (e.g. 1h)                                   |
| `-migrations`, `--migrations`           | Migrations, "on", "dry-run" or "off", dry run will exit, default "on" |

## Environment

A list of the environment variables:

| Variable | Description                                                                 |
|----------|-----------------------------------------------------------------------------|
| `DB_DNS` | Postgres dns url e.g. `postgresql://user:pass@host:5432/db?sslmode=disable` |

## Building in Go

Build the binary using GO locally, this will create an executable file.

    cd services/auth && go build -o bin/server cmd/server/main.go

## Regenerate gRPC Code

    cd services/auth/pkg/api/auth

    protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    auth.proto