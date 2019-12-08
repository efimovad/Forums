package app

type Config struct {
	BindAddr    string
	LogLevel    string
	DatabaseURL string
	SessionKey  string
	TokenSecret string
	ClientUrl	string
}

func NewConfig() *Config {
	return &Config{
		BindAddr:		":5000",
		LogLevel:		"debug",
		SessionKey:		"jdfhdfdj",
		DatabaseURL:	"dbname=docker sslmode=disable port=5432 password=docker user=docker",
		TokenSecret:	"golangsecpark",
	}
}
