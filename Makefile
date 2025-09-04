
testing:
	go run ./cmd/testing/

chat:
	AWS_REGION=us-east-1 \
		CHAT_DB_USER=$(DB_USER) \
		CHAT_DB_PASS=$(DB_PASS) \
			go run ./cmd/cli/ chat --session "testing session"

debug:
	AWS_REGION=us-east-1 dlv debug --headless --listen=:40237 --api-version=2 --accept-multiclient ./cmd/cli/ -- chat --session "testing session"
	#dlv debug --headless --listen=:40237 --api-version=2 --accept-multiclient ./cmd/testing/
