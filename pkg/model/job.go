package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	BaseModel `bson:",inline"`

	ApplicationId   primitive.ObjectID `bson:"application_id"`
	ApplicationName string             `bson:"application_name"`
	ProjectName     string             `bson:"project_name"`
	ManifestID      primitive.ObjectID `bson:"manifest_id"`
	ManifestName    string             `bson:"manifest_name"`
	Type            string             `bson:"type"`
}

func (*Job) CollectionName() string { return "job" }
