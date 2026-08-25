package response

import "github.com/gin-gonic/gin"

// JSON writes v as a JSON response with the given status code.
func JSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

// Error writes a JSON error payload with the given status code.
func Error(c *gin.Context, status int, message string) {
	JSON(c, status, gin.H{"error": message})
}
