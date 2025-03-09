package models

import (
	"fmt"
	"time"

	"github.com/vysogota0399/gophermart_protos/utils/amount"
)

type Transaction struct {
	UUID        string
	OrderNumber string
	AccountID   int64
	Amount      *TransactionAmount
	Operation   string
	CreatedAt   time.Time
	ProcessedAt time.Time
}

type TransactionAmount struct {
	*amount.Amount
}

func (amnt *TransactionAmount) Scan(value interface{}) error {
	if value == nil {
		*amnt = TransactionAmount{Amount: &amount.Amount{}}
		return nil
	}

	nanoBonuses, ok := value.(int64)
	if !ok {
		return fmt.Errorf("models/transaction: nanoBonuses invalid format error, expected int64")
	}

	amountContainer:= amount.FromInt64(nanoBonuses)

	*amnt = TransactionAmount{Amount: amountContainer}
	return nil
}
