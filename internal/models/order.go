package models

import "time"

type Order struct {
	UUID       string
	Number     string
	State      int32
	Accrual    int64
	AccountID  int64
	UploadedAt time.Time
}
