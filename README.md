# gRPC Services

[![Go](https://github.com/AccentDesign/grpc/actions/workflows/go-test.yml/badge.svg)](https://github.com/AccentDesign/grpc/actions/workflows/go-test.yml)

Services hosted within Accent Design using Google Remote Procedure Call.

## Services

A list of all the service built in this library:

| Service                 | Description                       |
|-------------------------|-----------------------------------|
| [Auth](./services/auth) | Email and password authentication |

## Tools

Some useful external tools:

#### gRPC Web UI

Connect to a running gRPC service that has reflection enabled.

    docker run \
    --publish 8080:8080 \
    fullstorydev/grpcui:latest \
    -bind 0.0.0.0 -port 8080 -plaintext <host-ip>:50051