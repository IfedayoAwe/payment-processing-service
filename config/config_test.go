package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_HOST", "test-host")
	os.Setenv("DATABASE_PORT", "5433")
	os.Setenv("DATABASE_NAME", "test_db")
	os.Setenv("DATABASE_USERNAME", "test_user")
	os.Setenv("DATABASE_PASSWORD", "test_pass")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@test-rabbitmq:5672/")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_HOST")
		os.Unsetenv("DATABASE_PORT")
		os.Unsetenv("DATABASE_NAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
		os.Unsetenv("RABBITMQ_URL")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Expected Port to be '9090', got '%s'", cfg.Port)
	}

	if cfg.DatabaseHost != "test-host" {
		t.Errorf("Expected DatabaseHost to be 'test-host', got '%s'", cfg.DatabaseHost)
	}

	if cfg.RabbitMQURL != "amqp://guest:guest@test-rabbitmq:5672/" {
		t.Errorf("Expected RabbitMQURL to be 'amqp://guest:guest@test-rabbitmq:5672/', got '%s'", cfg.RabbitMQURL)
	}

	expectedURL := "postgres://test_user:test_pass@test-host:5433/test_db?sslmode=disable"
	if cfg.DatabaseURL != expectedURL {
		t.Errorf("Expected DatabaseURL to be '%s', got '%s'", expectedURL, cfg.DatabaseURL)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Expected default Port to be '8080', got '%s'", cfg.Port)
	}

	if cfg.DatabaseHost != "localhost" {
		t.Errorf("Expected default DatabaseHost to be 'localhost', got '%s'", cfg.DatabaseHost)
	}
}
