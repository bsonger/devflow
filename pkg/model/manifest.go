package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"strconv"
	"time"
)

type ManifestStatus string

const (
	ManifestPending   ManifestStatus = "Pending"
	ManifestRunning   ManifestStatus = "Running"
	ManifestSucceeded ManifestStatus = "Succeeded"
	ManifestFailed    ManifestStatus = "Failed"
)

type StepStatus string

const (
	StepPending   StepStatus = "Pending"
	StepRunning   StepStatus = "Running"
	StepSucceeded StepStatus = "Succeeded"
	StepFailed    StepStatus = "Failed"
)

type Manifest struct {
	BaseModel       `bson:",inline"`
	ApplicationId   primitive.ObjectID `json:"application_id" bson:"application_id"` // 关联 Application
	Name            string             `json:"name" bson:"name"`
	ApplicationName string             `json:"application_name" bson:"application_name"`
	Version         string             `json:"version" bson:"version"`
	Branch          string             `json:"branch" bson:"branch"`           // git branch
	GitRepo         string             `json:"git_repo" bson:"git_repo"`       // 对应 Application repo
	Image           string             `json:"image" bson:"image"`             // Docker 镜像地址
	PipelineID      string             `json:"pipeline_id" bson:"pipeline_id"` // Tekton PipelineRun ID
	Steps           []ManifestStep     `json:"steps" bson:"steps"`             // 每个步骤状态
	Status          ManifestStatus     `json:"status" bson:"status"`           // running, success, failed
}

type ManifestStep struct {
	TaskName  string     `bson:"task_name" json:"task_name"`
	TaskRun   string     `bson:"task_run,omitempty" json:"task_run,omitempty"`
	Status    StepStatus `bson:"status" json:"status"`
	StartTime *time.Time `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime   *time.Time `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Message   string     `bson:"message,omitempty" json:"message,omitempty"`
}

func GenerateManifestVersion(name string) string {
	t := time.Now().Format("20060102150405")
	r := rand.Intn(100)
	return fmt.Sprintf("%s%s%s", name, t, strconv.Itoa(r))
}

func (Manifest) CollectionName() string { return "manifests" }
