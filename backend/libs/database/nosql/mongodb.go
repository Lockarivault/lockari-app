package nosql

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseService interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	GetType() string
	GetConnection() (*mongo.Database, error)
	Collection(collection string) (*mongo.Collection, error)
}

type MongoDBConfig struct {
	Kind    string         `json:"kind" yaml:"kind"`
	MongoDB MongoDBConnect `json:"config" yaml:"config"`
}

func (m *MongoDBConfig) Validate() error {
	if m == nil {
		return fmt.Errorf("config is nil")
	}
	return m.MongoDB.Validate()
}

type MongoDBConnect struct {
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Database string `json:"database" yaml:"database"`
	Password string `json:"password" yaml:"password"`
	Port     string `json:"port" yaml:"port"`
	UseSRV   bool   `json:"use_srv" yaml:"use_srv"`
}

func (m *MongoDBConnect) GetURI() string {
	if m == nil {
		panic("MongoDBConnect is nil")
	}

	if m.UseSRV {
		return fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
			m.User, m.Password, m.Host, m.Database)
	}

	port := m.Port
	if port == "" {
		port = "27017"
	}

	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		m.User, m.Password, m.Host, port, m.Database)
}

func (m *MongoDBConnect) Validate() error {
	if m == nil {
		return fmt.Errorf("connect config is nil")
	}

	if m.Host == "" {
		return fmt.Errorf("host is required")
	}

	if m.Database == "" {
		return fmt.Errorf("database is required")
	}

	// For standard connections, Port is highly recommended but we can default it.
	// For SRV, Port is not used.

	return nil
}

// MongoDB implementation
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	uri      string
	dbName   string
}

// NewMongoDB initializes the MongoDB struct without connecting.
func NewMongoDB(uri, dbName string) (DatabaseService, error) {
	if uri == "" {
		return nil, fmt.Errorf("uri is required")
	}
	if dbName == "" {
		return nil, fmt.Errorf("dbName is required")
	}

	return &MongoDB{
		uri:    uri,
		dbName: dbName,
	}, nil
}

func (m *MongoDB) Connect(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("MongoDB instance is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(m.uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetConnectTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.dbName)
	return nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("MongoDB instance is nil")
	}

	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

func (m *MongoDB) GetType() string {
	return "mongodb"
}

func (m *MongoDB) GetConnection() (*mongo.Database, error) {
	if m == nil {
		return nil, fmt.Errorf("MongoDB instance is nil")
	}
	if m.database == nil {
		return nil, fmt.Errorf("MongoDB database connection is nil - call Connect() first")
	}
	return m.database, nil
}

func (m *MongoDB) Collection(collection string) (*mongo.Collection, error) {
	if m == nil {
		return nil, fmt.Errorf("MongoDB instance is nil")
	}
	if m.database == nil {
		return nil, fmt.Errorf("MongoDB database connection is nil - call Connect() first")
	}
	return m.database.Collection(collection), nil
}
