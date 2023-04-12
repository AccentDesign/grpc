# Email Service

An email service that implements:

* SendEmail
  * Plain & HTML
  * Attachments

A connection is attempted during the init process of the server to test valid credentials.

## Arguments

Command line arguments the service accepts:

| Argument                                | Description                                |
|-----------------------------------------|--------------------------------------------|
| `-h`, `--help`                          | Show help message and exit                 |
| `-reflection`, `--reflection`           | Used to allow gRPC Web UI tools to connect |
| `-port`, `--port`                       | Port to bind to                            |

## Environment

A list of the environment variables:

| Variable        | Description                                        |
|-----------------|----------------------------------------------------|
| `SMTP_HOST`     | SMTP server host (e.g. smtp.sendgrid.net)          |
| `SMTP_PORT`     | SMTP server port (e.g. 587)                        |
| `SMTP_USERNAME` | SMTP server username (e.g. apikey)                 |
| `SMTP_PASSWORD` | SMTP server password (e.g. my-sendgrid-key)        |
| `SMTP_STARTTLS` | SMTP server start TLS (e.g. t,1,true or f,0,false) |

## Building in Go

Build the binary using GO locally, this will create an executable file.

    cd services/email && go build -o bin/server cmd/server/main.go

## Regenerate gRPC Code

    cd services/email/pkg/api/email

    protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    email.proto