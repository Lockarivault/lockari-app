package webserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", " X-Authorization", " X-USERID", " X-APP", " X-USER", " X-TRACE-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// OpenTelemetry Middleware
	if cfg.ServiceName != "" {
		engine.Use(otelgin.Middleware(cfg.ServiceName))
	}

	engine.UseH2C = true

	engine.ForwardedByClientIP = true
	engine.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"})

	engine.Use(gin.Logger(), gin.Recovery())

	// Swagger Endpoint
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Metrics
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
