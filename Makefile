sonar-scanner-adapter:
	CGO_ENABLED=0 go build -o bin/sonar-scanner-adapter main.go

test:
	go test ./...