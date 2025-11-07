package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// AppConfig - app.yml for service const values
type AppConfig struct {
	Service                string  `yaml:"service" env-required:"true" json:"service"`
	Port                   string  `yaml:"port" env-default:"8000" json:"port,omitempty"`
	Host                   string  `yaml:"host" env-required:"true" json:"host"`
	Protocol               string  `yaml:"protocol" env-required:"true" json:"protocol"`
	GrafanaSlowRequest     int32   `yaml:"grafana_slow_request" env-required:"true" json:"grafanaSlowRequest"`
	SentryDSN              string  `yaml:"sentry_dsn" env-required:"true" json:"sentryDSN"`
	SentryEnableTracing    bool    `yaml:"sentry_enable_tracing" json:"sentryEnableTracing"`
	SentryTracesSampleRate float64 `yaml:"sentry_traces_sample_rate" json:"sentryTracesSampleRate"`
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	Enabled      bool
}

type QdrantConfig struct {
	Host     string
	HttpPort int
	GrpcPort int
}

type EmbedderConfig struct {
	Url        string
	Host       string
	Port       int
	Dimensions int
}

type MongoConfig struct {
	Database string
	User     string
	Password string
	Host     string
	Port     int
}

type JaegerConfig struct {
	Host    string
	Port    int
	Enabled bool
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// Config - app.yml + .env for secrets & dev/prod values
type Config struct {
	*AppConfig
	Environment      string
	SecretKeyJWT     string
	CryptoKey        string
	IsDebug          bool
	Mongo            MongoConfig
	Google           GoogleConfig
	Jaeger           JaegerConfig
	Qdrant           QdrantConfig
	Embedder         EmbedderConfig
	Redis            RedisConfig
	SMTP             SMTPConfig
	PyroscopeAddress string
	FrontendUrl      string
	JsonLog          bool
	EnabledPyroscope bool
	EnabledSentry    bool
	OpenAIKey        string
}

const configPath = "app.yml"

var (
	instance *Config
	once     sync.Once
)

func GetConfig(rootPath string) *Config {
	once.Do(func() {
		if rootPath != "" {
			err := os.Chdir(rootPath)
			if err != nil {
				log.Fatalf("[GetConfig] Chdir to %v error: %v]", rootPath, err)
			}
		}

		instance = &Config{}

		// region app.yml
		instanceApp := &AppConfig{}
		if err := cleanenv.ReadConfig(configPath, instanceApp); err != nil {
			help, _ := cleanenv.GetDescription(instanceApp, nil)
			log.Fatalf("[GetConfig] cleanenv: {%s}, {%s}", err, help)
		}
		instance.AppConfig = instanceApp
		// endregion

		// region .env
		if err := godotenv.Load(); err != nil {
			log.Println("[GetConfig] No .env file")
		}
		instance.Environment = getEnvKey("ENVIRONMENT")
		instance.IsDebug = instance.Environment != "production"

		instance.SecretKeyJWT = getEnvKey("JWT_SECRET_KEY")
		instance.CryptoKey = getEnvKey("CRYPTO_SECRET_KEY")

		instance.Google.Enabled = strings.ToLower(getEnvKey("ENABLED_GOOGLE_AUTH")) == "true"
		if instance.Google.Enabled {
			instance.Google.ClientID = getEnvKey("GOOGLE_CLIENT_ID")
			instance.Google.ClientSecret = getEnvKey("GOOGLE_CLIENT_SECRET")
		}

		instance.Jaeger.Enabled = strings.ToLower(getEnvKey("ENABLED_JAEGER")) == "true"
		if instance.Jaeger.Enabled {
			instance.Jaeger.Host = getEnvKey("JAEGER_HOST")
			instance.Jaeger.Port = getEnvKeyInt("JAEGER_PORT")
		}

		instance.Mongo.Database = getEnvKey("MONGO_DATABASE")
		instance.Mongo.Password = getEnvKey("MONGO_PASSWORD")
		instance.Mongo.Host = getEnvKey("MONGO_HOST_NAME")
		instance.Mongo.User = getEnvKey("MONGO_USER")
		instance.Mongo.Port = getEnvKeyInt("MONGO_PORT")

		instance.EnabledPyroscope = strings.ToLower(getEnvKey("ENABLED_PYROSCOPE")) == "true"
		if instance.EnabledPyroscope {
			instance.PyroscopeAddress = getEnvKey("PYROSCOPE_ADDRESS")
		}

		instance.Qdrant.Host = getEnvKey("QDRANT_HOST")
		instance.Qdrant.GrpcPort = getEnvKeyInt("QDRANT_GRPC_PORT")
		instance.Qdrant.HttpPort = getEnvKeyInt("QDRANT_HTTP_PORT")

		instance.Redis.Host = getEnvKey("REDIS_HOST")
		instance.Redis.Password = getEnvKey("REDIS_PASSWORD")
		instance.Redis.Port = getEnvKeyInt("REDIS_PORT")

		instance.Embedder.Host = getEnvKey("EMBEDDER_HOST")
		instance.Embedder.Port = getEnvKeyInt("EMBEDDER_PORT")
		instance.Embedder.Url = fmt.Sprintf("http://%s:%d", instance.Embedder.Host, instance.Embedder.Port)
		instance.Embedder.Dimensions = getEnvKeyInt("EMBEDDER_DIMENSIONS")

		instance.FrontendUrl = getEnvKey("FRONTEND_URL")
		instance.JsonLog = getEnvKey("JSON_LOG") == "true"

		instance.Embedder.Host = getEnvKey("EMBEDDER_HOST")
		instance.Embedder.Port = getEnvKeyInt("EMBEDDER_PORT")

		instance.SMTP.Host = getEnvKey("SMTP_HOST")
		instance.SMTP.Port = getEnvKeyInt("SMTP_PORT")
		instance.SMTP.Username = getEnvKey("SMTP_USERNAME")
		instance.SMTP.Password = getEnvKey("SMTP_PASSWORD")
		instance.SMTP.From = getEnvKey("SMTP_FROM")
		instance.SMTP.FromName = getEnvKey("SMTP_FROM_NAME")

		instance.OpenAIKey = getEnvKey("OPEN_AI_KEY")
		// endregion
	})
	return instance
}

func getEnvKey(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("[getEnvKey] no value for %s", key)
	return ""
}

func getEnvKeyInt(key string) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("[getEnvKeyInt] no value for %s", key)
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("[getEnvKeyInt] cannot convert to int: %s", value)
	}

	return num
}
