package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subcategory struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Keywords []string           `json:"keywords" bson:"keywords"`
	Name     string             `json:"name" bson:"name"`
}
