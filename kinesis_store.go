package theta

import (
	consumer "github.com/harlow/kinesis-consumer"
	psql "github.com/harlow/kinesis-consumer/store/postgres"
)

// KinesisCollectorStore represents a kinesis collector store
type KinesisCollectorStore = consumer.Store

// KinesisCollectorWithStore creates a store option
func KinesisCollectorWithStore(store KinesisCollectorStore) KinesisCollectorOption {
	return consumer.WithStore(store)
}

// NewKinesisCollectorPSQLStore creates a new PostgreSQL store
func NewKinesisCollectorPSQLStore(name, table string, url string) (KinesisCollectorStore, error) {
	return psql.New(name, table, url)
}
