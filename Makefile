sonar-scanner-action:
	CGO_ENABLED=0 go build -o bin/sonar-scanner-action action.go

test:
	go test ./...