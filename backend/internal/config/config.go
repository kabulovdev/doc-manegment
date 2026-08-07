package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env            string        `env:"APP_ENV" envDefault:"dev"`
	HTTPAddr       string        `env:"HTTP_ADDR" envDefault:":8080"`
	MongoURI       string        `env:"MONGO_URI,required"`
	MongoDB        string        `env:"MONGO_DB" envDefault:"docmgmt"`
	JWTSecret      string        `env:"JWT_SECRET,required"`
	MasterKeyB64   string        `env:"MASTER_KEY,required"`
	FrontendOrigin string        `env:"FRONTEND_ORIGIN" envDefault:"http://localhost:3001"`
	CookieDomain   string        `env:"COOKIE_DOMAIN" envDefault:""`
	CookieSecure   bool          `env:"COOKIE_SECURE" envDefault:"false"`
	AccessTTL      time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL     time.Duration `env:"REFRESH_TTL" envDefault:"720h"`
	MaxFileSize    int64         `env:"MAX_FILE_SIZE_BYTES" envDefault:"10737418240"`
	AllowHTTP      bool          `env:"ALLOW_HTTP_ENDPOINTS" envDefault:"false"`
	QdrantURL      string        `env:"QDRANT_URL" envDefault:""`
	QdrantAPIKey   string        `env:"QDRANT_API_KEY" envDefault:""`

	MasterKey []byte `env:"-"`
}

func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(c.MasterKeyB64)
	if err != nil {
		return nil, fmt.Errorf("MASTER_KEY is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("MASTER_KEY must decode to exactly 32 bytes")
	}
	if len(c.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	c.MasterKey = key
	return &c, nil
}
