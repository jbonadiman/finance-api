package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Transaction struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	Date        time.Time          `json:"date" bson:"date"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	ModifiedAt  time.Time          `json:"modifiedAt" bson:"modifiedAt"`
	Description string             `json:"description" bson:"description"`
	Value       float64            `json:"value" bson:"value"`
	Category    Category           `json:"category" bson:"category"`
}

type Category struct {
	Name        string    `json:"name" bson:"name"`
	Subcategory *Category `json:"subcategory" bson:"subcategory"`
}
