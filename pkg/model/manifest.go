package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

type Manifest struct {
	BaseModel     `bson:",inline"`
	ApplicationID primitive.ObjectID `json:"application_id"` // 关联 Application
	Name          string             `json:"name"`           // application + 日期 + 随机数
	Version       string             `json:"version"`
	Branch        string             `json:"branch"`      // git branch
	GitRepo       string             `json:"git_repo"`    // 对应 Application repo
	Image         string             `json:"image"`       // Docker 镜像地址
	PipelineID    string             `json:"pipeline_id"` // Tekton PipelineRun ID
	Steps         []Step             `json:"steps"`       // 每个步骤状态
	Status        string             `json:"status"`      // running, success, failed
}

type Step struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pending, running, success, failed
}

func GenerateManifestVersion() string {
	t := time.Now().Format("20060102150405")
	r := rand.Intn(100)
	return fmt.Sprintf("%s-%02d", t, r)
}

func (Manifest) CollectionName() string { return "manifests" }
