package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Postgres struct {
		DSN          string `yaml:"dsn" env:"PG_DSN" env-required:"true"`
		MigrationURL string `yaml:"migration_url" env:"PG_MIGRATION_URL" env-default:"file://migrations"`
	} `yaml:"postgres"`

	HTTP struct {
		PORT string `yaml:"port" env:"PORT" env-default:"8080"`
	} `yaml:"http"`
}

func MustLoad() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("./local.yaml", &cfg); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
