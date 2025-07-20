//go:build !release

package main

import (
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func configureAssetsForRouter(router *gin.Engine, path string) {
	router.Static(path, "./assets")
}

func getTemplate(templateName string) (*template.Template, error) {
	path := filepath.Join("templates", templateName)
	return template.ParseFiles(path)
}
