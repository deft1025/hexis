package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"hexis/internal/handler"
	"hexis/internal/web"
)

func New(db *gorm.DB) *gin.Engine {
	return NewWithGuidePreview(db, false)
}

func NewWithGuidePreview(db *gorm.DB, enabled bool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), CORS(), JSONHeader())

	h := handler.NewWithGuidePreview(db, enabled)

	api := r.Group("/api/v1")
	{
		api.GET("/patches/latest", h.LatestPatch)
		api.GET("/champions", h.Champions)
		api.GET("/champions/:id", h.Champion)
		api.GET("/champions/:id/skills", h.ChampionSkills)
		api.GET("/augments", h.Augments)
		api.GET("/items", h.Items)
		api.GET("/combos", h.Combos)
		api.GET("/guides", h.Guides)
	}

	r.GET("/healthz", h.Health)
	r.NoRoute(gin.WrapH(web.Handler()))
	return r
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func JSONHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" || c.Request.URL.Path == "/healthz" {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}
		c.Next()
	}
}
