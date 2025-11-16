package api

import (
	"github.com/bsonger/devflow/pkg/model"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

var ManifestRouteApi = NewManifestHandler()

type ManifestHandler struct {
	svc *service.ManifestService
}

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{svc: service.NewManifestService()}
}

// Create
// @Summary      创建 Manifest
// @Description  根据 Application 创建 Manifest，自动生成名称
// @Tags         Manifest
// @Accept       json
// @Produce      json
// @Param        application_id  path  string             true "Application ID"
// @Param        data            body  model.Manifest    true "Manifest 数据（branch 必填）"
// @Success      200  {object}  model.Manifest
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/applications/{application_id}/manifests [post]
func (h *ManifestHandler) Create(c *gin.Context) {

	appID, err := primitive.ObjectIDFromHex(c.Param("application_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	var m model.Manifest
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if m.Branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch is required"})
		return
	}

	m.ApplicationID = appID

	// 根据 Application 获取 gitRepo
	application, err := ApplicationRouteApi.svc.Get(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application is not found"})
		return
	}

	m.GitRepo = application.RepoURL

	// 自动生成 Manifest 名称
	m.Name = model.GenerateName("application")

	// 初始化 Steps
	m.Steps = []model.Step{
		{Name: "build_image", Status: "pending"},
		{Name: "package_manifest", Status: "pending"},
		{Name: "push_github", Status: "pending"},
		{Name: "push_aliyun", Status: "pending"},
	}

	m.Status = "running"

	// 保存 Manifest
	id, err := h.svc.CreateManifest(c.Request.Context(), &m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步触发 Pipeline
	//go h.svc.TriggerPipeline(&m)

	c.JSON(http.StatusOK, gin.H{"id": id.Hex()})
}
