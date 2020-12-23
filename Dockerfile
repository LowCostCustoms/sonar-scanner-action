# Build the command-line tool first.
FROM golang:1.15-apline AS builder

WORKDIR /build

COPY . .

RUN make

# Build the image containing both the sonar-scanner and command-line tool.
ARG BASE_IMAGE=sonarsource/sonar-scanner-cli:latest

FROM $BASE_IMAGE AS final

COPY --from=builder /build/sonar-scanner-adapter /usr/bin/sonar-scanner-adapter

ENTRYPOINT /usr/bin/sonar-scanner-adapter
