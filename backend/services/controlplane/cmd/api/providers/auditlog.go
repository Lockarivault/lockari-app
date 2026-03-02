package providers

import (
	"context"

	"github.com/lockarivault/lockari-app/backend/libs/auditlog"
	"github.com/lockarivault/lockari-app/backend/libs/database/nosql"
	"go.uber.org/fx"
)

// auditStoreWrapper adapta o nosql.DatabaseService para a interface auditlog.AuditStore.
type auditStoreWrapper struct {
	db nosql.DatabaseService
}

// Save implementa a persistência no MongoDB para a biblioteca de auditoria.
func (w *auditStoreWrapper) Save(ctx context.Context, entry auditlog.AuditEntry) error {
	collection, err := w.db.Collection("audit_logs")
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, entry)
	return err
}

// ProvideAuditlog inicializa a lib de auditoria com workers e canal interno.
func ProvideAuditlog(lc fx.Lifecycle, db nosql.DatabaseService) auditlog.Service {
	// 1. Criamos a instância da lib usando o wrapper para satisfazer a interface Store
	service := auditlog.New(auditlog.Config{
		Store:      &auditStoreWrapper{db: db},
		Workers:    5,    // 5 goroutines processando em paralelo
		BufferSize: 1000, // Fila de espera de até 1000 logs
	})

	// 2. Integramos ao Lifecycle do Go
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Aqui iniciamos o Worker Pool
			return service.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			// Shutdown Gracioso: espera processar o que sobrou no canal
			return service.Stop(ctx)
		},
	})

	return service
}
