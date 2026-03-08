package request

import (
	"strconv"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func ValidateStruct[T any](c *gin.Context, obj T) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.Failed(c, 400, "Invalid request: "+err.Error())
		return false
	}
	return true
}

func ParsePagination(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	offset, _ := strconv.Atoi(strings.TrimSpace(c.Query("offset")))

	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	if limit == 0 {
		limit = 20
	}

	return limit, offset
}
