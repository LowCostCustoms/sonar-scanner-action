sonar-scanner-adapter:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o sonar-scanner-adapter main.go

test:
	go test