package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	OriginalTaskID string             `json:"originalId" bson:"originalId"`
	Date           time.Time          `json:"date" bson:"date"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	ModifiedAt     time.Time          `json:"modifiedAt" bson:"modifiedAt"`
	Description    string             `json:"description" bson:"description"`
	Cost           float64            `json:"value" bson:"value"`
	Category       string             `json:"category" bson:"category"`
}
