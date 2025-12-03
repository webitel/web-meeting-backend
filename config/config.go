package config

type Config struct {
	Service     Service
	Log         LogSettings
	SqlSettings SqlSettings
	Pubsub      Pubsub
}

type Pubsub struct {
	Address string
}

type SqlSettings struct {
	DSN string
}

type Service struct {
	Id        string
	Address   string
	Consul    string
	SecretKey string
}

type LogSettings struct {
	Lvl     string
	Json    bool
	Otel    bool
	File    string
	Console bool
}
