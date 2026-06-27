
clear-log:
	-rm execution.log

chat: clear-log
	CHAT_DB_USER=$(DB_USER) \
		CHAT_DB_PASS=$(DB_PASS) \
			go run ./cmd/cli/ chat

debug:
	dlv debug --headless --listen=:40237 --api-version=2 --accept-multiclient ./cmd/cli/ -- chat --session "testing session"

doc-add:
	go run ./cmd/cli/ doc add \
		--path "https://docs.cielo.com.br/ecommerce-cielo/page/abecs" \
		--path "https://api.abecs.org.br/wp-content/uploads/2019/09/Normativo-021.pdf" \
		--path "https://www.coursera.org/resources/markdown-cheat-sheet" \
		--path "https://atendimento.vindi.com.br/hc/pt-br/articles/360016739851-Motivos-C%C3%B3digos-de-rejei%C3%A7%C3%A3o-pelas-operadoras-de-cart%C3%A3o-de-cr%C3%A9dito" \
		--path "https://docs.cielo.com.br/ecommerce-cielo/page/abecs" \
		--path "https://www.maxipago.com/developers/retorno-abecs/" \
		--path "https://docs.cielo.com.br/ecommerce-cielo/reference/retentativa-bandeira" \
		--path "https://sites.google.com/hyperlocal.com.br/central-de-ajuda-avec/avecpay/m%C3%A1quina/entenda-alguns-erros-que-podem-aparecer-na-sua-m%C3%A1quina" \
		--path "https://paymentcloudinc.com/blog/credit-card-decline-codes/error-code-63/" \
		--path "https://www.maxipago.com/developers/retorno-abecs/" \
		--path "https://paymentcloudinc.com/blog/credit-card-decline-codes/"
doc-query:
	go run ./cmd/cli/ doc query ABECS code

snapshot:
	goreleaser release --snapshot --clean

cache-ls:
	go run ./cmd/cli/ cache ls

test:
	go test ./... -cover

vulncheck:
	go tool govulncheck ./...

lint:
	golangci-lint run

validate: test lint vulncheck
	@echo "Validation finished with success..."
