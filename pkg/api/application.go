package api

import (
	"net/http"

	"github.com/bsonger/devflow/pkg/model"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ApplicationRouteApi = NewApplicationHandler()

type ApplicationHandler struct {
	svc *service.ApplicationService
}

func NewApplicationHandler() *ApplicationHandler {
	return &ApplicationHandler{svc: service.NewApplicationService()}
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
	id, err := h.svc.Create(c.Request.Context(), app)
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

	app, err := h.svc.Get(c.Request.Context(), id)
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

	if err := h.svc.Update(c.Request.Context(), &app); err != nil {
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

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
