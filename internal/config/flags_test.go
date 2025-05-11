package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		expectedConfig map[string]string
	}{
		{
			name: "command line flags",
			args: []string{"-a=:9090", "-l=info", "-f=/tmp/data", "-d=postgres://user:pass@localhost:5432/db"},
			expectedConfig: map[string]string{
				"FlagServerPort":  ":9090",
				"FlagShortURL":    "http://localhost:8080/",
				"FlagLogLevel":    "info",
				"FileStoragePath": "/tmp/data",
				"DatabaseDsn":     "postgres://user:pass@localhost:5432/db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сохраняем оригинальные аргументы и env vars
			oldArgs := os.Args
			oldEnv := make(map[string]string)
			for k := range tt.envVars {
				oldEnv[k] = os.Getenv(k)
			}

			// Восстанавливаем после теста
			defer func() {
				os.Args = oldArgs
				for k, v := range oldEnv {
					os.Setenv(k, v)
				}
				// Сбрасываем флаги для следующего теста
				flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			}()

			// Устанавливаем env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Устанавливаем аргументы
			os.Args = append([]string{"cmd"}, tt.args...)

			// Вызываем тестируемую функцию
			cfg := ParseFlags()

			// Проверяем результаты
			assert.Equal(t, tt.expectedConfig["FlagServerPort"], cfg.ServerPort)
			assert.Equal(t, tt.expectedConfig["FlagShortURL"], cfg.BaseURL)
			assert.Equal(t, tt.expectedConfig["FlagLogLevel"], cfg.LogLevel)
			assert.Equal(t, tt.expectedConfig["FileStoragePath"], cfg.FileStoragePath)
			assert.Equal(t, tt.expectedConfig["DatabaseDsn"], cfg.DatabaseDsn)
		})
	}
}

func TestNetAddressString(t *testing.T) {
	addr := NetAddress{Host: "localhost", Port: 8080}
	assert.Equal(t, "localhost:8080/", addr.String())
}

func TestNetAddressSet(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    NetAddress
		expectError bool
	}{
		{
			name:        "valid address",
			input:       "http://localhost:8080",
			expected:    NetAddress{Host: "http://localhost", Port: 8080},
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "localhost:8080",
			expectError: true,
		},
		{
			name:        "invalid port",
			input:       "http://localhost:abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := &NetAddress{}
			err := addr.Set(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Host, addr.Host)
				assert.Equal(t, tt.expected.Port, addr.Port)
			}
		})
	}
}
