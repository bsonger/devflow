package api

import (
	"errors"
	"net/http"

	"github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ApplicationRouteApi = NewApplicationHandler()

type ApplicationHandler struct {
}

func NewApplicationHandler() *ApplicationHandler {
	return &ApplicationHandler{}
}

type UpdateActiveManifestRequest struct {
	ManifestID string `json:"manifest_id" binding:"required"`
}

// Create
// @Summary 创建应用
// @Description 创建一个新的应用
// @Tags Application
// @Accept json
// @Produce json
// @Param data body model.Application true "Application Data"
// @Success 200 {object} map[string]string
// @Router /api/v1/applications [post]
func (h *ApplicationHandler) Create(c *gin.Context) {
	var app *model.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app.WithCreateDefault()
	id, err := service.ApplicationService.Create(c.Request.Context(), app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id.Hex()})
}

// Get
// @Summary	获取应用
// @Tags		Application
// @Param		id	path		string	true	"Application ID"
// @Success	200	{object}	model.Application
// @Router		/api/v1/applications/{id} [get]
func (h *ApplicationHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	app, err := service.ApplicationService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// Update
// @Summary	更新应用
// @Tags		Application
// @Param		id		path		string				true	"Application ID"
// @Param		data	body		model.Application	true	"Application Data"
// @Success	200		{object}	map[string]string
// @Router		/api/v1/applications/{id} [put]
func (h *ApplicationHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var app model.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app.SetID(id)

	if err := service.ApplicationService.Update(c.Request.Context(), &app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete
// @Summary	删除应用
// @Tags		Application
// @Param		id	path		string	true	"Application ID"
// @Success	200	{object}	map[string]string
// @Router		/api/v1/applications/{id} [delete]
func (h *ApplicationHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := service.ApplicationService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// UpdateActiveManifest
// @Summary	更新应用的 Active Manifest
// @Tags		Application
// @Param		id	path		string	true	"Application ID"
// @Param		data	body		UpdateActiveManifestRequest	true	"Active Manifest Data"
// @Success	200	{object}	map[string]string
// @Router		/api/v1/applications/{id}/active_manifest [patch]
func (h *ApplicationHandler) UpdateActiveManifest(c *gin.Context) {
	appID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateActiveManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manifestID, err := primitive.ObjectIDFromHex(req.ManifestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest_id"})
		return
	}

	if err := service.ApplicationService.UpdateActiveManifest(c.Request.Context(), appID, manifestID); err != nil {
		if errors.Is(err, service.ErrManifestNotForApplication) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// List
// @Summary 获取应用列表
// @Tags    Application
// @Success 200 {array} model.Application
// @Router  /api/v1/applications [get]
func (h *ApplicationHandler) List(c *gin.Context) {
	filter := primitive.M{}
	if !includeDeleted(c) {
		filter["deleted_at"] = primitive.M{"$exists": false}
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter["project_name"] = projectName
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if releaseType := c.Query("type"); releaseType != "" {
		filter["type"] = releaseType
	}
	if repoURL := c.Query("repo_url"); repoURL != "" {
		filter["repo_url"] = repoURL
	}

	apps, err := service.ApplicationService.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paging, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total := len(apps)
	apps = paginateSlice(apps, paging)
	setPaginationHeaders(c, total, paging)

	c.JSON(http.StatusOK, apps)
}
