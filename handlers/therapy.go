package handlers

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/middleware"
    "github.com/dzuura/neurodyx-be/models"
    "github.com/dzuura/neurodyx-be/services"
)

// AddTherapyQuestionHandler creates a new therapy question.
func AddTherapyQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var question models.TherapyQuestion
    if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    requiredFields := map[string]string{
        "type":    question.Type,
        "category": question.Category,
    }
    for field, value := range requiredFields {
        if value == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required field: " + field})
            return
        }
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionID, err := services.SaveTherapyQuestion(r.Context(), question, userID)
    if err != nil {
        log.Printf("Error saving therapy question: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to save therapy question: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"questionID": questionID})
}

// GetTherapyCategoriesHandler retrieves available categories for a given type with descriptions.
func GetTherapyCategoriesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    questionType := r.URL.Query().Get("type")
    if questionType == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required query parameter: type"})
        return
    }

    categories, err := services.GetTherapyCategories(r.Context(), questionType)
    if err != nil {
        log.Printf("Error retrieving therapy categories for type %s: %v", questionType, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve categories: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(categories)
}

// GetTherapyQuestionsHandler retrieves therapy questions based on type and category.
func GetTherapyQuestionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionType := r.URL.Query().Get("type")
    category := r.URL.Query().Get("category")
    if questionType == "" || category == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required query parameters: type and category"})
        return
    }

    cacheKey := questionType + ":" + category
    if cached, ok := config.LoadFromCache(config.TherapyQuestionCache, cacheKey); ok {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(cached)
        return
    }

    questions, err := services.GetTherapyQuestions(r.Context(), questionType, category, userID)
    if err != nil {
        log.Printf("Error retrieving therapy questions for type %s, category %s: %v", questionType, category, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve questions: " + err.Error()})
        return
    }

    config.StoreInCache(config.TherapyQuestionCache, cacheKey, questions)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(questions)
}

// SubmitTherapyAnswerHandler processes and saves therapy answer submissions.
func SubmitTherapyAnswerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var submission struct {
        Submissions []models.TherapySubmission `json:"submissions"`
        Type        string                     `json:"type"`
        Category    string                     `json:"category"`
    }
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if len(submission.Submissions) == 0 || len(submission.Submissions) > 100 || submission.Type == "" || submission.Category == "" {
        log.Printf("Received invalid submission: submissions: %d, type: %s, category: %s", len(submission.Submissions), submission.Type, submission.Category)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Submissions cannot be empty or exceed 100, and type/category are required"})
        return
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    totalCorrect := 0
    for _, sub := range submission.Submissions {
        result, err := services.SaveTherapyResult(r.Context(), userID, sub, "")
        if err != nil {
            log.Printf("Error submitting answer for question %s: %v", sub.QuestionID, err)
            continue
        }
        totalCorrect += result.CorrectAnswers
    }

    result := models.TherapyResult{
        Type:           submission.Type,
        Category:       submission.Category,
        CorrectAnswers: totalCorrect,
        TotalQuestions: len(submission.Submissions),
        Status:         "completed",
    }

    log.Printf("Successfully processed %d therapy submissions for userID: %s, type: %s, category: %s, correct: %d", len(submission.Submissions), userID, submission.Type, submission.Category, totalCorrect)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "result": result,
    })
}

// GetTherapyResultsHandler retrieves a user's therapy results for a specific type and category.
func GetTherapyResultsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionType := r.URL.Query().Get("type")
    category := r.URL.Query().Get("category")
    if questionType == "" || category == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required query parameters: type and category"})
        return
    }

    result, err := services.GetTherapyResults(r.Context(), userID, questionType, category)
    if err != nil {
        log.Printf("Error retrieving therapy results for type %s, category %s: %v", questionType, category, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve results: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}