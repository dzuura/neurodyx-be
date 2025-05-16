<h1 align="center"> Neurodyx Backend </h1>

## 📖 Overview

Neurodyx Backend is a robust Go-based API server that powers the Neurodyx application, a platform designed to support cognitive development through therapy, assessment, and screening services. This backend manages questions, user submissions, and progress tracking, integrating seamlessly with Google Cloud Firestore for data storage and employing JWT-based authentication for secure access.

## ✨ Key Features

- ✅ **Screening Management**: Create, retrieve, update, and delete screening questions for different age groups (e.g., adult, kid).
- 🧠 **Assessment Management**: Handle assessment questions across various types (e.g., visual, auditory) and categories (e.g., letter_recognition, word_recognition).
- 🧩 **Therapy Management**: Manage therapy questions with detailed descriptions, categories, and multimedia support (e.g., sound URLs).
- 📥 **User Submission Processing**: Process and validate user answers for screening, assessment, and therapy, calculating scores and risk levels.
- 📊 **Progress Tracking**: Monitor user progress on a weekly and monthly basis.
- 🔐 **Secure Authentication**: Implement JWT-based authentication with Firebase integration and admin role checks.
- ⚡ **Performance Optimization**: Utilize in-memory caching for frequently accessed data to enhance performance.

## 🛠 Tech Stack

