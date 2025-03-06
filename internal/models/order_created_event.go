package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	OrderEventNewState int32 = iota
	OrderEventProcessingState
	OrderEventFinishedState
	OrderEventFailedState
)

type OrderCreatedEvent struct {
	UUID  string
	State int32
	Name  string
	Meta  *OrderCreatedEventMeta
}

type OrderCreatedEventMeta struct {
	UUID       string
	Number     string
	State      int32
	Accrual    int64
	UploadedAt time.Time
	AccountID  int64
}

func (m *OrderCreatedEventMeta) Scan(value interface{}) error {
	if value == nil {
		*m = OrderCreatedEventMeta{}
		return nil
	}

	b, ok := value.(string)

	if !ok {
		return fmt.Errorf("models/order_created_event: meta invalid format error, expected json")
	}

	return json.Unmarshal([]byte(b), &m)
}

func (m OrderCreatedEventMeta) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("models/event meta json marshal error %w", err)
	}

	return b, nil
}
