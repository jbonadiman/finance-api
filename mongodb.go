package main

import (
	"github.com/juju/mgosession"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

const (
	MongoDbVarName = "MONGODB_CON"
)

type repo struct {
	pool *mgosession.Pool
}

func NewMongoRepository(p *mgosession.Pool) Repository {
	return &repo {
		pool: p,
	}
}

func (r *repo) Find(id primitive.ObjectID) (*Transaction, error) {
	result := Transaction{}

	session := r.pool.Session(nil)

	coll := session.DB(os.Getenv("MONGODB_DATABASE")).C("transactions")

	err := coll.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *repo) FindByDate(date time.Time) ([]*Transaction, error) {
	panic("implement me")
}

func (r *repo) FindByCategory(category Category) ([]*Transaction, error) {
	panic("implement me")
}

func (r *repo) FindByPeriod(startDate time.Time, endDate time.Time) ([]*Transaction, error) {
	panic("implement me")
}

func (r *repo) FindAll() ([]*Transaction, error) {
	panic("implement me")
}

func (r *repo) Update(transaction *Transaction) error {
	panic("implement me")
}

func (r *repo) Store(transaction *Transaction) error {
	panic("implement me")
}