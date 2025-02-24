package models

import "time"

type Order struct {
	UUID       string
	Number     string
	State      string
	Accrual    int64
	AccountID  int64
	UploadedAt time.Time
}
