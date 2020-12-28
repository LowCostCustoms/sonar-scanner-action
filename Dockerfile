# Build the command-line tool first.
ARG BASE_IMAGE=sonarsource/sonar-scanner-cli:latest
FROM golang:1.15-alpine AS builder

WORKDIR /build

COPY . .

RUN apk add --update make && make sonar-scanner-action

# Build the image containing both the sonar-scanner and command-line tool.
FROM $BASE_IMAGE AS final

COPY --from=builder /build/bin/sonar-scanner-action /usr/bin/sonar-scanner-action

ENTRYPOINT /usr/bin/sonar-scanner-action
