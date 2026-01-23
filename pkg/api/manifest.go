package api

import (
	"errors"
	"github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// 保存 Manifest
	id, err := service.ManifestService.CreateManifest(c.Request.Context(), &m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id.Hex()})
}

// List
// @Summary 获取应用列表
// @Tags    Manifest
// @Success 200 {array} model.Manifest
// @Router  /api/v1/manifests [get]
func (h *ManifestHandler) List(c *gin.Context) {
	manifests, err := service.ManifestService.List(c.Request.Context(), primitive.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(manifests) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []model.Manifest{}})
	}
	c.JSON(http.StatusOK, manifests)
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

// Patch
// @Summary		Patch Manifest
// @Description	部分更新 Manifest（仅支持 digest / commit_hash）
// @Tags		Manifest
// @Accept		json
// @Produce		json
// @Param		id		path		string			true	"Manifest ID"
// @Param		data	body		model.PatchManifestRequest	false	"Patch 数据"
// @Success		200		{object}	map[string]string
// @Failure		400		{object}	map[string]string
// @Failure		404		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/api/v1/manifests/{id} [patch]
func (h *ManifestHandler) Patch(c *gin.Context) {

	// 1️⃣ 解析 ID
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 2️⃣ 解析 Patch Body
	var patch model.PatchManifestRequest
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3️⃣ 调用 Service Patch
	err = service.ManifestService.Patch(
		c.Request.Context(),
		id,
		&patch,
	)
	if err != nil {
		// 不存在
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "manifest not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4️⃣ 返回成功
	c.JSON(http.StatusOK, gin.H{
		"message": "patched",
		"id":      id.Hex(),
	})
}
