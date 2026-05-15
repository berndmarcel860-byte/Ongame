package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web/*
var webFS embed.FS

func (r *Router) registerUIRoutes(e *gin.Engine) {
	assets, err := fs.Sub(webFS, "web")
	if err != nil {
		panic("failed to load embedded web assets")
	}

	e.StaticFS("/static", http.FS(assets))
	e.GET("/", r.servePage(http.FS(assets), "landing.html"))
	e.GET("/app", r.servePage(http.FS(assets), "app.html"))
	e.GET("/admin-dashboard", r.servePage(http.FS(assets), "admin.html"))
}

func (r *Router) servePage(filesystem http.FileSystem, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.FileFromFS(name, filesystem)
	}
}
