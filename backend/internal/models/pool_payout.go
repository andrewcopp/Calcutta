package models

import "time"

type PoolPayout struct {
	ID          string     `json:"id"`
	PoolID      string     `json:"poolId"`
	Position    int        `json:"position"`
	AmountCents int        `json:"amountCents"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}
