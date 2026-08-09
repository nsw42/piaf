//go:build !release

package main

import (
	"fmt"
	"html/template"
	"os"
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

func assetURL(name string) string {
	// Cache-bust on the asset's own mtime, so edits (eg sass -w rebuilds) are picked up immediately
	info, err := os.Stat(filepath.Join("assets", name))
	if err != nil {
		return "/assets/" + name
	}
	return fmt.Sprintf("/assets/%s?v=%d", name, info.ModTime().Unix())
}

func getTemplate(templateName string) (*template.Template, error) {
	base := filepath.Join("templates", "base.templ")
	path := filepath.Join("templates", templateName)
	funcs := template.FuncMap{"asset": assetURL}
	return template.New(templateName).Funcs(funcs).ParseFiles(path, base)
}
