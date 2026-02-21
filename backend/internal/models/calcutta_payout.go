package models

import "time"

type CalcuttaPayout struct {
	ID          string     `json:"id"`
	CalcuttaID  string     `json:"calcuttaId"`
	Position    int        `json:"position"`
	AmountCents int        `json:"amountCents"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}
