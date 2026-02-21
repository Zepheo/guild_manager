set shell := ["powershell.exe", "-c"]

# Variables

compose_file := "deployments/docker-compose.yml"
docker_cmd := "docker-compose -f " + compose_file
DB_USER := env("DB_USER", "admin")
DB_NAME := env("DB_NAME", "sr_plus")

# Start the whole stack
up:
    {{ docker_cmd }} up -d

# Stop the stack
down:
    {{ docker_cmd }} down

# Rebuild a specific service (defaults to 'api')

# Usage: `just build` or `just build db`
build service='api':
    {{ docker_cmd }} build {{ service }}

# Show logs for a specific service (defaults to 'api')

# Usage: `just logs` or `just logs db`
logs service='api':
    {{ docker_cmd }} logs -f {{ service }}

# Run Go tests

# Usage: `just test` or `just test ./pkg/auth`
test path='./...':
    go test -v  {{ path }}

#

# Backup the production database to a timestamped .sql file
backup:
    @mkdir -p backups
    docker exec guild_manager_db pg_dump -U {{ DB_USER }} {{ DB_NAME }} > backups/backup_$(date +%Y%m%d_%H%M%S).sql
    @echo "Backup saved to backups/ folder."
