package models

import "time"

type Transaction struct {
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"createdAt"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	Description string    `json:"description"`
	Cost        float64   `json:"value"`
	Category    string    `json:"category"`
}
