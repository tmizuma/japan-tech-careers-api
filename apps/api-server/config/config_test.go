package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "環境変数が全て設定されている場合",
			envVars: map[string]string{
				"ENVIRONMENT":  "production",
				"LOG_LEVEL":    "debug",
				"API_ENDPOINT": "https://api.production.com",
				"API_TIMEOUT":  "60",
			},
			expected: Config{
				Environment: "production",
				LogLevel:    "debug",
				ApiEndpoint: "https://api.production.com",
				ApiTimeout:  60,
			},
		},
		{
			name:    "環境変数が未設定でデフォルト値が使用される場合",
			envVars: map[string]string{},
			expected: Config{
				Environment: "local",
				LogLevel:    "info",
				ApiEndpoint: "https://api.example.com",
				ApiTimeout:  30,
			},
		},
		{
			name: "一部の環境変数のみ設定されている場合",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"API_TIMEOUT": "45",
			},
			expected: Config{
				Environment: "staging",
				LogLevel:    "info",
				ApiEndpoint: "https://api.example.com",
				ApiTimeout:  45,
			},
		},
		{
			name: "無効なタイムアウト値でデフォルト値が使用される場合",
			envVars: map[string]string{
				"API_TIMEOUT": "invalid",
			},
			expected: Config{
				Environment: "local",
				LogLevel:    "info",
				ApiEndpoint: "https://api.example.com",
				ApiTimeout:  30,
			},
		},
		{
			name: "dev環境の設定",
			envVars: map[string]string{
				"ENVIRONMENT":  "dev",
				"LOG_LEVEL":    "debug",
				"API_ENDPOINT": "https://api.dev.com",
				"API_TIMEOUT":  "15",
			},
			expected: Config{
				Environment: "dev",
				LogLevel:    "debug",
				ApiEndpoint: "https://api.dev.com",
				ApiTimeout:  15,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: 環境変数をクリア
			os.Clearenv()

			// 環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Act: Configを作成
			cfg := NewConfig()

			// Assert: 期待値と一致することを検証
			if *cfg != tt.expected {
				t.Errorf("Config mismatch:\n  expected: %+v\n  got:      %+v", tt.expected, *cfg)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable exists",
			key:          "TEST_KEY",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "Environment variable does not exist",
			key:          "NONEXISTENT_KEY",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Act
			result := getEnv(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "Valid integer",
			key:          "TEST_INT",
			envValue:     "100",
			defaultValue: 50,
			expected:     100,
		},
		{
			name:         "Invalid integer",
			key:          "TEST_INVALID_INT",
			envValue:     "not_a_number",
			defaultValue: 50,
			expected:     50,
		},
		{
			name:         "Environment variable does not exist",
			key:          "NONEXISTENT_INT",
			envValue:     "",
			defaultValue: 50,
			expected:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Act
			result := getEnvAsInt(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
