package models

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	TransactionEventNewState        = "new"
	TransactionEventProcessingState = "processing"
	TransactionEventFinishedState   = "finished"
	TransactionEventFailedState     = "failed"
)

type TransactionEvent struct {
	UUID  string                `json:"uuid"`
	State string                `json:"state"`
	Name  string                `json:"name"`
	Meta  *TransactionEventMeta `json:"meta"`
}

type TransactionEventMeta struct {
	UUID            string    `json:"uuid"`
	Amount          int64     `json:"amount"`
	Operation       string    `json:"operation"`
	AccountID       int64     `json:"account_id"`
	TransactionUUID string    `json:"transaction_uuid"`
	OrderNumber     string    `json:"order_number"`
	ProcessedAt     time.Time `json:"processed_at"`
}

func (m *TransactionEventMeta) Scan(value interface{}) error {
	if value == nil {
		*m = TransactionEventMeta{}
		return nil
	}

	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("models/transaction_event: meta invalid format error, expected json")
	}

	return json.Unmarshal([]byte(b), &m)
}
