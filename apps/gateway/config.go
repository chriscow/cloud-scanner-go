package main

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Name             string `json:"name" default:"Scanner Gateway"`
	Domain           string `json:"domain" default:"http://localhost:3333"`
	Port             int    `json:"port" default:"3333"`
	HealthPath       string `json:"health_path" envconfig:"health_path" default:"/healthz"`
	ReadTimeoutSecs  int    `json:"read_timeout_secs" envconfig:"read_timeout_secs" default:"5"`
	WriteTimeoutSecs int    `json:"write_timeout_secs" envconfig:"write_timeout_secs" default:"10"`
	LogLevel         string `json:"log_level" envconfig:"log_level" default:"error"`
	LogFormatJSON    bool   `json:"log_format_json" envconfig:"log_format_json" default:"false"`
	SessionSecret    string `json:"session_secret" envconfig:"session_secret" default:"mysessionsecret"`

	JWTSecret string `json:"jwt_secret" envconfig:"jwt_secret"`

	Driver     string `json:"driver" envconfig:"driver" default:"sqlite3"`
	DataSource string `json:"datasource" envconfig:"datasource" default:"file:users.db?mode=memory&cache=shared&_fk=1"`

	// goth
	GoogleClientID string `json:"google_client_id" envconfig:"google_client_id"`
	GoogleSecret   string `json:"google_secret" envconfig:"google_secret"`
}

func loadConfig(configFile string, envPrefix string) (config, error) {
	var cfg config
	if err := loadEnvironment(configFile); err != nil {
		return cfg, err
	}

	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		return cfg, err
	}

	if err := verifyEnvironment(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Load(filename)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

func verifyEnvironment() error {
	if os.Getenv("JWT_SECRET") == "" {
		return errors.New("Environment error: JWT_SECRET not set")
	}

	return nil
}
