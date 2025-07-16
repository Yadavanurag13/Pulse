# Health Tracker Project

This project aims to build a cloud-native personal health dashboard and activity tracker using Go microservices, Docker, Kubernetes, and CI/CD pipelines.

## Project Structure

- `services/`: Contains all individual Go microservices.
- `kubernetes/`: Kubernetes manifests for deployment.
- `terraform/`: Infrastructure as Code (IaC) for cloud resources.
- `.github/`: CI/CD workflows using GitHub Actions.

## Getting Started

### Prerequisites

- Go (1.22+)
- Docker
- Make

### Running User Service Locally

1.  Navigate to the project root:
    `cd health-tracker-project`
2.  Build the `user-service` binary:
    `make build-user-service`
3.  Run the `user-service` binary:
    `make run-user-service`

    The service should start on `http://localhost:8080`.

    - Access users: `http://localhost:8080/users`
    - Health check: `http://localhost:8080/health`

### Running User Service with Docker

1.  Navigate to the project root:
    `cd health-tracker-project`
2.  Build the `user-service` Docker image:
    `make docker-build-user-service`
3.  Run the `user-service` Docker container:
    `make docker-run-user-service`

    The service should be accessible on `http://localhost:8080`.

### Cleaning Up

To remove built binaries and Docker images:
`make clean`