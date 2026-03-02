package nosql

import (
	"testing"
)

func TestMongoDBConnect_GetURI(t *testing.T) {
	tests := []struct {
		name     string
		config   MongoDBConnect
		expected string
	}{
		{
			name: "Standard Connection",
			config: MongoDBConnect{
				Host:     "localhost",
				User:     "user",
				Password: "pass",
				Database: "testdb",
				Port:     "27017",
				UseSRV:   false,
			},
			expected: "mongodb://user:pass@localhost:27017/testdb?authSource=admin",
		},
		{
			name: "SRV Connection",
			config: MongoDBConnect{
				Host:     "cluster0.mongodb.net",
				User:     "user",
				Password: "pass",
				Database: "testdb",
				UseSRV:   true,
			},
			expected: "mongodb+srv://user:pass@cluster0.mongodb.net/testdb?retryWrites=true&w=majority",
		},
		{
			name: "Standard with Default Port",
			config: MongoDBConnect{
				Host:     "localhost",
				User:     "user",
				Password: "pass",
				Database: "testdb",
				UseSRV:   false,
			},
			expected: "mongodb://user:pass@localhost:27017/testdb?authSource=admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := tt.config.GetURI()
			if uri != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, uri)
			}
		})
	}
}

func TestMongoDBConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *MongoDBConfig
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: &MongoDBConfig{
				MongoDB: MongoDBConnect{
					Host:     "localhost",
					Database: "testdb",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Host",
			config: &MongoDBConfig{
				MongoDB: MongoDBConnect{
					Database: "testdb",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Database",
			config: &MongoDBConfig{
				MongoDB: MongoDBConnect{
					Host: "localhost",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
