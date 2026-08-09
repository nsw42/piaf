//go:build release

package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var assets embed.FS

//go:embed templates/*
var templates embed.FS

var templateCache = make(map[string]*template.Template, 0)

// Embedded files don't carry real mtimes (embed.FS reports the zero time for all of them),
// so cache-bust on process start time instead: a new deploy is a new process.
var assetVersion = fmt.Sprintf("%d", time.Now().Unix())

func assetURL(name string) string {
	return fmt.Sprintf("/assets/%s?v=%s", name, assetVersion)
}

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
		funcs := template.FuncMap{"asset": assetURL}
		var err error
		cache, err = template.New("base.templ").Funcs(funcs).ParseFS(dir, "base.templ", templateName)
		if err != nil {
			return nil, err
		}
		templateCache[templateName] = cache
	}
	return cache, nil
}
