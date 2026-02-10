# Código em Go

## Estrutura de diretórios

O projeto deve seguir a seguinte estrutura de diretórios:

```
backend/services/controlplane/cmd/api
backend/services/controlplane/docs/swagger
backend/services/controlplane/cmd/api/hooks
backend/services/controlplane/cmd/api/providers
backend/services/controlplane/cmd/api/types
backend/services/controlplane/internal/core/<core_name>
backend/services/controlplane/internal/core/<core_name>/model
backend/services/controlplane/internal/core/<core_name>/module.go
backend/services/controlplane/internal/core/<core_name>/repository
backend/services/controlplane/internal/core/<core_name>/usecase
backend/services/controlplane/internal/core/<core_name>/service
backend/services/controlplane/internal/core/<core_name>/tools
backend/services/controlplane/internal/core/<core_name>/handlers
backend/services/controlplane/internal/pkg/<directory_package_name>

```

Todos os códigos devem seguir os melhores padrões de Golang e com bastante enfoque em padrões do Google.

## Padrões

- Segurança em primeiro lugar;
- Performance em segundo lugar;
- Código limpo e organizado;
- Código documentado (comentado, godoc, api com swagger);
- Código testado;
- Código versionado;

## Validações
Muito importante, realizar validações com o Go trace e com o Go pprof para garantir que o código esteja performático.  Para validar a performance, pode usar a variavel de ambiente **GODEBUG=gctrace=1**.

## Performance

Muito importante, validar possíveis utilizações com:
- **channel**
- **sync.WaitGroup**
- **sync.Mutex**
- **sync.RWMutex**
- **goroutines**
- **context**

