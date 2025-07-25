## 🔌 API Endpoints

The `user-service` provides a RESTful API for user authentication and management. All protected endpoints require a valid JWT (JSON Web Token) to be sent as an `HttpOnly` cookie named `jwt_token`.

* **Base URL (Local Docker Compose):** `http://localhost:8080` (Note: `/v1` is handled by the application's routing, not part of the base URL here.)
* **Base URL (Minikube):** `http://<MINIKUBE_IP>:<NODEPORT>` (Use the URL from `make k8s-get-user-service-url`)

---

### **Public Endpoints (No Authentication Required)**

These endpoints are accessible without a JWT.

#### `GET /health`
* **Description:** Checks the health and readiness of the user service.
* **Response:** `200 OK` with plain text:
    ```
    User Service is healthy
    ```
* **`curl` Example:**
    ```bash
    curl http://localhost:8080/health
    ```

#### `POST /register`
* **Description:** Creates a new user account.
* **Request Body (JSON):**
    ```json
    {
      "name": "John Doe",
      "email": "john.doe@example.com",
      "password": "SecurePassword123"
    }
    ```
* **Response (JSON):** `201 Created` with the newly created user's public details.
    ```json
    {
      "id": "a-uuid-for-the-user",
      "name": "John Doe",
      "email": "john.doe@example.com",
      "created_at": "2025-07-24T12:00:00Z"
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If required fields are missing.
    * `409 Conflict`: If a user with the provided email already exists.
* **`curl` Example:**
    ```bash
    curl -X POST \
      http://localhost:8080/register \
      -H 'Content-Type: application/json' \
      -d '{
        "name": "John Doe",
        "email": "john.doe@example.com",
        "password": "SecurePassword123"
      }'
    ```

#### `POST /login`
* **Description:** Authenticates a user and issues a JWT token.
* **Request Body (JSON):**
    ```json
    {
      "email": "john.doe@example.com",
      "password": "SecurePassword123"
    }
    ```
* **Response (JSON):** `200 OK` with the JWT token, user details, and token expiration. The JWT is also set as an `HttpOnly` cookie named `jwt_token`.
    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiI...",
      "user": {
        "id": "a-uuid-for-the-user",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "created_at": "2025-07-24T12:00:00Z"
      },
      "expires_in_sec": 900
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If required fields are missing.
    * `401 Unauthorized`: If credentials are invalid.
* **`curl` Example (Crucial for capturing the cookie for subsequent requests):**
    ```bash
    curl -X POST \
      http://localhost:8080/login \
      -H 'Content-Type: application/json' \
      -c cookies.txt \
      -d '{
        "email": "john.doe@example.com",
        "password": "SecurePassword123"
      }'
    ```

---

### **Protected Endpoints (Authentication Required)**

These endpoints require a valid `jwt_token` cookie obtained from the `/login` endpoint. Use `-b cookies.txt` in your `curl` commands.

#### `GET /protected`
* **Description:** An example endpoint to verify JWT authentication.
* **Response (JSON):** `200 OK` if authenticated.
    ```json
    {
      "message": "Welcome to the protected area, User ID: a-uuid-for-the-user!"
    }
    ```
* **Error Responses:**
    * `401 Unauthorized`: If no token is provided or the token is invalid/expired.
* **`curl` Example:**
    ```bash
    curl -X GET \
      http://localhost:8080/protected \
      -b cookies.txt
    ```

#### `POST /users`
* **Description:** Creates a new user (often used for admin-like creation, distinct from `/register`).
* **Request Body (JSON):**
    ```json
    {
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "password": "AnotherSecurePwd456"
    }
    ```
* **Response (JSON):** `201 Created` with the newly created user's details.
    ```json
    {
      "id": "a-new-uuid-for-jane-smith",
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "created_at": "2025-07-24T12:00:00Z"
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If required fields are missing.
    * `401 Unauthorized`: If not authenticated.
    * `409 Conflict`: If a user with the provided email already exists.
* **`curl` Example:**
    ```bash
    curl -X POST \
      http://localhost:8080/users \
      -H 'Content-Type: application/json' \
      -b cookies.txt \
      -d '{
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "password": "AnotherSecurePwd456"
      }'
    ```

#### `GET /users`
* **Description:** Retrieves a list of all registered users.
* **Response (JSON):** `200 OK` with an array of user objects.
    ```json
    [
      {
        "id": "uuid-of-john-doe",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "created_at": "2025-07-24T12:00:00Z"
      },
      {
        "id": "uuid-of-jane-smith",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "created_at": "2025-07-24T12:05:00Z"
      }
    ]
    ```
* **Error Responses:**
    * `401 Unauthorized`: If not authenticated.
* **`curl` Example:**
    ```bash
    curl -X GET \
      http://localhost:8080/users \
      -b cookies.txt
    ```

#### `GET /users/{id}`
* **Description:** Retrieves a specific user by their ID.
* **URL Parameter:** `{id}` - The UUID of the user.
* **Response (JSON):** `200 OK` with the user's details.
    ```json
    {
      "id": "uuid-of-john-doe",
      "name": "John Doe",
      "email": "john.doe@example.com",
      "created_at": "2025-07-24T12:00:00Z"
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If the ID format is invalid.
    * `401 Unauthorized`: If not authenticated.
    * `404 Not Found`: If no user with the given ID exists.
* **`curl` Example:**
    ```bash
    curl -X GET \
      http://localhost:8080/users/YOUR_USER_ID_HERE \
      -b cookies.txt
    ```

#### `GET /users/by-email?email={email}`
* **Description:** Retrieves a specific user by their email address.
* **Query Parameter:** `email` - The email address of the user.
* **Response (JSON):** `200 OK` with the user's details.
    ```json
    {
      "id": "uuid-of-jane-smith",
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "created_at": "2025-07-24T12:05:00Z"
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If the `email` query parameter is missing.
    * `401 Unauthorized`: If not authenticated.
    * `404 Not Found`: If no user with the given email exists.
* **`curl` Example:**
    ```bash
    curl -X GET \
      'http://localhost:8080/users/by-email?email=jane.smith@example.com' \
      -b cookies.txt
    ```

#### `PUT /users/{id}`
* **Description:** Updates an existing user's details.
* **URL Parameter:** `{id}` - The UUID of the user to update.
* **Request Body (JSON):** Provide fields to update. `password` is optional (`omitempty`).
    ```json
    {
      "name": "Jane Updated",
      "email": "jane.updated@example.com",
      "password": "NewSecurePassword789" # Optional: omit this field if not updating password
    }
    ```
* **Response (JSON):** `200 OK` with the updated user's public details.
    ```json
    {
      "id": "uuid-of-jane-smith",
      "name": "Jane Updated",
      "email": "jane.updated@example.com",
      "created_at": "2025-07-24T12:05:00Z"
    }
    ```
* **Error Responses:**
    * `400 Bad Request`: If the request payload is invalid or validation fails (e.g., new email already in use, missing required fields if you add more validation).
    * `401 Unauthorized`: If not authenticated.
    * `404 Not Found`: If the user with the given ID does not exist.
* **`curl` Example:**
    ```bash
    curl -X PUT \
      http://localhost:8080/users/YOUR_USER_ID_HERE \
      -H 'Content-Type: application/json' \
      -b cookies.txt \
      -d '{
        "name": "Jane Updated",
        "email": "jane.updated@example.com"
      }'
    ```

#### `DELETE /users/{id}`
* **Description:** Deletes a user by their ID.
* **URL Parameter:** `{id}` - The UUID of the user to delete.
* **Response:** `204 No Content` on successful deletion.
* **Error Responses:**
    * `400 Bad Request`: If the ID format is invalid.
    * `401 Unauthorized`: If not authenticated.
    * `404 Not Found`: If the user with the given ID does not exist.
* **`curl` Example:**
    ```bash
    curl -X DELETE \
      http://localhost:8080/users/USER_ID_TO_DELETE_HERE \
      -b cookies.txt
    ```

#### `POST /logout`
* **Description:** Logs out the current user by invalidating their JWT cookie.
* **Response (JSON):** `200 OK`
    ```json
    {
      "message": "Logged out successfully"
    }
    ```
* **Error Responses:**
    * `401 Unauthorized`: If not authenticated (though it will still attempt to clear the cookie).
* **`curl` Example:**
    ```bash
    curl -X POST \
      http://localhost:8080/logout \
      -b cookies.txt \
      -c cookies.txt # This will ensure the cookie is cleared in your local cookies.txt file
    ```