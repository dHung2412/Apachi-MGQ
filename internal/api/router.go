package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"

	"DP_Maintenance/internal/api/handlers"
	custommw "DP_Maintenance/internal/api/middleware"
	"DP_Maintenance/internal/service"
)

// Router wires all HTTP handlers and middleware together.
type Router struct {
	e              *echo.Echo
	authHandler    *handlers.AuthHandler
	ingestHandler  *handlers.IngestHandler
	datasetHandler *handlers.DatasetHandler
	lineageHandler *handlers.LineageHandler
	healthHandler  *handlers.HealthHandler
	jwtMiddleware  *custommw.JWTMiddleware
}

// NewRouter creates a new Router with all handlers and services wired.
func NewRouter(
	authHandler *handlers.AuthHandler,
	ingestHandler *handlers.IngestHandler,
	datasetHandler *handlers.DatasetHandler,
	lineageHandler *handlers.LineageHandler,
	healthHandler *handlers.HealthHandler,
	authService *service.AuthService,
) *Router {
	return &Router{
		e:              echo.New(),
		authHandler:    authHandler,
		ingestHandler:  ingestHandler,
		datasetHandler: datasetHandler,
		lineageHandler: lineageHandler,
		healthHandler:  healthHandler,
		jwtMiddleware:  custommw.NewJWTMiddleware(authService),
	}
}

// Setup configures all routes and middleware, returns the Echo instance.
func (r *Router) Setup() *echo.Echo {
	r.e.HTTPErrorHandler = customErrorHandler

	// Global middleware
	r.e.Use(emw.Logger())
	r.e.Use(emw.Recover())
	r.e.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// --- Public routes ---
	r.e.GET("/health", r.healthHandler.Health)

	authGroup := r.e.Group("/auth")
	authGroup.POST("/login", r.authHandler.Login)
	authGroup.POST("/register", r.authHandler.Register)

	// --- Protected routes (JWT required) ---
	apiGroup := r.e.Group("/api/v1")
	apiGroup.Use(r.jwtMiddleware.Authenticate)

	// Ingestion
	apiGroup.POST("/ingest", r.ingestHandler.Ingest)
	apiGroup.GET("/ingest/events", r.ingestHandler.ListEvents)

	// Datasets
	apiGroup.GET("/datasets", r.datasetHandler.List)
	apiGroup.GET("/datasets/:urn", r.datasetHandler.Get)
	apiGroup.POST("/datasets", r.datasetHandler.Create)
	apiGroup.PUT("/datasets/:urn", r.datasetHandler.Update)
	apiGroup.DELETE("/datasets/:urn", r.datasetHandler.Delete)

	// Lineage
	apiGroup.GET("/datasets/:urn/lineage", r.lineageHandler.GetLineage)
	apiGroup.GET("/lineage/trace", r.lineageHandler.Trace)

	return r.e
}

// customErrorHandler provides consistent JSON error responses.
func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.JSON(code, map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	})
}