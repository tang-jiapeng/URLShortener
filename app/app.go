package app

import (
	"URLShortener/cache"
	"URLShortener/config"
	"URLShortener/controller"
	"URLShortener/db"
	"URLShortener/pkg/shortcode"
	"URLShortener/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	router     *gin.Engine
	server     *http.Server
	cfg        *config.Config
	db         *sql.DB
	cache      cache.Cache
	generator  shortcode.Generator
	urlService service.URLService
	urlHandler *controller.URLHandler
}

func NewApplication() (*Application, error) {
	a := &Application{}

	if err := a.loadConfig(); err != nil {
		return nil, fmt.Errorf("loadConfig: %v", err)
	}
	if err := a.initDB(); err != nil {
		return nil, fmt.Errorf("initDB: %v", err)
	}
	if err := a.initCache(); err != nil {
		return nil, fmt.Errorf("initCache: %v", err)
	}

	a.initGenerator()
	a.initHandler()
	a.SetupRouter()

	return a, nil
}

func (a *Application) Run() {
	log.Printf("Run: ...")
	go a.cleanup()
	go a.startServer()
	a.shutdown()
}

func (a *Application) loadConfig() error {
	cfg, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		panic(err)
	}
	a.cfg = cfg
	return nil
}

func (a *Application) initDB() error {
	sqlDB, err := db.InitDB(a.cfg.MySQL)
	if err != nil {
		return err
	}
	a.db = sqlDB
	return nil
}

func (a *Application) initCache() error {
	redisCache, err := cache.NewRedisCache(a.cfg.Redis)
	if err != nil {
		return err
	}
	a.cache = redisCache
	return nil
}

func (a *Application) initGenerator() {
	a.generator = shortcode.NewShortCodeGenerator(a.cfg.ShortCode.MinLength)
}

func (a *Application) initHandler() {
	a.urlService = service.NewUrlService(a.db, a.cache, a.generator, a.cfg)
	a.urlHandler = controller.NewURLHandler(a.urlService, a.cfg.App.BaseURL)
}

func (a *Application) SetupRouter() {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("api/url", a.urlHandler.CreateURL)
	router.GET("/:code", a.urlHandler.RedirectURL)

	a.router = router
}

func (a *Application) cleanup() {
	ticker := time.NewTicker(a.cfg.App.CleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := a.urlService.Cleanup(context.Background()); err != nil {
			log.Printf("Failed to clean expired URLs: %v", err)
		}
	}
}

func (a *Application) startServer() {
	a.server = &http.Server{
		Addr:    a.cfg.Server.Address,
		Handler: a.router,
	}
	log.Printf("Starting Server on %s", a.cfg.Server.Address)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func (a *Application) shutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	defer a.db.Close()
	defer a.cache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited!")
}
