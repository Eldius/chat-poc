
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

doc-add:
	go run ./cmd/cli/ doc add \
		--path "https://docs.cielo.com.br/ecommerce-cielo/page/abecs" \
		--path "https://api.abecs.org.br/wp-content/uploads/2019/09/Normativo-021.pdf"

doc-query:
	go run ./cmd/cli/ doc query ABECS code
