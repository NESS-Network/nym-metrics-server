package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nymtech/nym-directory/healthcheck"
	"github.com/nymtech/nym-directory/metrics"
	"github.com/nymtech/nym-directory/presence"
	"github.com/nymtech/nym-directory/server/html"
	"github.com/nymtech/nym-directory/server/websocket"

	"github.com/gin-contrib/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// New returns a new REST API server
func New() *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Add cors middleware
	router.Use(cors.Default())

	// Serve Swagger frontend static files using gin-swagger middleware
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Add HTML templates to the router
	t, err := html.LoadTemplate()
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(t)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/server/html/index.html", nil)
	})

	// Set up websocket handlers
	hub := websocket.NewHub()
	go hub.Run()
	router.GET("/ws", func(c *gin.Context) {
		websocket.Serve(hub, c.Writer, c.Request)
	})

	policy := bluemonday.UGCPolicy()

	// Metrics wiring
	sanitizer := metrics.NewSanitizer(policy)
	metricsDb := metrics.NewMetricsDb()
	metricsService := *metrics.NewService(metricsDb, hub)

	metricsCfg := metrics.Config{
		Service:   &metricsService,
		Sanitizer: sanitizer,
	}

	// Presence wiring
	cocoSan := presence.NewCoconodeSanitizer(policy)
	mixSan := presence.NewMixnodeSanitizer(policy)
	providerSan := presence.NewMixproviderSanitizer(policy)

	presenceCfg := presence.Config{
		CocoHostSanitizer:        &cocoSan,
		MixHostSanitizer:         &mixSan,
		MixProviderHostSanitizer: &providerSan,
	}

	// Register all HTTP controller routes
	healthcheck.New().RegisterRoutes(router)
	metrics.New(metricsCfg).RegisterRoutes(router)
	presence.New(presenceCfg).RegisterRoutes(router)

	return router
}
