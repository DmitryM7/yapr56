package conf

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const defaultSecretKeyTime = 25

type Config struct {
	BndAdr        string
	DSN           string
	SecretKey     string
	SecretKeyTime time.Duration
}

func (s *Config) ParseFlags() {
	flag.StringVar(&s.BndAdr, "a", "localhost:8080", "host where server is run")
	flag.StringVar(&s.DSN, "d", "", "database dsn")
	flag.StringVar(&s.SecretKey, "k", "", "Secret key for JWT")
	flag.DurationVar(&s.SecretKeyTime, "kt", 30*time.Minute, "Time secret key in minutes")
}

func (s *Config) ParseEnv() {
	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		s.BndAdr = env
	}

	if env := os.Getenv("DATABASE_DSN"); env != "" {
		s.DSN = env
	}

	if env := os.Getenv("SECRET_KEY"); env != "" {
		s.SecretKey = env
	}

	if env := os.Getenv("SECRET_KEY_TIME"); env != "" {
		duration, err := strconv.Atoi(env)

		if err == nil {
			s.SecretKeyTime = time.Duration(duration) * time.Minute
		} else {
			s.SecretKeyTime = time.Duration(defaultSecretKeyTime) * time.Minute
		}
	}
}

func NewConf() Config {
	c := Config{}
	c.ParseFlags()
	flag.Parse()
	c.ParseEnv()
	return c
}
