package gindocs

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Mount registers Gin Docs routes on the given router.
// db is optional — pass nil if not using GORM models.
// configs is variadic — pass zero or one Config.
func Mount(router *gin.Engine, db *gorm.DB, configs ...Config) *GinDocs {
	cfg := mergeConfig(configs...)

	gd := newGinDocs(router, db, cfg)
	gd.registerHandlers()

	return gd
}
