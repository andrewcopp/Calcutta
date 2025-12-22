package models

import "time"

type CalcuttaPayout struct {
	ID          string     `json:"id"`
	CalcuttaID  string     `json:"calcuttaId"`
	Position    int        `json:"position"`
	AmountCents int        `json:"amountCents"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
	Deleted     *time.Time `json:"deleted,omitempty"`
}
