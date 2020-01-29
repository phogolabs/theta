package theta

import (
	consumer "github.com/harlow/kinesis-consumer"
	psql "github.com/harlow/kinesis-consumer/store/postgres"
)

// KinesisCollectorStore represents a kinesis collector store
type KinesisCollectorStore = consumer.Store

// NewKinesisCollectorPSQLStore creates a new PostgreSQL store
func NewKinesisCollectorPSQLStore(name, table string, url string) (KinesisCollectorStore, error) {
	return psql.New(name, table, url)
}
