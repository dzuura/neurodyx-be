package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/middleware"
    "github.com/dzuura/neurodyx-be/models"
    "github.com/dzuura/neurodyx-be/services"
)

// AddScreeningQuestionHandler creates a new screening question.
func AddScreeningQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var question models.ScreeningQuestion
    if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if question.AgeGroup == "" || question.Question == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required fields: ageGroup or question"})
        return
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionID, err := services.SaveScreeningQuestion(r.Context(), question, userID)
    if err != nil {
        log.Printf("Error saving screening question: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to save screening question: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"questionID": questionID})
}

// GetScreeningQuestionsHandler retrieves screening questions based on ageGroup.
func GetScreeningQuestionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    ageGroup := r.URL.Query().Get("ageGroup")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    if cached, ok := config.LoadFromCache(&config.ScreeningQuestionCache, ageGroup); ok {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(cached)
        return
    }

    questions, err := services.GetScreeningQuestions(r.Context(), ageGroup, userID)
    if err != nil {
        log.Printf("Error retrieving screening questions: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve screening questions: " + err.Error()})
        return
    }

    config.StoreInCache(&config.ScreeningQuestionCache, ageGroup, questions)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(questions)
}

// SubmitScreeningHandler processes and saves screening submissions.
func SubmitScreeningHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var submission models.ScreeningSubmission
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if submission.AgeGroup == "" || len(submission.Answers) == 0 || len(submission.Answers) > 50 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required field: ageGroup, or answers empty/exceed 50"})
        return
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questions, err := services.GetScreeningQuestions(r.Context(), submission.AgeGroup, userID)
    if err != nil {
        log.Printf("Error retrieving screening questions: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve screening questions: " + err.Error()})
        return
    }

    expectedAnswerCount := len(questions)
    if expectedAnswerCount == 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "No screening questions found for ageGroup: " + submission.AgeGroup})
        return
    }
    if len(submission.Answers) != expectedAnswerCount {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid number of answers: expected " + strconv.Itoa(expectedAnswerCount) + ", got " + strconv.Itoa(len(submission.Answers))})
        return
    }

    score := 0
    for _, answer := range submission.Answers {
        if answer {
            score++
        }
    }

    totalQuestions := float64(expectedAnswerCount)
    truePercentage := float64(score) / totalQuestions * 100
    var riskLevel string
    switch {
    case truePercentage <= 40:
        riskLevel = "low"
    case truePercentage <= 70:
        riskLevel = "moderate"
    default:
        riskLevel = "high"
    }

    err = services.SaveScreeningResult(r.Context(), userID, submission.AgeGroup, submission.Answers, riskLevel, "")
    if err != nil {
        log.Printf("Error saving screening result: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to save screening result: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"riskLevel": riskLevel})
}