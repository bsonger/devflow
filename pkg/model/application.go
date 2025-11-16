package model

type Application struct {
	BaseModel `bson:",inline"`

	Name    string `bson:"name" json:"name"`
	RepoURL string `bson:"repo_url" json:"repo_url"`
}

func (Application) CollectionName() string { return "applications" }
