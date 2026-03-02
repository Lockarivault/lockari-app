package tenanttools

import "context"

const (
	SystemResource = "controlplane"
	ResourceType   = "tenant"
)

// ContextKey é um tipo para chaves de contexto que evita colisões de nomes.
type ContextKey string

const (
	// ContextKeyTenantID é a chave para armazenar o ID do tenant no contexto.
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyUserID é a chave para armazenar o ID do usuário no contexto.
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyActorType é a chave para armazenar o tipo de ator no contexto.
	ContextKeyActorType ContextKey = "actor_type"
	// ContextKeyUserAgent é a chave para armazenar o user agent no contexto.
	ContextKeyUserAgent ContextKey = "user_agent"
	// ContextKeyIPAddress é a chave para armazenar o IP address no contexto.
	ContextKeyIPAddress ContextKey = "ip_address"
	// ContextKeyResource é a chave para armazenar o recurso no contexto.
	ContextKeyResource ContextKey = "resource"
	// ContextKeyTraceID é a chave para armazenar o trace ID no contexto.
	ContextKeyTraceID ContextKey = "trace_id"
	// ContextKeySpanID é a chave para armazenar o span ID no contexto.
	ContextKeySpanID ContextKey = "span_id"
)

func GetStringFromContext(ctx context.Context, key ContextKey) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func SetStringFromContext(ctx context.Context, key ContextKey, value string) context.Context {
	ctxCopy := ctx

	if key == "" || value == "" {
		return ctxCopy
	}

	return context.WithValue(ctxCopy, key, value)
}

func GetActorTypeFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyActorType)
}

func SetActorTypeFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyActorType, value)
}

func GetIPAddressFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyIPAddress)
}

func SetIPAddressFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyIPAddress, value)
}

func GetUserAgentFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyUserAgent)
}

func SetUserAgentFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyUserAgent, value)
}

func GetUserIDFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyUserID)
}

func SetUserIDFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyUserID, value)
}

func GetTenantIDFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyTenantID)
}

func SetTenantIDFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyTenantID, value)
}

func GetTraceIDFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeyTraceID)
}

func SetTraceIDFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeyTraceID, value)
}

func GetSpanIDFromContext(ctx context.Context) string {
	return GetStringFromContext(ctx, ContextKeySpanID)
}

func SetSpanIDFromContext(ctx context.Context, value string) context.Context {
	return SetStringFromContext(ctx, ContextKeySpanID, value)
}
