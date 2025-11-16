package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Manifest struct {
	BaseModel     `bson:",inline"`
	ApplicationID primitive.ObjectID `json:"application_id"` // 关联 Application
	Name          string             `json:"name"`           // application + 日期 + 随机数
	Branch        string             `json:"branch"`         // git branch
	GitRepo       string             `json:"git_repo"`       // 对应 Application repo
	Image         string             `json:"image"`          // Docker 镜像地址
	PipelineID    string             `json:"pipeline_id"`    // Tekton PipelineRun ID
	Steps         []Step             `json:"steps"`          // 每个步骤状态
	Status        string             `json:"status"`         // running, success, failed
}

type Step struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pending, running, success, failed
}

// GenerateName 生成 name: application + 日期 + 随机数
func GenerateName(appName string) string {
	t := time.Now().Format("20060102")
	randNum := RandString(6)
	return appName + "-" + t + "-" + randNum
}

func (Manifest) CollectionName() string { return "manifests" }

// RandString 随机字符串
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
