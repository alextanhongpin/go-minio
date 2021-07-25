package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     int64  `envconfig:"PORT" required:"true"`
	Name     string `envconfig:"NAME" required:"true"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASS" required:"true"`
	SSLMode  string `envconfig:"SSLMODE" default:"disable"`
}

func (c DBConfig) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func NewDBConfig() DBConfig {
	var cfg DBConfig
	if err := envconfig.Process("db", &cfg); err != nil {
		log.Fatalln(err)
	}
	return cfg
}

func NewDB(cfg DBConfig) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.String())
	if err != nil {
		log.Fatalln(err)
	}
	return db
}
