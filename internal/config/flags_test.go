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
			name: "default values",
			args: []string{},
			expectedConfig: map[string]string{
				"FlagServerPort":  ":8080",
				"FlagShortURL":    "http://localhost:8080/",
				"FlagLogLevel":    "debug",
				"FileStoragePath": "",
				"DatabaseDsn":     "user=videos password=password dbname=shortenerdatabase sslmode=disable",
			},
		},
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
		{
			name: "environment variables",
			args: []string{},
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":9090",
				"BASE_URL":          "http://example.com",
				"LOG_LEVEL":         "warn",
				"FILE_STORAGE_PATH": "/env/path",
				"DATABASE_DSN":      "env_dsn",
			},
			expectedConfig: map[string]string{
				"FlagServerPort":  ":9090",
				"FlagShortURL":    "http://example.com/",
				"FlagLogLevel":    "warn",
				"FileStoragePath": "/env/path",
				"DatabaseDsn":     "user=videos password=password dbname=shortenerdatabase sslmode=disable",
			},
		},
		{
			name: "mixed flags and env vars",
			args: []string{"-a=:7070", "-l=error"},
			envVars: map[string]string{
				"BASE_URL": "http://env.url",
			},
			expectedConfig: map[string]string{
				"FlagServerPort":  ":7070",
				"FlagShortURL":    "http://env.url/",
				"FlagLogLevel":    "error",
				"FileStoragePath": "",
				"DatabaseDsn":     "user=videos password=password dbname=shortenerdatabase sslmode=disable",
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
			ParseFlags()

			// Проверяем результаты
			assert.Equal(t, tt.expectedConfig["FlagServerPort"], FlagServerPort)
			assert.Equal(t, tt.expectedConfig["FlagShortURL"], FlagShortURL)
			assert.Equal(t, tt.expectedConfig["FlagLogLevel"], FlagLogLevel)
			assert.Equal(t, tt.expectedConfig["FileStoragePath"], FileStoragePath)
			assert.Equal(t, tt.expectedConfig["DatabaseDsn"], DatabaseDsn)
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

func TestGetEnvAddresses(t *testing.T) {
	tests := []struct {
		name         string
		serverAddr   string
		baseURL      string
		expectedAddr string
		expectedBase string
	}{
		{
			name:         "both set",
			serverAddr:   ":9090",
			baseURL:      "http://test.com",
			expectedAddr: ":9090",
			expectedBase: "http://test.com",
		},
		{
			name:         "only server address",
			serverAddr:   ":9090",
			baseURL:      "",
			expectedAddr: ":9090",
			expectedBase: "",
		},
		{
			name:         "only base URL",
			serverAddr:   "",
			baseURL:      "http://test.com",
			expectedAddr: "",
			expectedBase: "http://test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сохраняем оригинальные env vars
			oldServerAddr := os.Getenv("SERVER_ADDRESS")
			oldBaseURL := os.Getenv("BASE_URL")

			// Восстанавливаем после теста
			defer func() {
				os.Setenv("SERVER_ADDRESS", oldServerAddr)
				os.Setenv("BASE_URL", oldBaseURL)
			}()

			// Устанавливаем env vars
			os.Setenv("SERVER_ADDRESS", tt.serverAddr)
			os.Setenv("BASE_URL", tt.baseURL)

			// Вызываем тестируемую функцию
			serverAddress, baseURL := getEnvAddresses()

			// Проверяем результаты
			assert.Equal(t, tt.expectedAddr, serverAddress)
			assert.Equal(t, tt.expectedBase, baseURL)
		})
	}
}

func TestMain(m *testing.M) {
	// Запускаем тесты
	code := m.Run()

	// Восстанавливаем оригинальные значения глобальных переменных
	FlagServerPort = ":8080"
	FlagShortURL = "http://localhost:8080/"
	FlagLogLevel = "debug"
	FileStoragePath = ""
	DatabaseDsn = "user=videos password=password dbname=shortenerdatabase sslmode=disable"
	IsDatabaseExist = true
	IsFileExist = true

	os.Exit(code)
}
