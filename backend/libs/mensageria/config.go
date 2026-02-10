package mensageria

type QueueConfig struct {
	Name               string            `yaml:"name" json:"name"`
	Durable            bool              `yaml:"durable" json:"durable"`
	AutoDelete         bool              `yaml:"auto_delete" json:"auto_delete"`
	Exclusive          bool              `yaml:"exclusive" json:"exclusive"`
	NoWait             bool              `yaml:"no_wait" json:"no_wait"`
	Args               map[string]any    `yaml:"args" json:"args"`
	DeadLetter         *DeadLetterConfig `yaml:"dead_letter" json:"dead_letter"`
	PrefetchCount      int               `yaml:"prefetch_count" json:"prefetch_count"`
	MaxDeliveryRetries int               `yaml:"max_delivery_retries" json:"max_delivery_retries"`
}

type ExchangeConfig struct {
	Name        string         `yaml:"name" json:"name"`
	Kind        string         `yaml:"kind" json:"kind"`
	Durable     bool           `yaml:"durable" json:"durable"`
	AutoDeleted bool           `yaml:"auto_deleted" json:"auto_deleted"`
	Internal    bool           `yaml:"internal" json:"internal"`
	NoWait      bool           `yaml:"no_wait" json:"no_wait"`
	Args        map[string]any `yaml:"args" json:"args"`
}

type BindingConfig struct {
	QueueName    string         `yaml:"queue" json:"queue"`
	ExchangeName string         `yaml:"exchange" json:"exchange"`
	RoutingKey   string         `yaml:"routing_key" json:"routing_key"`
	NoWait       bool           `yaml:"no_wait" json:"no_wait"`
	Args         map[string]any `yaml:"args" json:"args"`
}

type DeadLetterConfig struct {
	Exchange string `yaml:"exchange" json:"exchange"`
	Key      string `yaml:"routing_key" json:"routing_key"`
}

type FullConfig struct {
	RabbitMQ  Config           `yaml:"config" json:"config"`
	Queues    []QueueConfig    `yaml:"queues" json:"queues"`
	Exchanges []ExchangeConfig `yaml:"exchanges" json:"exchanges"`
	Bindings  []BindingConfig  `yaml:"bindings" json:"bindings"`
}

type Config struct {
	Addr     string `yaml:"addr" json:"addr"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	VHost    string `yaml:"vhost" json:"vhost"`
	Port     string `yaml:"port" json:"port"`
	TLS      bool   `yaml:"tls" json:"tls"`
}
