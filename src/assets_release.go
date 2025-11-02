//go:build release

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var assets embed.FS

//go:embed templates/*
var templates embed.FS

var templateCache = make(map[string]*template.Template, 0)

func configureAssetsForRouter(router *gin.Engine, path string) {
	router.Use(func(c *gin.Context) {
		// Set a cache timeout (1hr)
		if strings.HasPrefix(c.Request.URL.Path, path) {
			c.Header("Cache-Control", "max-age=3600")
		}
		c.Next()
	})
	dir, _ := fs.Sub(assets, "assets")
	router.StaticFS(path, http.FS(dir))
}

func getTemplate(templateName string) (*template.Template, error) {
	// Do we already have it cached?
	cache, ok := templateCache[templateName]
	if !ok {
		// No, so load it and save it for next time
		dir, _ := fs.Sub(templates, "templates")
		var err error
		cache, err = template.ParseFS(dir, "base.templ", templateName)
		if err != nil {
			return nil, err
		}
		templateCache[templateName] = cache
	}
	return cache, nil
}
