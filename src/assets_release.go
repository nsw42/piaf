//go:build release

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var assets embed.FS

func configureAssetsForRouter(router *gin.Engine, path string) {
	dir, _ := fs.Sub(assets, "assets")
	fmt.Println("assets", dir)
	router.StaticFS(path, http.FS(dir))
}
