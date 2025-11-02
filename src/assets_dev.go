//go:build !release

package main

import (
	"html/template"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func configureAssetsForRouter(router *gin.Engine, path string) {
	router.Use(func(c *gin.Context) {
		// Set a short cache timeout (1min)
		if strings.HasPrefix(c.Request.URL.Path, path) {
			c.Header("Cache-Control", "max-age=60")
		}
		c.Next()
	})
	router.Static(path, "./assets")
}

func getTemplate(templateName string) (*template.Template, error) {
	base := filepath.Join("templates", "base.templ")
	path := filepath.Join("templates", templateName)
	return template.ParseFiles(path, base)
}
