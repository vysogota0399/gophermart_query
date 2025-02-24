package models

import (
	"encoding/json"
	"fmt"
)

type OrderUpdatedEvent struct {
	UUID  string
	State string
	Name  string
	Meta  *OrderUpdatedEventMeta
}

type OrderUpdatedEventMeta struct {
	State string
	UUID  string
}

func (m *OrderUpdatedEventMeta) Scan(value interface{}) error {
	if value == nil {
		*m = OrderUpdatedEventMeta{}
		return nil
	}

	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("models/order_updated_event: meta invalid format error, expected json")
	}

	return json.Unmarshal([]byte(b), &m)
}
