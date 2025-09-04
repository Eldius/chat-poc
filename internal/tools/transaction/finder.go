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
	return "Useful for when you need to retrieve a specific transaction by its unique ID. The input must be a transaction ID, for example, 'tx_12345' or a UUID key."
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

	//return "", errors.New("transaction not found")
}
