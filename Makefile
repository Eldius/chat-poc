
testing:
	go run ./cmd/testing/

chat:
	AWS_REGION=us-east-1 go run ./cmd/cli/ chat

debug:
	dlv debug --headless --listen=:40237 --api-version=2 --accept-multiclient ./cmd/testing/
