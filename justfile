set shell := ["powershell.exe", "-c"]

# Variables

compose_file := "deployments/docker-compose.yml"
docker_cmd := "docker-compose -f " + compose_file

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
