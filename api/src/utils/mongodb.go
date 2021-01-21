package utils

import (
	"github.com/jbonadiman/finance-bot/src/entities"
	"github.com/jbonadiman/finance-bot/src/repositories"
	"github.com/juju/mgosession"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
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

func NewMongoRepository(p *mgosession.Pool) repositories.Repository {
	return &repo{
		pool: p,
	}
}

func GetPool(maxSessions int) (*mgosession.Pool, error){
	stringConnection := os.Getenv(MongoDbVarName)
	session, err := mgo.Dial(stringConnection)
	if err != nil {
		return nil, err
	}

	var pool = mgosession.NewPool(nil, session, maxSessions)
	return pool, nil
}

func (r *repo) Find(id primitive.ObjectID) (*entities.Transaction, error) {
	result := entities.Transaction{}

	session := r.pool.Session(nil)

	coll := session.DB(os.Getenv("MONGODB_DATABASE")).C("transactions")

	err := coll.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *repo) FindByDate(date time.Time) (*[]entities.Transaction, error) {
	panic("implement me")
}

func (r *repo) FindByCategory(category entities.Category) (*[]entities.Transaction, error) {
	panic("implement me")
}

func (r *repo) FindByPeriod(startDate time.Time, endDate time.Time) (*[]entities.Transaction, error) {
	panic("implement me")
}

func (r *repo) FindAll() (*[]entities.Transaction, error) {
	panic("implement me")
}

func (r *repo) Update(transaction *entities.Transaction) error {
	panic("implement me")
}

func (r *repo) Store(transaction *entities.Transaction) error {
	session := r.pool.Session(nil)

	coll := session.DB(os.Getenv("MONGODB_DATABASE")).C("transactions")

	err := coll.Insert(transaction)
	if err != nil {
		return err
	}

	return nil
}