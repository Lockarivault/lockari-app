package auditlog

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// AuditEntry representa um registro único de auditoria no sistema.
type AuditEntry struct {
	TenantID     string         `json:"tenant_id"`     // ID do cliente (SaaS)
	UserID       string         `json:"user_id"`       // ID de quem fez a ação (UUID ou AppID)
	ActorType    UserType       `json:"actor_type"`    // Tipo de ator (human, app, system)
	ResourceType string         `json:"resource_type"` // ex: "secret"
	ResourceID   string         `json:"resource_id"`   // ID do recurso
	Action       Action         `json:"action"`        // ex: "read", "update"
	Timestamp    time.Time      `json:"timestamp"`     // Gerado automaticamente se vazio
	IPAddress    string         `json:"ip_address"`    // IP do solicitante
	UserAgent    string         `json:"user_agent"`    // Browser/Client info
	Metadata     map[string]any `json:"metadata"`      // Dados extras
}

// AuditStore define o contrato para persistência de logs de auditoria.
// Isso permite que a biblioteca seja agnóstica a banco de dados.
type AuditStore interface {
	Save(ctx context.Context, entry AuditEntry) error
}

// Service define o contrato público da biblioteca de auditoria.
// Retornar uma interface facilita o uso de Mocks em testes unitários.
type Service interface {
	CreateAuditLog(ctx context.Context, entry AuditEntry) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Config contém as configurações para inicializar o serviço de auditoria.
type Config struct {
	Store      AuditStore
	Workers    int
	BufferSize int
}

// auditService é a implementação concreta do serviço de auditoria.
type auditService struct {
	store      AuditStore
	logChan    chan AuditEntry
	workers    int
	bufferSize int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// New cria uma nova instância do serviço de auditoria retornando a interface Service.
func New(cfg Config) Service {
	// Se workers não for informado, usamos um cálculo baseado em CPU
	// Limitamos a um mínimo de 1 e um máximo de 50 para evitar excessos.
	if cfg.Workers <= 0 {
		cpus := runtime.NumCPU()
		cfg.Workers = cpus * 7
		if cfg.Workers > 50 {
			cfg.Workers = 50
		}
	}
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}

	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &auditService{
		store:      cfg.Store,
		logChan:    make(chan AuditEntry, cfg.BufferSize),
		workers:    cfg.Workers,
		bufferSize: cfg.BufferSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start inicia o Worker Pool de auditoria.
func (a *auditService) Start(ctx context.Context) error {
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker()
	}
	slog.Info("auditlog service started", "workers", a.workers, "buffer", a.bufferSize)
	return nil
}

// Stop realiza o shutdown gracioso da biblioteca.
func (a *auditService) Stop(ctx context.Context) error {
	a.cancel()
	close(a.logChan)
	a.wg.Wait()
	slog.Info("auditlog service stopped gracefully")
	return nil
}

// CreateAuditLog enfileira um log de auditoria de forma não-bloqueante.
func (a *auditService) CreateAuditLog(ctx context.Context, entry AuditEntry) error {
	// 1. Fail Fast: Validação básica
	if entry.Action == "" || entry.ResourceType == "" || entry.TenantID == "" || entry.UserID == "" || entry.ActorType == "" {
		return errors.New("auditlog: missing required fields (Action, ResourceType, TenantID, UserID, ActorType)")
	}

	// 2. Enriquecimento
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// 3. Envio Não-bloqueante
	select {
	case a.logChan <- entry:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Se o buffer estiver cheio, logamos o erro para monitoramento
		slog.Error("auditlog buffer is full, dropping log entry",
			"tenant_id", entry.TenantID,
			"action", entry.Action,
			"resource", entry.ResourceType)
		return errors.New("auditlog: buffer is full, dropping log")
	}
}

// worker consome os logs e chama o Store.
func (a *auditService) worker() {
	defer a.wg.Done()

	for entry := range a.logChan {
		// Timeout individual para cada operação de salvamento
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		if a.store != nil {
			if err := a.store.Save(saveCtx, entry); err != nil {
				// Registro de erro estruturado caso o banco/mensageria falhe
				slog.Error("failed to save audit log",
					"error", err,
					"tenant_id", entry.TenantID,
					"action", entry.Action)
			}
		}

		cancel()
	}
}
