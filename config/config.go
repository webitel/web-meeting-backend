package config

type Config struct {
	ServerAddress string
	Redis         RedisConfig
	Consul        Consul
	DatabaseURI   string
}

type RedisConfig struct {
	Address  string
	Login    string
	Password string
	DB       int
}

type Consul struct {
	Address string
}
