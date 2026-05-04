package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"mydeploy-project/internal/api/handler"
	"mydeploy-project/internal/api/router"
	"mydeploy-project/internal/config"
	"mydeploy-project/internal/middleware"
	"mydeploy-project/internal/service"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		executablePath, execErr := os.Executable()
		if execErr != nil {
			log.Fatalf("resolve app base directory failed: getwd=%v executable=%v", err, execErr)
		}
		baseDir = filepath.Dir(executablePath)
	}

	configPathFlag := flag.String("config", "", "application config file path")
	addrFlag := flag.String("addr", "", "server listen address, for example 0.0.0.0:8080")
	frontendDistFlag := flag.String("frontend-dist", "", "frontend dist directory path")
	flag.Parse()

	appConfig, err := config.Load(config.LoadOptions{
		BaseDir:                 baseDir,
		ConfigPath:              *configPathFlag,
		OverrideAddress:         *addrFlag,
		OverrideFrontendDistDir: *frontendDistFlag,
	})
	if err != nil {
		log.Fatalf("load app config failed: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.ErrorHandler())

	connectionManager := service.NewConnectionManager()
	connectionHandler := handler.NewConnectionHandler(connectionManager)
	metadataHandler := handler.NewMetadataHandler(service.NewMetadataService())
	dataHandler := handler.NewDataHandler(service.NewDataService())
	queryHandler := handler.NewQueryHandler(service.NewQueryService())
	securityHandler := handler.NewSecurityHandler(service.NewSecurityService())
	backupHandler := handler.NewBackupHandler(service.NewBackupService(baseDir))
	schemaCompareHandler := handler.NewSchemaCompareHandler(service.NewSchemaCompareService())

	router.Register(engine, connectionHandler, connectionManager, metadataHandler, dataHandler, queryHandler, securityHandler, backupHandler, schemaCompareHandler)
	registerFrontend(engine, appConfig.Frontend.DistDir)

	log.Printf("server started on %s with frontend dist %s", appConfig.Server.Address, appConfig.Frontend.DistDir)
	if err := engine.Run(appConfig.Server.Address); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}

func registerFrontend(engine *gin.Engine, distDir string) {
	indexPath := filepath.Join(distDir, "index.html")
	appConfigPath := filepath.Join(distDir, "app-config.json")

	if _, err := os.Stat(indexPath); err != nil {
		log.Printf("frontend dist not found at %s, only API routes will be available", indexPath)
		return
	}

	engine.StaticFS("/assets", http.Dir(filepath.Join(distDir, "assets")))
	faviconPath := filepath.Join(distDir, "favicon.ico")
	serveFavicon := func(ctx *gin.Context) {
		disableFrontendCache(ctx)
		if _, err := os.Stat(faviconPath); err == nil {
			ctx.File(faviconPath)
			return
		}

		ctx.Status(http.StatusNoContent)
	}
	engine.GET("/favicon.ico", serveFavicon)
	engine.HEAD("/favicon.ico", serveFavicon)
	if _, err := os.Stat(appConfigPath); err == nil {
		serveAppConfig := func(ctx *gin.Context) {
			disableFrontendCache(ctx)
			ctx.File(appConfigPath)
		}
		engine.GET("/app-config.json", serveAppConfig)
		engine.HEAD("/app-config.json", serveAppConfig)
	}
	engine.NoRoute(func(ctx *gin.Context) {
		method := ctx.Request.Method
		if (method != http.MethodGet && method != http.MethodHead) || ctx.Request.URL.Path == "/api" || strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "resource not found",
			})
			return
		}

		disableFrontendCache(ctx)
		ctx.File(indexPath)
	})
}

func disableFrontendCache(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")
	ctx.Header("Surrogate-Control", "no-store")
}
