package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Timeout struct{}

type FileConfig struct {
	Extentsion     string
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

type AppConfigField struct {
	Fields map[string]any
	Keys   []string
}

type AppConfig map[string]interface{}

type Connections struct {
	PathConfigFile string `mapstructure:"path_config_file"`
	FileConfig     *FileConfig
	Fields         map[string]interface{}
	Redis          RedisConfig         `mapstructure:"redis"` // Added Redis config field
	MessageQueues  MessageQueuesConfig `mapstructure:"messagequeues"`
	Vault          VaultConfig         `mapstructure:"vault"`
	Quotas         QuotasConfig        `mapstructure:"quotas"`
	App            AppDomainConfig     `mapstructure:"app"`
	Settings       AppConfig
	*AppConfigField
}

type AppDomainConfig struct {
	BaseDomain string `mapstructure:"base_domain"`
}

type VaultConfig struct {
	RootKey   string            `mapstructure:"root_key"`
	RootKeyID string            `mapstructure:"root_key_id"`
	RootKeys  map[string]string `mapstructure:"root_keys"` // Support for multiple root keys (Rotation)
}

type QuotasConfig struct {
	MaxSecrets      int   `mapstructure:"max_secrets"`
	MaxUsers        int   `mapstructure:"max_users"`
	MaxStorageBytes int64 `mapstructure:"max_storage_bytes"`
}

// RedisConfig holds the configuration for Redis connection.
type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MessageQueuesConfig struct {
	Addr     string            `mapstructure:"addr"`
	User     string            `mapstructure:"user"`
	Password string            `mapstructure:"password"`
	VHost    string            `mapstructure:"vhost"`
	Port     string            `mapstructure:"port"`
	TLS      bool              `mapstructure:"tls"`
	Data     MessageQueuesData `mapstructure:"config"`
}

type MessageQueuesData struct {
	Exchanges []MQExchangeConfig `mapstructure:"exchanges"`
	Queues    []MQQueueConfig    `mapstructure:"queues"`
	Bindings  []MQBindingConfig  `mapstructure:"bindings"`
}

type MQExchangeConfig struct {
	Name        string         `mapstructure:"name"`
	Kind        string         `mapstructure:"kind"`
	Durable     bool           `mapstructure:"durable"`
	AutoDeleted bool           `mapstructure:"auto_deleted"`
	Internal    bool           `mapstructure:"internal"`
	NoWait      bool           `mapstructure:"no_wait"`
	Args        map[string]any `mapstructure:"args"`
}

type MQQueueConfig struct {
	Name               string              `mapstructure:"name"`
	Durable            bool                `mapstructure:"durable"`
	AutoDelete         bool                `mapstructure:"auto_delete"`
	Exclusive          bool                `mapstructure:"exclusive"`
	NoWait             bool                `mapstructure:"no_wait"`
	Args               map[string]any      `mapstructure:"args"`
	DeadLetter         *MQDeadLetterConfig `mapstructure:"dead_letter"`
	PrefetchCount      int                 `mapstructure:"prefetch_count"`
	MaxDeliveryRetries int                 `mapstructure:"max_delivery_retries"`
}

type MQBindingConfig struct {
	Queue      string         `mapstructure:"queue"`
	Exchange   string         `mapstructure:"exchange"`
	RoutingKey string         `mapstructure:"routing_key"`
	NoWait     bool           `mapstructure:"no_wait"`
	Args       map[string]any `mapstructure:"args"`
}

type MQDeadLetterConfig struct {
	Exchange string `mapstructure:"exchange"`
	Key      string `mapstructure:"routing_key"`
}

func LoadConfig() (*Connections, error) {
	var appConfig string

	if os.Getenv("PATH_CONFIG") != "" {
		appConfig = os.Getenv("PATH_CONFIG")
	}

	if appConfig == "" {
		appConfig = "./config.yaml"
	}

	// Basic validation of file existence
	if _, err := os.Stat(appConfig); err != nil {
		return nil, err
	}
	log.Println("App config path: ", appConfig)

	fc := FileConfig{
		Extentsion:     "yaml",
		FileName:       "config",
		ConfigFilePath: appConfig,
	}

	// Initialize the target struct
	cfg := &Connections{
		PathConfigFile: fc.ConfigFilePath,
		FileConfig:     &fc,
	}

	// Use SetConfigFile for fixed path
	viper.SetConfigFile(appConfig)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// Unmarshal into the initialized struct
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// Populate dynamic fields
	cfg.Fields = viper.AllSettings()
	cfg.AppConfigField = &AppConfigField{
		Fields: viper.AllSettings(),
		Keys:   viper.AllKeys(),
	}
	cfg.Settings = viper.AllSettings()

	log.Println("Config file loaded successfully")
	return cfg, nil
}
