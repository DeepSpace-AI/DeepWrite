package main

import (
	"github.com/deepwrite/serivces/gateway/bootstrap"
	_ "github.com/deepwrite/serivces/gateway/docs"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	gatewaylogger "github.com/deepwrite/serivces/gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// @title           DeepWrite Gateway API
// @version         1.0

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer {token}" 格式的访问令牌
func main() {
	cfg := config.GetGlobalConfig()
	if err := gatewaylogger.InitLogger(cfg.AppEnv, gatewaylogger.Config{
		Dir:        cfg.Logger.Dir,
		Filename:   cfg.Logger.Filename,
		Level:      cfg.Logger.Level,
		MaxSizeMB:  cfg.Logger.MaxSizeMB,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAgeDays: cfg.Logger.MaxAgeDays,
		Compress:   cfg.Logger.Compress,
	}); err != nil {
		panic(err)
	}
	defer gatewaylogger.Sync()

	if cfg.AppEnv == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	bootstrap.SetupRouter(r)

	bootstrap.SetupDB()

	if err := bootstrap.SetupCache(*cfg); err != nil {
		panic(err)
	}

	if err := bootstrap.SetupStorage(*cfg); err != nil {
		panic(err)
	}

	if err := bootstrap.SetupMailer(*cfg); err != nil {
		panic(err)
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
