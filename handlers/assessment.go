package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"

    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/middleware"
    "github.com/dzuura/neurodyx-be/models"
    "github.com/dzuura/neurodyx-be/services"
)

// AddAssessmentQuestionHandler creates a new assessment question.
func AddAssessmentQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var question models.AssessmentQuestion
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

    questionID, err := services.SaveAssessmentQuestion(r.Context(), question, userID)
    if err != nil {
        log.Printf("Error saving assessment question: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to save assessment question: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"questionID": questionID})
}

// GetAssessmentQuestionsHandler retrieves assessment questions based on type.
func GetAssessmentQuestionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    types := r.URL.Query().Get("type")
    var typeFilters []string
    if types == "all" || types == "" {
        typeFilters = []string{"visual", "auditory", "kinesthetic", "tactile"}
    } else {
        typeFilters = strings.Split(types, ",")
    }

    questions := []models.AssessmentQuestion{}
    for _, t := range typeFilters {
        if cached, ok := config.LoadFromCache(&config.AssessmentQuestionCache, t); ok {
            questions = append(questions, cached.([]models.AssessmentQuestion)...)
            continue
        }
        qs, err := services.GetAssessmentQuestions(r.Context(), t, userID)
        if err != nil {
            log.Printf("Error retrieving assessment questions for type %s: %v", t, err)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve questions: " + err.Error()})
            return
        }
        questions = append(questions, qs...)
        config.StoreInCache(&config.AssessmentQuestionCache, t, qs)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(questions)
}

// SubmitAnswerHandler processes and saves assessment answer submissions.
func SubmitAnswerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var submission struct {
        Submissions []models.AssessmentSubmission `json:"submissions"`
        Type        string                        `json:"type"`
    }
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if len(submission.Submissions) == 0 || len(submission.Submissions) > 100 || submission.Type == "" {
        log.Printf("Received invalid submission: submissions: %d, type: %s", len(submission.Submissions), submission.Type)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Submissions cannot be empty or exceed 100, and type is required"})
        return
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    totalCorrect := 0
    typeMap := make(map[string]int)
    for _, sub := range submission.Submissions {
        result, err := services.SaveAssessmentResult(r.Context(), userID, sub, "")
        if err != nil {
            log.Printf("Error submitting answer for question %s: %v", sub.QuestionID, err)
            continue
        }
        totalCorrect += result.CorrectAnswers
        typeMap[submission.Type] = typeMap[submission.Type] + result.CorrectAnswers
    }

    result := models.AssessmentResult{
        Type:           submission.Type,
        CorrectAnswers: totalCorrect,
        TotalQuestions: len(submission.Submissions),
        Status:         "completed",
    }

    log.Printf("Successfully processed %d submissions for userID: %s, type: %s, correct: %d", len(submission.Submissions), userID, submission.Type, totalCorrect)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "result": result,
    })
}

// GetAssessmentResultsHandler retrieves a user's assessment results.
func GetAssessmentResultsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    results, err := services.GetAssessmentResults(r.Context(), userID, "")
    if err != nil {
        log.Printf("Error retrieving assessment results: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve results: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(results)
}