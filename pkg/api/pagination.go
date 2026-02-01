package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultPageSize = 20

type pagination struct {
	enabled  bool
	limit    int
	offset   int
	page     int
	pageSize int
}

func parsePagination(c *gin.Context) (pagination, error) {
	var p pagination

	limitStr := strings.TrimSpace(c.Query("limit"))
	offsetStr := strings.TrimSpace(c.Query("offset"))
	pageStr := strings.TrimSpace(c.Query("page"))
	pageSizeStr := strings.TrimSpace(c.Query("page_size"))

	if limitStr != "" || offsetStr != "" || pageStr != "" || pageSizeStr != "" {
		p.enabled = true
	}

	if limitStr != "" || offsetStr != "" {
		limit := defaultPageSize
		if limitStr != "" {
			parsed, err := strconv.Atoi(limitStr)
			if err != nil || parsed < 1 {
				return pagination{}, fmt.Errorf("invalid limit")
			}
			limit = parsed
		}

		offset := 0
		if offsetStr != "" {
			parsed, err := strconv.Atoi(offsetStr)
			if err != nil || parsed < 0 {
				return pagination{}, fmt.Errorf("invalid offset")
			}
			offset = parsed
		}

		p.limit = limit
		p.offset = offset
		p.pageSize = limit
		p.page = (offset / limit) + 1
		return p, nil
	}

	if pageStr != "" || pageSizeStr != "" {
		page := 1
		if pageStr != "" {
			parsed, err := strconv.Atoi(pageStr)
			if err != nil || parsed < 1 {
				return pagination{}, fmt.Errorf("invalid page")
			}
			page = parsed
		}

		pageSize := defaultPageSize
		if pageSizeStr != "" {
			parsed, err := strconv.Atoi(pageSizeStr)
			if err != nil || parsed < 1 {
				return pagination{}, fmt.Errorf("invalid page_size")
			}
			pageSize = parsed
		}

		p.page = page
		p.pageSize = pageSize
		p.limit = pageSize
		p.offset = (page - 1) * pageSize
		return p, nil
	}

	return p, nil
}

func paginateSlice[T any](items []T, p pagination) []T {
	if !p.enabled {
		return items
	}
	if p.offset >= len(items) {
		return []T{}
	}

	end := p.offset + p.limit
	if end > len(items) {
		end = len(items)
	}
	return items[p.offset:end]
}

func setPaginationHeaders(c *gin.Context, total int, p pagination) {
	if !p.enabled {
		return
	}

	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Page", strconv.Itoa(p.page))
	c.Header("X-Page-Size", strconv.Itoa(p.pageSize))
	c.Header("X-Limit", strconv.Itoa(p.limit))
	c.Header("X-Offset", strconv.Itoa(p.offset))
}

func includeDeleted(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.Query("include_deleted")), "true")
}
