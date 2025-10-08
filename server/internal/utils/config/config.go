package config

import (
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
	Port                   string  `yaml:"port" env-default:"8080" json:"port,omitempty"`
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

type MongoConfig struct {
	Database string
	User     string
	Password string
	HostName string
	Port     int
}

type JaegerConfig struct {
	Host    string
	Port    int
	Enabled bool
}

// Config - app.yml + .env for secrets & dev/prod values
type Config struct {
	*AppConfig
	Environment            string
	SecretKeyJWT           string
	CryptoKey              string
	IsDebug                bool
	Mongo                  MongoConfig
	Google                 GoogleConfig
	Jaeger                 JaegerConfig
	PyroscopeServerAddress string
	LokiURL                string
	FrontendUrl            string
	JsonLog                bool
	EnabledPyroscope       bool
	EnabledSentry          bool
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
			num, err := strconv.Atoi(getEnvKey("JAEGER_PORT"))
			if err != nil {
				log.Print("[GetConfig] JAEGER_PORT env variable not set or not an int")
			}
			instance.Jaeger.Port = num
		}

		instance.Mongo.Database = getEnvKey("MONGO_DATABASE")
		instance.Mongo.Password = getEnvKey("MONGO_PASSWORD")
		instance.Mongo.HostName = getEnvKey("MONGO_HOST_NAME")
		instance.Mongo.User = getEnvKey("MONGO_USER")
		num, err := strconv.Atoi(getEnvKey("MONGO_PORT"))
		if err != nil {
			log.Fatal("[GetConfig] MONGO_PORT env variable not set or not an int")
		}
		instance.Mongo.Port = num

		instance.EnabledPyroscope = strings.ToLower(getEnvKey("ENABLED_PYROSCOPE")) == "true"
		if instance.EnabledPyroscope {
			instance.PyroscopeServerAddress = getEnvKey("PYROSCOPE_SERVER_ADDRESS")
		}

		instance.FrontendUrl = getEnvKey("FRONTEND_URL")
		instance.JsonLog = getEnvKey("JSON_LOG") == "true"

		// endregion
	})
	return instance
}

func getEnvKey(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("[getEnvKey] no value for " + key)
	return ""
}
