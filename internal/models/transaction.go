package models

import "time"

type Transaction struct {
	UUID        string
	OrderNumber string
	AccountID   int64
	Amount      int64
	Operation   string
	CreatedAt   time.Time
	ProcessedAt time.Time
}
