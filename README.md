# Health Tracker Project

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/your-username/your-repo/actions) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) ## 🚀 Project Overview

The **Health Tracker Project** aims to be a modern, cloud-native personal health dashboard and activity tracker. It's built with a strong emphasis on modularity, scalability, and robust engineering practices. This project serves as a comprehensive example of building microservices using Go, orchestrated with Docker Compose for local development, and deployed to Kubernetes, with infrastructure managed by Terraform, all automated via CI/CD pipelines on GitHub Actions.

### What it does (User Service in focus):

Currently, the core **User Service** is implemented, providing foundational functionalities for:
* **User Registration:** Securely create new user accounts.
* **User Authentication:** Login/logout functionality using JWT (JSON Web Tokens) with HttpOnly cookies for secure session management.
* **User Management (CRUD):** API endpoints to create, retrieve (all, by ID, by email), update, and delete user profiles.
* **Health Check:** A dedicated endpoint to monitor service status.

## ✨ Features

* **Go Microservices:** Highly performant and efficient services built with Go.
* **Layered Architecture:** Clear separation of concerns with Handler, Service, and Repository layers, adhering to SOLID principles.
* **PostgreSQL Database:** Robust and reliable data storage.
* **Secure Authentication:** JWT-based authentication with `bcrypt` for password hashing and HttpOnly cookies.
* **Structured Logging:** Integrated Zap logger for configurable, multi-level (Debug, Info, Warn, Error, Fatal) logging.
* **Containerization:** Services packaged and run efficiently using Docker.
* **Local Orchestration:** Docker Compose for easy local development and multi-service management.
* **Infrastructure as Code (IaC):** Terraform to define and provision cloud resources.
* **Container Orchestration:** Kubernetes manifests for scalable deployments in production.
* **CI/CD Automation:** GitHub Actions workflows for automated testing, building, and deployment.

## 🛠️ Technologies Used

* **Backend:** Go (1.22+)
* **Database:** PostgreSQL (16-alpine)
* **Web Framework:** Go `net/http` standard library (with Go 1.22+ routing)
* **ORM/DB Driver:** `database/sql` with `github.com/lib/pq`
* **Authentication:** `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
* **Logging:** `go.uber.org/zap`
* **Containerization:** Docker, Docker Compose
* **Orchestration:** Kubernetes
* **Infrastructure:** Terraform
* **CI/CD:** GitHub Actions
* **Utilities:** `Makefile`

---

## 🚀 Getting Started

Follow these steps to get the project up and running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed:

* **Go:** Version 1.22 or higher. ([Download Go](https://go.dev/doc/install))
* **Docker:** Latest stable version. ([Install Docker Engine](https://docs.docker.com/engine/install/) | [Docker Desktop](https://www.docker.com/products/docker-desktop/))
    * Ensure the **Docker Compose plugin** is installed (usually comes with Docker Desktop or separate installation on Linux).
* **Make:** Typically pre-installed on Linux/macOS. For Windows, consider Git Bash or WSL.
* **`curl`:** For testing API endpoints from the command line.

### ⬇️ Clone the Repository

```bash
git clone [https://github.com/your-username/your-repo.git](https://github.com/your-username/your-repo.git) # Replace with your actual repo URL
cd health-tracker-project