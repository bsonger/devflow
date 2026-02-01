package api

import (
	"net/http"

	"github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var JobRouteApi = NewJobHandler()

type JobHandler struct {
}

func NewJobHandler() *JobHandler {
	return &JobHandler{}
}

// Create
// @Summary 创建Job
// @Description 创建一个新的Job
// @Tags Job
// @Accept json
// @Produce json
// @Param data body model.Job true "Job Data"
// @Success 200 {object} map[string]string
// @Router /api/v1/jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	var job *model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job.WithCreateDefault()
	id, err := service.JobService.Create(c.Request.Context(), job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id.Hex()})
}

// Get
// @Summary	获取Job
// @Tags		Job
// @Param		id	path		string	true	"Job ID"
// @Success	200	{object}	model.Job
// @Router		/api/v1/jobs/{id} [get]
func (h *JobHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	job, err := service.JobService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// Update
// @Summary	更新Job
// @Tags		Job
// @Param		id		path		string				true	"Job ID"
// @Param		data	body		model.Job	true	"Job Data"
// @Success	200		{object}	map[string]string
// @Router		/api/v1/jobs/{id} [put]
func (h *JobHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var job model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job.SetID(id)

	if err := service.JobService.Update(c.Request.Context(), &job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete
// @Summary	删除Job
// @Tags		Job
// @Param		id	path		string	true	"Job ID"
// @Success	200	{object}	map[string]string
// @Router		/api/v1/jobs/{id} [delete]
func (h *JobHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := service.JobService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List
// @Summary 获取Job列表
// @Tags    Job
// @Success 200 {array} model.Job
// @Router  /api/v1/jobs [get]
func (h *JobHandler) List(c *gin.Context) {
	filter := primitive.M{}
	if !includeDeleted(c) {
		filter["deleted_at"] = primitive.M{"$exists": false}
	}
	if appID := c.Query("application_id"); appID != "" {
		id, err := primitive.ObjectIDFromHex(appID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application_id"})
			return
		}
		filter["application_id"] = id
	}
	if manifestID := c.Query("manifest_id"); manifestID != "" {
		id, err := primitive.ObjectIDFromHex(manifestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest_id"})
			return
		}
		filter["manifest_id"] = id
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if jobType := c.Query("type"); jobType != "" {
		filter["type"] = jobType
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter["project_name"] = projectName
	}
	if appName := c.Query("application_name"); appName != "" {
		filter["application_name"] = appName
	}

	jobs, err := service.JobService.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paging, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total := len(jobs)
	jobs = paginateSlice(jobs, paging)
	setPaginationHeaders(c, total, paging)

	c.JSON(http.StatusOK, jobs)
}
