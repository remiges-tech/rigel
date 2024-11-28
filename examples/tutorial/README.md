# Rigel Tutorial Example

This tutorial demonstrates a complete setup of Rigel with etcd, including:
- Setting up etcd using Docker Compose
- Creating and managing Rigel schemas
- Setting up configuration values using rigelctl
- Using configuration values in a Go application

## Directory Structure

```
tutorial/
├── docker-compose.yml    # Docker Compose configuration for etcd
├── setup.sh             # Configuration setup script
├── main.go              # Sample Go application
├── usersvc-schema.json  # Rigel schema definition
└── README.md            # This file
```

## Getting Started

1. Start etcd:
   ```bash
   docker compose up -d
   ```

2. Run the setup script:
   ```bash
   ./setup.sh
   ```

The setup script will:
- Wait for etcd to be ready
- Install rigelctl if not already installed
- Load the schema definition
- Set up all configuration values with retry logic

3. Run the sample Go application:
   ```bash
   go run main.go
   ```

The Go application demonstrates:
- Connecting to etcd
- Reading configuration values
- Proper error handling

## Schema Structure

The schema defines configuration for a user service including:
- Database connection details (host, port, user, password, dbname)
- Server configuration (port)

## Sample Go Application

The `main.go` file shows how to:
- Initialize etcd storage
- Create and use the Rigel client
- Read configuration values using proper paths
- Handle configuration-related errors

## Cleanup

To stop and remove etcd container:
```bash
docker compose down
