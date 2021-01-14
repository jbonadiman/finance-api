package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Reader interface {
	Find(id primitive.ObjectID) (*Transaction, error)
	FindByDate(date time.Time) (*[]Transaction, error)
	FindByCategory(category Category) (*[]Transaction, error)
	FindByPeriod(startDate time.Time, endDate time.Time) (*[]Transaction, error)
	FindAll() (*[]Transaction, error)
}

type Writer interface {
	Update(transaction *Transaction) error
	Store(transaction *Transaction) error
}

type Repository interface {
	Reader
	Writer
}