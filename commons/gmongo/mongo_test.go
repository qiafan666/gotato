package gmongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AccountInterface interface {
	Create(ctx context.Context, accounts ...*Account) error
	Take(ctx context.Context, userId string) (*Account, error)
	Update(ctx context.Context, userID string, data map[string]any) error
	UpdatePassword(ctx context.Context, userId string, password string) error
	Delete(ctx context.Context, userIDs []string) error
	GetCountUser(ctx context.Context) (int64, error)
}

type Account struct {
	UserID         string    `bson:"user_id"`
	Password       string    `bson:"password"`
	PayPassword    string    `bson:"pay_password"`
	CreateTime     time.Time `bson:"create_time"`
	ChangeTime     time.Time `bson:"change_time"`
	OperatorUserID string    `bson:"operator_user_id"`
}

func (Account) TableName() string {
	return "account"
}

func NewAccount(db *mongo.Database) (AccountInterface, error) {
	coll := db.Collection((&Account{}).TableName())
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &AccountImpl{coll: coll}, nil
}

type AccountImpl struct {
	coll *mongo.Collection
}

func (o *AccountImpl) Create(ctx context.Context, accounts ...*Account) error {
	return InsertMany(ctx, o.coll, accounts)
}

func (o *AccountImpl) Take(ctx context.Context, userId string) (*Account, error) {
	return FindOne[*Account](ctx, o.coll, bson.M{"user_id": userId})
}

func (o *AccountImpl) Update(ctx context.Context, userID string, data map[string]any) error {
	if len(data) == 0 {
		return nil
	}
	return UpdateOne(ctx, o.coll, bson.M{"user_id": userID}, bson.M{"$set": data}, false)
}

func (o *AccountImpl) UpdatePassword(ctx context.Context, userId string, password string) error {
	return UpdateOne(ctx, o.coll, bson.M{"user_id": userId}, bson.M{"$set": bson.M{"password": password, "change_time": time.Now()}}, false)
}

func (o *AccountImpl) Delete(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	return DeleteMany(ctx, o.coll, bson.M{"user_id": bson.M{"$in": userIDs}})
}

func (o *AccountImpl) GetCountUser(ctx context.Context) (int64, error) {
	return Count(ctx, o.coll, bson.M{})
}
