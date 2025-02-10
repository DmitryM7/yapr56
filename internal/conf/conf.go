package conf

import (
	"flag"
	"os"
)

type Config struct {
	BndAdr string
	DSN    string
}

func (s *Config) ParseFlags() {
	flag.StringVar(&s.BndAdr, "a", "localhost:8080", "host where server is run")
	flag.StringVar(&s.DSN, "d", "", "database dsn")
}

func (s *Config) ParseEnv() {
	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		s.BndAdr = env
	}

	if env := os.Getenv("DATABASE_DSN"); env != "" {
		s.DSN = env
	}
}

func NewConf() Config {
	c := Config{}
	c.ParseFlags()
	c.ParseEnv()
	return c
}
