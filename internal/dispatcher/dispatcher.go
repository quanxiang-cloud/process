package dispatcher

import "github.com/gin-gonic/gin"

// Dispatcher Dispatcher
type Dispatcher interface {
	Handler(c *gin.Context)
}
