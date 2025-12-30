package handler

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
)

const csrfTokenKey = "csrf_token"
const csrfTokenFormKey = "_csrf"

func RenderHTML(c *gin.Context, code int, name string, obj gin.H) {
	if obj == nil {
		obj = gin.H{}
	}

	if startTime, exists := c.Get("startTime"); exists {
		duration := time.Since(startTime.(time.Time))
		obj["processing_time"] = fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0)
	} else {
		obj["processing_time"] = "unknown"
	}

	if token, exists := c.Get(csrfTokenKey); exists {
		obj["csrfToken"] = token.(string)
		obj["csrfField"] = template.HTML(`<input type="hidden" name="` + csrfTokenFormKey + `" value="` + token.(string) + `">`)
	}

	c.HTML(code, name, obj)
}

