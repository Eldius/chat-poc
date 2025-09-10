package transaction

import (
	"chat-poc/internal/config"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/eldius/initial-config-go/logs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mashiike/redshift-data-sql-driver"
	"github.com/tmc/langchaingo/tools"
)

var (
	_ tools.Tool = &Lookup{}

	//go:embed queries/*.sql
	queries embed.FS
)

type Lookup struct {
	db *sqlx.DB
}

func NewDefaultLookup() (*Lookup, error) {
	dbHost := config.GetDBHost()
	slog.With("host", dbHost, "port", config.GetDBPort(), "db", config.GetDBName()).Debug("Creating transaction lookup")

	url := fmt.Sprintf("sslmode=require user=%v password=%v host=%v port=%v dbname=%v",
		config.GetDBUser(),
		config.GetDBPass(),
		dbHost,
		config.GetDBPort(),
		config.GetDBName())
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("creating transaction lookup: db connection to host %s: %w", dbHost, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("creating transaction lookup: ping db %s: %w", dbHost, err)
	}

	return NewTransactionLookup(db), nil
}

func NewTransactionLookup(db *sqlx.DB) *Lookup {
	return &Lookup{
		db: db,
	}
}

func (t Lookup) Name() string {
	return "transaction_lookup"
}

func (t Lookup) Description() string {
	return `Useful for when you need to retrieve a specific transaction by its unique ID. The input must be a transaction ID, for example, 'tx_12345' or a UUID key.
It will return the transaction details in JSON format.
Here some attributes that can be returned:
transaction_id: transaction id
acquirer: acquirer name
company_id: partner company
workspace_id: partner workspace
application_id: partner application
application_name: partner application name
raw_application_id: partner identification
merchant_id: acquirer merchant identification id
pos: acquirer point of sale identification
type: payment type
amount: payment amount (in cents, for example: 100 = R$ 1.00)
currency: payment currency
installments: payment installments count
recurrent: is this payment recurrent?
soft_descriptor: payment soft descriptor
status: payment status
payment_id: registered payment id
authorization_code: payment authorization code
nsu: payment nsu (Número Sequencial Único)
response_code: payment status code (following ABECS rules)
tokenized: this payment used a network token?
tid: payment tid
date_timestamp: event timestamp in unix milissecond timestamp notation
dt: a text formatted timestamp
refund_amount: total refunded amount (in cents, for example: 100 = R$ 1.00)
retry_after: if payment was denied, it could be retried after this date
has_customer_information: does this payment have customer information?
is_full_refunded: does this payment full refunded?
is_waiting_confirmation: does this payment waiting for confirmation?
is_authenticated: does this payment was authenticated?
is_data_only: does this payment was a data only?
is_denied: does this payment denied?
is_acquirer_fallback: does this payment an acquirer fallback?
is_payment_method_fallback: does this payment a payment method fallback?
is_digital_wallet: does this payment a digital wallet payment?
fallback_reason: the reason payment fallback was made (if it is a fallback)
contains_error: does this payment attemp has error?
card_token: card identification
card_hash: card hash
card_brand: card brand
card_holder: card holder name
card_type: card type
cvv_informed: does the CVV security code informed?
correlation_id: a payment identification similar to transaction_id attribute
trace_id: a code to identify payment request through systems
error_code: payment error code (if occurred)
error_status_code: payment error status code (if occurred)
error_status: error status (if occurred)
error_type: error type (if occurred)
error_operation: error operation (if occurred)
error_message: error message (if occurred)
transaction_date: transaction date
`
}

func (t Lookup) Call(ctx context.Context, input string) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"input": input,
	})

	log.Debug("Looking for transaction")

	q, err := queries.ReadFile("queries/transaction_lookup.sql")
	if err != nil {
		err := fmt.Errorf("looking for transaction: reading query: %w", err)
		log.WithError(err).Error("Looking for transaction -> reading query")
		return "", err
	}
	query := string(q)
	log.Debug("Query: " + query)

	res, err := t.db.NamedQueryContext(ctx, query, map[string]any{"transaction_id": input})
	if err != nil {
		err := fmt.Errorf("looking for transaction: executing query: %w", err)
		log.WithError(err).Error("Looking for transaction -> querying for transaction")
		return "", err
	}
	defer func() {
		_ = res.Close()
	}()

	var outputData []map[string]any
	for res.Next() {
		row := make(map[string]any)
		if err := res.MapScan(row); err != nil {
			err := fmt.Errorf("looking for transaction: mapping results: %w", err)
			log.WithError(err).Error("Looking for transaction -> mapping query results")
			return "", err
		}
		outputData = append(outputData, row)
	}
	b, err := json.Marshal(outputData)
	if err != nil {
		err := fmt.Errorf("looking for transaction: marshaling results: %w", err)
		log.WithError(err).Error("Looking for transaction -> marshaling query results")
		return "", err
	}

	return string(b), nil

}
