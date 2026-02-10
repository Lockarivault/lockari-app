package cmd

import (
	"fmt"
	"os"
	"strings"
)

func createModules(name, path *string) error {

	dirs := getDirectories(name)
	if len(dirs) == 0 {
		return fmt.Errorf("directories not found")
	}

	// precisa acessar o path
	// precisa criar os arquivos
	os.Chdir(*path)
	fmt.Printf("the directory path is %s \n", *path)
	for _, dir := range dirs {
		fmt.Printf("the directory is %s \n", dir)
		fmt.Printf("Creating directory: %s/%s \n", *path, dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		err = createFiles(&dir, name)
		if err != nil {
			return err
		}

	}

	return nil
}

func getDirectories(module *string) []string {
	if module == nil || *module == "" {
		return nil
	}

	m := *module
	return []string{
		fmt.Sprintf("internal/core/%s", m),
		fmt.Sprintf("internal/core/%s/model", m),
		fmt.Sprintf("internal/core/%s/repository", m),
		fmt.Sprintf("internal/core/%s/repository/database", m),
		fmt.Sprintf("internal/core/%s/service", m),
		fmt.Sprintf("internal/core/%s/handler", m),
		fmt.Sprintf("internal/core/%s/tools", m),
		fmt.Sprintf("internal/core/%s/usecase", m),
	}
}

func createFiles(path *string, module *string) error {
	return writeFile(path, module)
}

func writeFile(path *string, module *string) error {
	m := *module
	parts := strings.Split(*path, "/")
	dirName := parts[len(parts)-1]

	capitalize := func(s string) string {
		if s == "" {
			return ""
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}

	capM := capitalize(m)

	// Handle Root directory internal/core/%s
	if dirName == m {
		packageName := m + "entity"

		content := fmt.Sprintf(`package %s

import (
	"context"

	handler%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/handler"
	repository%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/repository"
	database%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/repository/database"
	service%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/service"
	tools%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/tools"
	usecase%s "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/%s/usecase"

	"go.uber.org/fx"
)

// provide%sConfig returns the runtime configuration for the %s module.
func provide%sConfig() *tools%s.Config {
	cfg := tools%s.DefaultConfig()
	return &cfg
}

// Module wires the %s components into the Fx container.
var Module = fx.Module("%s",
	fx.Provide(
		provide%sConfig,
		fx.Annotate(
			database%s.NewMongo%sRepository,
			fx.As(new(repository%s.Repository%s)),
		),
		service%s.InnicializeService%s,
		usecase%s.InnicializeUsecase%s,
	),
	fx.Invoke(
		handler%s.NewHandler,
		start%sConsumer,
	),
)

type start%sConsumerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Service   service%s.LifecyclePublisher
}

func start%sConsumer(params start%sConsumerParams) error {
	if params.Service == nil {
		return tools%s.ErrNilService
	}
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return params.Service.StartConsumer(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return params.Service.StopConsumer(ctx)
		},
	})
	return nil
}
`, packageName,
			m, m, // handler
			m, m, // repository
			m, m, // database
			m, m, // service
			m, m, // tools
			m, m, // usecase
			capM, m, // provideConfig comment
			capM, m, // provideConfig func
			m,    // DefaultConfig
			m, m, // Module
			capM,    // provideConfig Provide
			m, capM, // database.NewMongoRepository
			m, capM, // repository.Repository
			m, capM, // service.Innicialize
			m, capM, // usecase.Innicialize
			m,          // handler.NewHandler
			capM,       // startConsumer
			capM,       // startConsumerParams struct
			m,          // Service field
			capM, capM, // func params
			m) // tools error

		return os.WriteFile(fmt.Sprintf("%s/module.go", *path), []byte(content), 0644)
	}

	// Handle Subdirectories
	packageName := dirName + m
	symbolName := capitalize(dirName) + capM

	content := fmt.Sprintf(`package %s

type %s interface {
}

type %s struct {
}

func Innicialize%s() (%s, error) {
	m := %s{}
	return m, nil
}
`, packageName, symbolName, m, symbolName, symbolName, m)

	return os.WriteFile(fmt.Sprintf("%s/module.go", *path), []byte(content), 0644)
}
