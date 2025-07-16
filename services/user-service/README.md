## 🔌 API Endpoints

The `user-service` provides a RESTful API. Below are the primary endpoints:

* **Base URL (Local Docker Compose):** `http://localhost:8080/v1`
* **Base URL (Minikube):** `http://<MINIKUBE_IP>:<NODEPORT>/v1` (Use the URL from `make k8s-get-user-service-url`)

### `GET /health`
* **Description:** Checks the health and readiness of the user service. (Typically not versioned)
* **Response:** `200 OK` with "User Service is healthy"

### `POST /v1/users`

* **Description:** Creates a new user.
* **Request Body (JSON):**
    ```json
    {
      "name": "John Doe",
      "email": "john.doe@example.com",
      "password": "strongpassword123"
    }
    ```
* **Response (JSON):** `201 Created` with the newly created user's details (including generated ID).

### `GET /v1/users`

* **Description:** Retrieves a list of all registered users.
* **Response (JSON):** `200 OK` with an array of user objects.

### `GET /v1/users/{id}`

* **Description:** Retrieves a specific user by their ID.
* **URL Parameter:** `{id}` - The UUID of the user.

### `GET /v1/users/by-email?email={email}`

* **Description:** Retrieves a specific user by their email address.
* **Query Parameter:** `email` - The email address of the user.

### `PUT /v1/users/{id}`

* **Description:** Updates an existing user's details.
* **URL Parameter:** `{id}` - The UUID of the user.
* **Request Body (JSON):** (Provide an example, similar to POST)

### `DELETE /v1/users/{id}`

* **Description:** Deletes a user by their ID.
* **URL Parameter:** `{id}` - The UUID of the user.
