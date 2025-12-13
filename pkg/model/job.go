package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	BaseModel `bson:",inline"`

	ApplicationId   primitive.ObjectID `bson:"application_id" json:"application_id"`
	ApplicationName string             `bson:"application_name" json:"application_name"`
	ProjectName     string             `bson:"project_name" json:"project_name"`
	ManifestID      primitive.ObjectID `bson:"manifest_id" json:"manifest_id"`
	ManifestName    string             `bson:"manifest_name" json:"manifest_name"`
	Type            string             `bson:"type" json:"type"`
	Status          string             `bson:"status" json:"status"`
}

func (*Job) CollectionName() string { return "job" }
