FROM alpine:latest

ARG TARGETOS
ARG TARGETARCH
ARG BASE_URL=https://github.com/accentdesign/grpc/releases/download
ARG RELEASE=v0.0.1

ENV RELEASE_URL ${BASE_URL}/${RELEASE}/auth-server-${RELEASE}-${TARGETOS}-${TARGETARCH}.tar.gz

WORKDIR /usr/local/bin

RUN apk add curl

RUN curl -L -o release.tar.gz ${RELEASE_URL} && \
    tar -xzvf release.tar.gz && \
    rm release.tar.gz

EXPOSE 50051

ENTRYPOINT ["auth-server"]