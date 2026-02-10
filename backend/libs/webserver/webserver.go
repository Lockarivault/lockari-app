package webserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Config struct {
	Port         int
	CertFile     string
	KeyFile      string
	DisableHTTP2 bool
	ServiceName  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type ServerInterface interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type Server struct {
	Engine *gin.Engine
	Config Config
	server *http.Server
}

func New(cfg Config) (*Server, error) {
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("CertFile and KeyFile are mandatory for webserver initialization")
	}

	if cfg.Port == 0 {
		cfg.Port = 443
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 5 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// OpenTelemetry Middleware
	if cfg.ServiceName != "" {
		engine.Use(otelgin.Middleware(cfg.ServiceName))
	}

	// Swagger Endpoint
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &Server{
		Engine: engine,
		Config: cfg,
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Config.Port)

	// TLS 1.3 Enforcement
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		NextProtos: []string{"h2", "http/1.1"}, // Support HTTP/2 over TLS
	}

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.Engine,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		IdleTimeout:  s.Config.IdleTimeout,
		TLSConfig:    tlsConfig,
	}

	return s.server.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
