package handlertenant

import (
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/webserver"
	usecasetenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/usecase"
)

type HandlerTenant interface {
}

type tenant struct {
	server  webserver.ServerInterface
	logger  loggers.LoggerInterface
	usecase usecasetenant.UsecaseTenant
}

func NewHandler(
	server webserver.ServerInterface,
	logger loggers.LoggerInterface,
	usecase usecasetenant.UsecaseTenant,
) (HandlerTenant, error) {
	h := &tenant{
		server:  server,
		logger:  logger,
		usecase: usecase,
	}

	// Register routes
	if s, ok := server.(*webserver.Server); ok {
		v1 := s.Engine.Group("/api/v1/tenants")
		{
			v1.POST("", h.Create)
			v1.GET("", h.List)
			v1.GET("/id/:id", h.GetByID)
			v1.GET("/slug/:slug", h.GetBySlug)
			v1.PUT("/:id", h.Update)
			v1.DELETE("/:id", h.Delete)
			v1.GET("/check-slug", h.CheckSlugAvailability)
		}
	}

	return h, nil
}
