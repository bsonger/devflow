package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Application struct {
	BaseModel `bson:",inline"`

	Name               string              `bson:"name" json:"name"`
	RepoURL            string              `bson:"repo_url" json:"repo_url"`
	ActiveManifestName string              `bson:"active_manifest_name" json:"active_manifest_name"`
	ActiveManifestID   *primitive.ObjectID `bson:"active_manifest_id,omitempty" json:"active_manifest_id,omitempty"`
	// 当前状态（来自 Job 的结果）
	Status string `bson:"status" json:"status"` // Running / Failed / Degraded
}

func (Application) CollectionName() string { return "applications" }
