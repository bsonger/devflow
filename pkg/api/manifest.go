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
}

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{}
}

// Create
// @Summary      创建 Manifest
// @Description  根据 Manifest 创建 Manifest，自动生成名称
// @Tags         Manifest
// @Accept       json
// @Produce      json
// @Param        data            body  model.Manifest    true "Manifest 数据（branch 必填）"
// @Success      200  {object}  model.Manifest
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/manifests [post]
func (h *ManifestHandler) Create(c *gin.Context) {

	var m model.Manifest
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if m.Branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch is required"})
		return
	}

	//// 根据 Application 获取 gitRepo

	//// 初始化 Steps
	//m.Steps = []model.Step{
	//	{Name: "build_image", Status: "pending"},
	//	{Name: "package_manifest", Status: "pending"},
	//	{Name: "push_github", Status: "pending"},
	//	{Name: "push_aliyun", Status: "pending"},
	//}

	//m.Status = "running"

	// 保存 Manifest
	id, err := service.ManifestService.CreateManifest(c.Request.Context(), &m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步触发 Pipeline
	//go service.ManifestService.TriggerPipeline(&m)

	c.JSON(http.StatusOK, gin.H{"id": id.Hex()})
}

// List
// @Summary 获取应用列表
// @Tags    Manifest
// @Success 200 {array} model.Manifest
// @Router  /api/v1/manifests [get]
func (h *ManifestHandler) List(c *gin.Context) {
	apps, err := service.ManifestService.List(c.Request.Context(), primitive.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apps)
}

// Get
// @Summary	获取应用
// @Tags		Manifest
// @Param		id	path		string	true	"Manifest ID"
// @Success	200	{object}	model.Manifest
// @Router		/api/v1/manifests/{id} [get]
func (h *ManifestHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	app, err := service.ManifestService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}
