# Web Meeting Backend

Backend service for Webitel Web Meetings. Built with Go, using gRPC, Consul, RabbitMQ, and PostgreSQL.

## Configuration

The service is configured via environment variables or command-line flags.

| Environment Variable | Flag | Description | Default |
|----------------------|------|-------------|---------|
| `ID` | `--service-id`, `-i` | Unique service identifier | `1` |
| `BIND_ADDRESS` | `--bind-address`, `-b` | Address for internal cluster communication | `localhost:50011` |
| `CONSUL` | `--consul-discovery`, `-c` | Consul service discovery address | `127.0.0.1:8500` |
| `DATA_SOURCE` | `--postgresql-dsn` | PostgreSQL connection string | *Required* |
| `PUBSUB` | `--pubsub`, `-p` | RabbitMQ connection string | `amqp://admin:admin@127.0.0.1:5672/` |
| `SECRET_KEY` | `--data-encrypter` | Secret key for data encryption | `MY_SECRET_KEY` |
| `LOG_LVL` | `--log-level`, `-l` | Logging level (debug, info, error) | `debug` |
| `LOG_JSON` | `--log-json` | Enable JSON logging format | `false` |
| `LOG_CONSOLE` | `--log-console` | Enable console logging | `true` |

## Getting Started

1. **Dependencies**: Ensure Consul, PostgreSQL, and RabbitMQ are running.
2. **Build**:
   ```bash
   go build -o web-meeting-backend main.go
   ```
3. **Run**:
   ```bash
   ./web-meeting-backend server
   ```