- **Language**: Go
- **Database**: Google Cloud Firestore
- **Web Framework**: [Gorilla Mux](https://github.com/gorilla/mux)
- **Authentication**: Firebase Authentication with JWT using [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- **Caching**: In-memory caching with [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- **Other Libraries**:
  - [cloud.google.com/go/firestore](https://cloud.google.com/firestore/docs/client-libraries) for Firestore integration

## 🚀 Getting Started

### 📦 Prerequisites

- **Go**: Version 1.24 or later ([Download Go](https://go.dev/dl/))
- **Google Cloud SDK**: With Firestore enabled ([Install SDK](https://cloud.google.com/sdk/docs/install))
- **Firebase Project**: With Authentication and Firestore enabled ([Set Up Firebase](https://firebase.google.com/docs))
- **Service Account Key**: For Firestore authentication ([Create Key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys))
- **Postman**: For testing API endpoints ([Download Postman](https://www.postman.com/downloads/))

### 🔧 Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/dzuura/neurodyx-be.git
   cd neurodyx-be
   ```

2. **Set Up Environment Variables**:
   - Copy `.env.example` to `.env`:
     ```bash
     cp .env.example .env
     ```
   - Update `.env` with your Firebase credentials and other configurations:
     ```env
     GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceAccountKey.json
     FIREBASE_API_KEY=your-firebase-api-key
     ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Run the Application**:
   ```bash
   go run main.go
   ```
   The server will start at `http://localhost:8080` (or the port specified in `.env`).

## 📁 Project Structure

```
neurodyx-be/
├── config/                  # Configuration for Firestore and caching
│   └── firebase.go          # Firestore client initialization
├── handlers/                # HTTP handlers for API endpoints
│   ├── assessment.go        # Assessment-related endpoints
│   ├── auth.go              # Authentication endpoints
│   ├── progress.go          # Progress tracking endpoints
│   ├── screening.go         # Screening-related endpoints
│   └── therapy.go           # Therapy-related endpoints
├── middleware/              # Middleware for authentication, rate limiting, etc.
│   ├── admin.go             # Admin role verification
│   ├── auth.go              # JWT authentication
│   ├── panic_recovery.go    # Panic recovery
│   └── rate_limit.go        # Rate limiting
├── models/                  # Data models for requests and responses
│   ├── assessment.go        # Assessment question and result models
│   ├── error.go             # Error response model
│   ├── progress.go          # Progress tracking models
│   ├── screening.go         # Screening question and submission models
│   ├── therapy.go           # Therapy question and result models
│   └── user.go              # User and authentication models
├── services/                # Business logic and Firestore interactions
│   ├── assessment.go        # Assessment services
│   ├── firebase.go          # Firestore client setup
│   ├── screening.go         # Screening services
│   ├── therapy.go           # Therapy services
│   └── validation.go        # Answer validation logic
├── main.go                  # Application entry point
├── .env.example             # Environment variable template
├── .gcloudignore            # Google Cloud deployment ignore file
├── app.yaml                 # Google Cloud App Engine configuration
└── README.md                # Project documentation
```

## 📊 Firestore Structure

The backend relies on Firestore for data storage. Below is the structure of the main collections:

| Collection Path | Description | Fields |
|-----------------|-------------|--------|
| `screeningQuestions/{ageGroup}/questions/{questionID}` | Screening questions by age group | `ageGroup`, `question`, `timestamp` |
| `assessmentQuestions/{type}/{category}/{questionID}` | Assessment questions by type and category | `type`, `category`, `content`, `correctAnswer`, `options`, `leftItems`, `rightItems`, `correctSequence`, `correctPairs`, `timestamp` |
| `therapyQuestions/{type}/{category}/{questionID}` | Therapy questions by type and category | `type`, `category`, `content`, `description`, `imageURL`, `soundURL`, `options`, `correctAnswer`, `correctSequence`, `correctPairs`, `timestamp` |
| `users/{userID}/assessments/{type}/submissions/{questionID}` | User assessment submissions | `type`, `category`, `questionID`, `correctAnswers`, `answer`, `status`, `timestamp` |
| `users/{userID}/therapy/{type}/{category}/{questionID}` | User therapy submissions | `type`, `category`, `questionID`, `correctAnswers`, `answer`, `status`, `timestamp` |
| `users/{userID}/progress/{date}` | User progress data | `userID`, `date`, `therapyCount`, `streakAchieved` |

## 📡 API Documentation

### 🔒 Authentication

All endpoints (except Firebase registration and login) require a valid Bearer token in the `Authorization` header, obtained via the `/auth` endpoint. Admin-only endpoints (e.g., adding or updating questions) require the user to have `isAdmin: true` in their Firestore user document.

### 1. Health Check
- **Method**: GET
- **Endpoint**: `/health`
- **Description**: Checks if the server is running.
- **Response**:
  - **Status**: `200 OK`
  - **Body**: `"Neurodyx Backend is running"`
- **Error Responses**:
  - None

### 2. Authentication Endpoints
- **Firebase Register**
  - **Method**: POST
  - **Endpoint**: `https://identitytoolkit.googleapis.com/v1/accounts:signUp?key={firebase_api_key}`
  - **Description**: Registers a new user with Firebase Authentication.
  - **Request Body**:
    ```json
    {
      "email": "your-email",
      "password": "your-password",
      "returnSecureToken": true
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "idToken": "firebase-id-token",
        "email": "your-email",
        "refreshToken": "firebase-refresh-token",
        "expiresIn": "3600",
        "localId": "user-id"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Email already exists or invalid credentials.

- **Firebase Login**
  - **Method**: POST
  - **Endpoint**: `https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={firebase_api_key}`
  - **Description**: Logs in an existing user with Firebase Authentication.
  - **Request Body**:
    ```json
    {
      "email": "your-email",
      "password": "your-password",
      "returnSecureToken": true
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "idToken": "firebase-id-token",
        "email": "your-email",
        "refreshToken": "firebase-refresh-token",
        "expiresIn": "3600",
        "localId": "user-id"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid email or password.

- **Authenticate User**
  - **Method**: POST
  - **Endpoint**: `/auth`
  - **Description**: Authenticates a user using a Firebase or Google ID token and returns access and refresh tokens.
  - **Request Body**:
    ```json
    {
      "token": "firebase-or-google-id-token",
      "authType": "firebase"
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "token": "access-token",
        "refreshToken": "refresh-token"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid request body or auth type.
    - `401 Unauthorized`: Invalid token.
    - `500 Internal Server Error`: Authentication service unavailable.

- **Refresh Token**
  - **Method**: POST
  - **Endpoint**: `/refresh`
  - **Description**: Refreshes an access token using a refresh token.
  - **Request Body**:
    ```json
    {
      "refreshToken": "refresh-token"
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "token": "new-access-token",
        "refreshToken": "new-refresh-token"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid request body.
    - `401 Unauthorized`: Invalid or expired refresh token.
    - `500 Internal Server Error`: Error updating user data.

### 3. Screening Endpoints
- **Add Screening Question**
  - **Method**: POST
  - **Endpoint**: `/screening/questions`
  - **Description**: Adds a new screening question (admin only).
  - **Request Body**:
    ```json
    {
      "ageGroup": "adult",
      "question": "Do you avoid work projects or courses that require extensive reading?"
    }
    ```
  - **Response**:
    - **Status**: `201 Created`
    - **Body**:
      ```json
      {
        "questionID": "generated-id"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body or missing fields.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `500 Internal Server Error`: Failed to save to Firestore.

- **Get Screening Questions**
  - **Method**: GET
  - **Endpoint**: `/screening/questions?ageGroup={ageGroup}`
  - **Description**: Retrieves screening questions for a specific age group.
  - **Query Parameters**:
    - `ageGroup`: e.g., `adult` or `kid`
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "id": "question-id",
          "ageGroup": "adult",
          "question": "Do you avoid work projects or courses that require extensive reading?"
        }
      ]
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

- **Update Screening Question**
  - **Method**: PUT
  - **Endpoint**: `/screening/questions/{id}`
  - **Description**: Updates an existing screening question (admin only).
  - **Request Body**:
    ```json
    {
      "ageGroup": "adult",
      "question": "Updated screening question for adult"
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "id": "question-id",
        "ageGroup": "adult",
        "question": "Updated screening question for adult"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, missing fields, or attempt to change `ageGroup`.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to update in Firestore.

- **Delete Screening Question**
  - **Method**: DELETE
  - **Endpoint**: `/screening/questions/{id}`
  - **Description**: Deletes a screening question (admin only).
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "message": "Screening question deleted successfully"
      }
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to delete from Firestore.

- **Submit Screening Answers**
  - **Method**: POST
  - **Endpoint**: `/screening/submit`
  - **Description**: Submits screening answers and calculates risk level.
  - **Request Body**:
    ```json
    {
      "ageGroup": "adult",
      "answers": [true, false, false, false, false, false, true, true, true, true]
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "riskLevel": "moderate"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, missing fields, or incorrect number of answers.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to save results.

### 4. Assessment Endpoints
- **Add Assessment Question**
  - **Method**: POST
  - **Endpoint**: `/assessment/questions`
  - **Description**: Adds a new assessment question (admin only).
  - **Request Body**:
    ```json
    {
      "type": "visual",
      "category": "letter_recognition",
      "content": "m",
      "options": ["w", "m"],
      "correctAnswer": "m"
    }
    ```
  - **Response**:
    - **Status**: `201 Created`
    - **Body**:
      ```json
      {
        "questionID": "generated-id"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body or missing fields.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `500 Internal Server Error`: Failed to save to Firestore.

- **Get Assessment Questions**
  - **Method**: GET
  - **Endpoint**: `/assessment/questions?type={type}`
  - **Description**: Retrieves assessment questions for a specific type or all types.
  - **Query Parameters**:
    - `type`: e.g., `visual`, `auditory`, `kinesthetic`, `tactile`, or omitted for all types
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "id": "question-id",
          "type": "visual",
          "category": "letter_recognition",
          "content": "m",
          "options": ["w", "m"],
          "correctAnswer": "m"
        }
      ]
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

- **Update Assessment Question**
  - **Method**: PUT
  - **Endpoint**: `/assessment/questions/{id}`
  - **Description**: Updates an existing assessment question (admin only).
  - **Request Body**:
    ```json
    {
      "type": "visual",
      "category": "letter_recognition",
      "content": "x",
      "options": ["x", "o"],
      "correctAnswer": "x"
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "id": "question-id",
        "type": "visual",
        "category": "letter_recognition",
        "content": "x",
        "options": ["x", "o"],
        "correctAnswer": "x"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, missing fields, or attempt to change `type`/`category`.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to update in Firestore.

- **Delete Assessment Question**
  - **Method**: DELETE
  - **Endpoint**: `/assessment/questions/{id}`
  - **Description**: Deletes an assessment question (admin only).
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "message": "Assessment question deleted successfully"
      }
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to delete from Firestore.

- **Submit Assessment Answers**
  - **Method**: POST
  - **Endpoint**: `/assessment/submit`
  - **Description**: Submits assessment answers and calculates scores.
  - **Request Body**:
    ```json
    {
      "type": "auditory",
      "submissions": [
        {"questionId": "question-id", "answer": "p"},
        {"questionId": "question-id", "answer": "cat"}
      ]
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "result": {
          "type": "auditory",
          "correctAnswers": 1,
          "totalQuestions": 2,
          "status": "completed"
        }
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, empty submissions, or exceeding 100 submissions.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to save results.

- **Get Assessment Results**
  - **Method**: GET
  - **Endpoint**: `/assessment/results`
  - **Description**: Retrieves assessment results for a user.
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "type": "auditory",
          "correctAnswers": 1,
          "totalQuestions": 2,
          "status": "completed"
        }
      ]
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

### 5. Therapy Endpoints
- **Add Therapy Question**
  - **Method**: POST
  - **Endpoint**: `/therapy/questions`
  - **Description**: Adds a new therapy question (admin only).
  - **Request Body**:
    ```json
    {
      "type": "tactile",
      "category": "word_recognition_by_touch",
      "description": "Can you draw this letter? Try writing it in the box!",
      "content": "w",
      "soundURL": "your-sound-url",
      "correctAnswer": "w"
    }
    ```
  - **Response**:
    - **Status**: `201 Created`
    - **Body**:
      ```json
      {
        "questionID": "generated-id"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body or missing fields.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `500 Internal Server Error`: Failed to save to Firestore.

- **Get Therapy Categories**
  - **Method**: GET
  - **Endpoint**: `/therapy/categories?type={type}`
  - **Description**: Retrieves therapy categories for a specific type.
  - **Query Parameters**:
    - `type`: e.g., `kinesthetic`, `visual`, `auditory`, `tactile`
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "category": "number_letter_similarity",
          "description": "Identify similarities between numbers and letters"
        },
        {
          "category": "letter_matching",
          "description": "Drag the right letters to complete the word"
        }
      ]
      ```
  - **Error Responses**:
    - `400 Bad Request`: Missing `type` parameter.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

- **Get Therapy Questions**
  - **Method**: GET
  - **Endpoint**: `/therapy/questions?type={type}&category={category}`
  - **Description**: Retrieves therapy questions for a specific type and category.
  - **Query Parameters**:
    - `type`: e.g., `kinesthetic`, `visual`, `auditory`, `tactile`
    - `category`: e.g., `number_letter_similarity`, `word_recognition_by_touch`
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "id": "question-id",
          "type": "kinesthetic",
          "category": "number_letter_similarity",
          "content": "Compare 5 and S",
          "description": "Identify similarities between numbers and letters",
          "correctAnswer": "similar",
          "imageURL": "",
          "soundURL": "",
          "options": ["similar", "different"],
          "leftItems": null,
          "rightItems": null,
          "correctSequence": null,
          "correctPairs": null
        }
      ]
      ```
  - **Error Responses**:
    - `400 Bad Request`: Missing `type` or `category` parameters.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

- **Update Therapy Question**
  - **Method**: PUT
  - **Endpoint**: `/therapy/questions/{id}`
  - **Description**: Updates an existing therapy question (admin only).
  - **Request Body**:
    ```json
    {
      "type": "tactile",
      "category": "word_recognition_by_touch",
      "description": "Can you draw this letter? Try writing it in the box!",
      "content": "M",
      "soundURL": "your-sound-url",
      "correctAnswer": "M"
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "id": "question-id",
        "type": "tactile",
        "category": "word_recognition_by_touch",
        "description": "Can you draw this letter? Try writing it in the box!",
        "content": "M",
        "soundURL": "your-sound-url",
        "correctAnswer": "M"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, missing fields, or attempt to change `type`/`category`.
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to update in Firestore.

- **Delete Therapy Question**
  - **Method**: DELETE
  - **Endpoint**: `/therapy/questions/{id}`
  - **Description**: Deletes a therapy question (admin only).
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "message": "Therapy question deleted successfully"
      }
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `403 Forbidden`: User is not an admin.
    - `404 Not Found`: Question not found.
    - `500 Internal Server Error`: Failed to delete from Firestore.

- **Submit Therapy Answers**
  - **Method**: POST
  - **Endpoint**: `/therapy/submit`
  - **Description**: Submits therapy answers and calculates scores.
  - **Request Body**:
    ```json
    {
      "type": "visual",
      "category": "letter_recognition",
      "submissions": [
        {"questionID": "question-id", "answer": "c"},
        {"questionID": "question-id", "answer": "a"}
      ]
    }
    ```
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "result": {
          "type": "visual",
          "category": "letter_recognition",
          "correctAnswers": 1,
          "totalQuestions": 2,
          "status": "completed"
        }
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid body, empty submissions, or exceeding 100 submissions.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to save results.

- **Get Therapy Results**
  - **Method**: GET
  - **Endpoint**: `/therapy/results?type={type}&category={category}`
  - **Description**: Retrieves therapy results for a user.
  - **Query Parameters**:
    - `type`: e.g., `tactile`, `visual`, `auditory`, `kinesthetic`
    - `category`: e.g., `complete_the_word_by_touch`, `letter_recognition`
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      {
        "type": "tactile",
        "category": "complete_the_word_by_touch",
        "correctAnswers": 1,
        "totalQuestions": 2,
        "status": "completed"
      }
      ```
  - **Error Responses**:
    - `400 Bad Request`: Missing `type` or `category` parameters.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

### 6. Progress Endpoints
- **Get Weekly Progress**
  - **Method**: GET
  - **Endpoint**: `/progress/weekly`
  - **Description**: Retrieves daily progress for the last 7 days.
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "userID": "user-id",
          "date": "2025-05-10T00:00:00Z",
          "therapyCount": 3,
          "streakAchieved": false
        }
      ]
      ```
  - **Error Responses**:
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

- **Get Monthly Progress**
  - **Method**: GET
  - **Endpoint**: `/progress/monthly?year={year}&month={month}`
  - **Description**: Retrieves monthly progress for a specific year and month.
  - **Query Parameters**:
    - `year`: e.g., `2025`
    - `month`: 1-12, e.g., `4`
  - **Response**:
    - **Status**: `200 OK`
    - **Body**:
      ```json
      [
        {
          "date": "2025-04-01T00:00:00Z",
          "status": "active"
        }
      ]
      ```
  - **Error Responses**:
    - `400 Bad Request`: Invalid or missing `year`/`month` parameters.
    - `401 Unauthorized`: Missing or invalid token.
    - `500 Internal Server Error`: Failed to retrieve from Firestore.

## 🔒 Authentication and Security

The backend uses Firebase Authentication for user registration and login, followed by JWT-based authentication for API access. A valid token must be included in the `Authorization` header for all protected endpoints. Admin-only endpoints (e.g., adding or updating questions) verify the `isAdmin` field in the user's Firestore document. Rate limiting is implemented to prevent abuse, and panic recovery middleware ensures robust error handling.

## ☁️ Deployment

The backend can be deployed using Docker or directly to a cloud platform like Google Cloud App Engine. For Google Cloud deployment:

1. **Configure `app.yaml`**:
   Ensure `app.yaml` is set up for your environment:
   ```yaml
   runtime: go120
   instance_class: F1
   env_variables:
     GOOGLE_APPLICATION_CREDENTIALS: "/path/to/serviceAccountKey.json"
     JWT_SECRET: your-secret-key
     GOOGLE_CLIENT_ID: your-google-client-id
   ```

2. **Deploy to App Engine**:
   ```bash
   gcloud app deploy
   ```

3. **Access the Application**:
   After deployment, the application will be accessible at the URL provided by Google Cloud.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'Add some amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgements

- [Go](https://go.dev/)
- [Firebase](https://firebase.google.com/)
- [Gorilla Mux](https://github.com/gorilla/mux)
- [Google Cloud Firestore](https://cloud.google.com/firestore)
- All the amazing open-source libraries used in this project.
