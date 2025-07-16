.PHONY: test-user-service compose-up compose-down compose-build-user-service compose-logs-user-service compose-clean-volumes format lint # ADDED format and lint here

# User Service specific variables
USER_SERVICE_PATH := services/user-service
USER_SERVICE_IMAGE := health-tracker/user-service
USER_SERVICE_TAG := latest
USER_SERVICE_CONTAINER_NAME := health-tracker-user-service # Matches container_name in docker-compose.yml

# --- Code Quality Targets ---
# Go formatting
format:
	@echo "Running gofmt on all Go files..."
	find . -type f -name "*.go" | xargs gofmt -w
	@echo "gofmt completed."

# Go linting using golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./... # Runs lint on all packages in the current directory and its subdirectories
	@echo "golangci-lint completed."

# --- Testing Targets ---
# Test the user-service (Go unit tests)
test-user-service:
	@echo "Running tests for user-service..."
	cd $(USER_SERVICE_PATH) && go test ./...

# You could also create a combined 'test' target that runs both unit tests and linting:
test: test-user-service lint format # This target will run user-service tests, then lint, then format.
.PHONY: test


# --- Docker Compose Build Targets ---
# Build the Docker image for user-service using docker-compose
# This will build only the user-service image defined in docker-compose.yml
compose-build-user-service:
	@echo "Building user-service Docker image via Docker Compose..."
	docker compose build user-service
	@echo "User Service Docker image built."

# --- Docker Compose Lifecycle Targets ---
# Start all services defined in docker-compose.yml
# This will build images if they don-t exist, then start containers.
compose-up:
	@echo "Starting all services via Docker Compose..."
	docker compose up --build -d # --build to rebuild images if needed, -d for detached mode
	@echo "Services are starting up. Check logs for status."
	@echo "Access user-service: http://localhost:8080/users"
	@echo "To view user-service logs: make compose-logs-user-service"
	@echo "To stop all services: make compose-down"

# Stop all services defined in docker-compose.yml
compose-down:
	@echo "Stopping all services via Docker Compose..."
	docker compose down
	@echo "All services stopped."

# View logs for the user-service
compose-logs-user-service:
	@echo "Streaming logs for user-service..."
	docker compose logs -f user-service

# Clean up Docker images and volumes
compose-clean-volumes:
	@echo "Stopping and removing all services, including volumes..."
	docker compose down --volumes --rmi all
	@echo "All services, images, and volumes cleaned."

# Alias for a general clean (stops services, removes images and volumes)
clean: compose-clean-volumes