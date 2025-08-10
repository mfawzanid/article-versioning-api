package config

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PSQLUniqueViolationErrorCode string `envconfig:"PSQL_UNIQUE_VIOLATION_ERROR_CODE" default:"23505"`
	PSQLNotFoundErrorCode        string `envconfig:"PSQL_NOT_FOUND_ERROR_CODE" default:"20000"`
	DatabaseUrl                  string `envconfig:"DATABASE_URL" default:"host=localhost port=5432 user=postgres password=postgres dbname=database sslmode=disable"`
}

var config *Config

func GetConfig() *Config {
	if config != nil {
		return config
	}

	config = &Config{}
	err := envconfig.Process("", config)
	if err != nil {
		log.Fatal(fmt.Errorf("error get config: %s", err.Error()))
	}

	return config
}
