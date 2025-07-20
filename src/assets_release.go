//go:build release

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var assets embed.FS

//go:embed templates/*
var templates embed.FS

func configureAssetsForRouter(router *gin.Engine, path string) {
	dir, _ := fs.Sub(assets, "assets")
	router.StaticFS(path, http.FS(dir))
}

func getTemplate(templateName string) (*template.Template, error) {
	dir, _ := fs.Sub(templates, "templates")
	return template.ParseFS(dir, templateName)
}
