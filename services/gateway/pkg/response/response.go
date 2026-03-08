package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Time    int64  `json:"time"`
}

func Success(c *gin.Context, code int, data any) {
	response := Response{
		Code:    code,
		Message: "success",
		Data:    data,
		Time:    time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

func Failed(c *gin.Context, code int, message string) {
	response := Response{
		Code:    code,
		Message: message,
		Data:    nil,
		Time:    time.Now().Unix(),
	}
	c.JSON(http.StatusInternalServerError, response)
}
