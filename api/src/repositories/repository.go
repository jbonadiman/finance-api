package repositories

import (
	"github.com/jbonadiman/finance-bot/src/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Reader interface {
	Find(id primitive.ObjectID) (*entities.Transaction, error)
	FindByDate(date time.Time) (*[]entities.Transaction, error)
	FindByCategory(category entities.Category) (*[]entities.Transaction, error)
	FindByPeriod(startDate time.Time, endDate time.Time) (*[]entities.Transaction, error)
	FindAll() (*[]entities.Transaction, error)
}

type Writer interface {
	Update(transaction *entities.Transaction) error
	Store(transaction *entities.Transaction) error
}

type Repository interface {
	Reader
	Writer
}