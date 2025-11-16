package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoModel interface {
	CollectionName() string
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
}

type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
}

func (b BaseModel) GetID() primitive.ObjectID    { return b.ID }
func (b *BaseModel) SetID(id primitive.ObjectID) { b.ID = id }

func (b *BaseModel) WithCreateDefault() {
	b.CreatedAt = time.Now()
	b.WithUpdateDefault()
}

func (b *BaseModel) WithUpdateDefault() {
	b.UpdatedAt = time.Now()
}
