package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    "github.com/dzuura/neurodyx-be/models"
    "github.com/dzuura/neurodyx-be/services"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
    Error string `json:"error"`
}

// AddScreeningQuestionHandler handles the creation of a new screening question
func AddScreeningQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Decode request body
    var question models.ScreeningQuestion
    if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    // Validate required fields
    if question.AgeGroup == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing required field: ageGroup"})
        return
    }
    if question.Question == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing required field: question"})
        return
    }

    // Extract firebaseToken from context
    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Authentication token missing"})
        return
    }

    // Save the question
    questionID, err := services.SaveScreeningQuestion(r.Context(), question, firebaseToken)
    if err != nil {
        log.Printf("Error saving screening question: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to save screening question: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"questionID": questionID})
}

// GetScreeningQuestionsHandler retrieves screening questions, optionally filtered by ageGroup
func GetScreeningQuestionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Extract ageGroup from query parameters
    ageGroup := r.URL.Query().Get("ageGroup")

    // Extract firebaseToken from context
    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Authentication token missing"})
        return
    }

    // Retrieve questions
    questions, err := services.GetScreeningQuestions(r.Context(), ageGroup, firebaseToken)
    if err != nil {
        log.Printf("Error retrieving screening questions: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve screening questions: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(questions)
}

// SubmitScreeningHandler handles the submission of screening answers and calculates the risk level
func SubmitScreeningHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Decode request body
    var submission models.ScreeningSubmission
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    // Validate required fields
    if submission.AgeGroup == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing required field: ageGroup"})
        return
    }
    if len(submission.Answers) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Answers cannot be empty"})
        return
    }

    // Extract firebaseToken and userID from context
    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Authentication token missing"})
        return
    }
    userID, ok := r.Context().Value("userID").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID missing"})
        return
    }

    // Fetch the number of screening questions for the given ageGroup
    questions, err := services.GetScreeningQuestions(r.Context(), submission.AgeGroup, firebaseToken)
    if err != nil {
        log.Printf("Error retrieving screening questions: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve screening questions: " + err.Error()})
        return
    }

    // Validate the number of answers matches the number of questions
    expectedAnswerCount := len(questions)
    if expectedAnswerCount == 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "No screening questions found for ageGroup: " + submission.AgeGroup})
        return
    }
    if len(submission.Answers) != expectedAnswerCount {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid number of answers: expected " + strconv.Itoa(expectedAnswerCount) + ", got " + strconv.Itoa(len(submission.Answers))})
        return
    }

    // Calculate score based on answers
    score := 0
    for _, answer := range submission.Answers {
        if answer {
            score++
        }
    }

    // Determine risk level based on percentage of "true" answers
    totalQuestions := float64(expectedAnswerCount)
    truePercentage := float64(score) / totalQuestions * 100
    var riskLevel string
    switch {
    case truePercentage < 40:
        riskLevel = "low"
    case truePercentage <= 70:
        riskLevel = "moderate"
    default:
        riskLevel = "high"
    }

    // Save the submission
    err = services.SaveScreeningResult(r.Context(), userID, submission.AgeGroup, submission.Answers, riskLevel, firebaseToken)
    if err != nil {
        log.Printf("Error saving screening result: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to save screening result: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"riskLevel": riskLevel})
}