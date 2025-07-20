//go:build !release

package main

import "github.com/gin-gonic/gin"

func configureAssetsForRouter(router *gin.Engine, path string) {
	router.Static(path, "./assets")
}
