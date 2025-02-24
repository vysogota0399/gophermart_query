package transaction_inbox

import (
	"github.com/caarlos0/env"
)

type Config struct {
	OrderCreatedPollInterval         int   `json:"poll_accrual_finished_inverval" env:"DAEMON_ORDER_CREATED_EVENT_POLL_INTERVAL" envDefault:"250"`
	OrderUpdatedPollInterval         int   `json:"poll_accrual_failed_inverval" env:"DAEMON_ORDER_UPDATED_EVENT_POLL_INTERVAL" envDefault:"250"`
	TransactionProcessedPollInterval int   `json:"poll_accrual_started_inverval" env:"DAEMON_TRANSACTIO_PROCESSED_EVENT_POLL_INTERVAL" envDefault:"250"`
	WorkersCount                     int64 `json:"workersCount" env:"DAEMON_WORKERS_COUNT" envDefault:"5"`

	KafkaOrderCreatedTopic                  string `json:"kafka_order_created_topic" env:"KAFKA_ORDER_CREATED_TOPIC" envDefault:"order_created"`
	KafkaOrderCreatedGroupID                string `json:"kafka_order_created_group_id" env:"KAFKA_ORDER_CREATED_GROUP_ID" envDefault:"query_order_created_consumer_group"`
	KafkaOrderCreatedPartitionWatchInterval int    `json:"kafka_order_created_partition_watch_interval" env:"KAFKA_ORDER_PARTITION_WATCH_INTERWAL" envDefault:"50000"`
	KafkaOrderCreatedMaxWaitInterval        int    `json:"kafka_order_created_max_wait_interval" env:"KAFKA_ORDER_CREATED_MAX_WAIT_INTERWAL" envDefault:"250"`

	KafkaOrderUpdatedTopic                  string `json:"kafka_order_updated_topic" env:"KAFKA_ORDER_UPDATED_TOPIC" envDefault:"order_updated"`
	KafkaOrderUpdatedGroupID                string `json:"kafka_order_updated_group_id" env:"KAFKA_ORDER_UPDATED_GROUP_ID" envDefault:"query_order_updated_consumer_group"`
	KafkaOrderUpdatedPartitionWatchInterval int    `json:"kafka_order_updated_partition_watch_interval" env:"KAFKA_ORDER_UPDATED_PARTITION_WATCH_INTERWAL" envDefault:"50000"`
	KafkaOrderUpdatedMaxWaitInterval        int    `json:"kafka_order_updated_max_wait_interval" env:"KAFKA_ORDER_UPDATED_MAX_WAIT_INTERWAL" envDefault:"250"`

	KafkaTransactionProcessedTopic                  string `json:"kafka_transaction_processed_topic" env:"KAFKA_TRANSACTION_PROCESSED_TOPIC" envDefault:"transaction_processed"`
	KafkaTransactionProcessedGroupID                string `json:"kafka_transaction_processed_group_id" env:"KAFKA_TRANSACTION_PROCESSED_GROUP_ID" envDefault:"query_transaction_processed_consumer_group"`
	KafkaTransactionProcessedPartitionWatchInterval int    `json:"kafka_transaction_processed_partition_watch_interval" env:"KAFKA_TRANSACTION_PROCESSED_PARTITION_WATCH_INTERWAL" envDefault:"50000"`
	KafkaTransactionProcessedMaxWaitInterval        int    `json:"kafka_transaction_processed_max_wait_interval" env:"KAFKA_TRANSACTION_PROCESSED_MAX_WAIT_INTERWAL" envDefault:"250"`
}

func MustNewConfig() *Config {
	c := &Config{}
	env.Parse(c)

	return c
}
