package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/dzuura/neurodyx-be/config"
	"github.com/dzuura/neurodyx-be/models"
	"github.com/dzuura/neurodyx-be/services"
)

// AddAssessmentQuestionHandler handles the creation of a new assessment question
func AddAssessmentQuestionHandler(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in AddAssessmentQuestionHandler: %v", r)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Internal server error"})
        }
    }()

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

    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Authentication token missing"})
        return
    }

    questionID, err := services.SaveAssessmentQuestion(r.Context(), question, firebaseToken)
    if err != nil {
        log.Printf("Error saving assessment question: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to save assessment question: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"questionID": questionID})
}

// GetAssessmentQuestionsHandler retrieves assessment questions by type
func GetAssessmentQuestionsHandler(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in GetAssessmentQuestionsHandler: %v", r)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Internal server error"})
        }
    }()

    w.Header().Set("Content-Type", "application/json")

    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Authentication token missing"})
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
        qs, err := services.GetAssessmentQuestions(r.Context(), t, firebaseToken)
        if err != nil {
            log.Printf("Error retrieving assessment questions for type %s: %v", t, err)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve questions: " + err.Error()})
            return
        }
        questions = append(questions, qs...)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(questions)
}

// GetAssessmentResultsHandler retrieves the user's assessment results
func GetAssessmentResultsHandler(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in GetAssessmentResultsHandler: %v", r)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Internal server error"})
        }
    }()

    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value("userID").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }
    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Authentication token missing"})
        return
    }

    results, err := services.GetAssessmentResults(r.Context(), userID, firebaseToken)
    if err != nil {
        log.Printf("Error retrieving assessment results: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve results: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(results)
}

// SubmitAnswerHandler handles answer submissions for different question types
func SubmitAnswerHandler(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in SubmitAnswerHandler: %v", r)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Internal server error"})
        }
    }()

    w.Header().Set("Content-Type", "application/json")

    var submission struct {
        Submissions []models.AssessmentSubmission `json:"submissions"`
    }
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if len(submission.Submissions) == 0 {
        log.Printf("Received empty submissions")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Submissions cannot be empty"})
        return
    }

    userID, ok := r.Context().Value("userID").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }
    firebaseToken, ok := r.Context().Value("firebaseToken").(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Authentication token missing"})
        return
    }

    firestoreClient, err := config.App.Firestore(r.Context())
    if err != nil {
        log.Printf("Failed to connect to Firestore: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to connect to Firestore"})
        return
    }
    defer firestoreClient.Close()

    totalCorrect := 0
    typeMap := make(map[string]int)
    for _, sub := range submission.Submissions {
        question, err := services.GetQuestionByID(r.Context(), sub.QuestionID)
        if err != nil {
            log.Printf("Failed to fetch question %s: %v", sub.QuestionID, err)
            continue
        }

        result, err := services.SaveAssessmentResult(r.Context(), userID, sub, firebaseToken)
        if err != nil {
            log.Printf("Error submitting answer for question %s: %v", sub.QuestionID, err)
            continue
        }
        totalCorrect += result.CorrectAnswers
        typeMap[question.Type] = typeMap[question.Type] + result.CorrectAnswers
    }

    totalQuestionsByType := make(map[string]int)
    for _, sub := range submission.Submissions {
        question, err := services.GetQuestionByID(r.Context(), sub.QuestionID)
        if err == nil {
            docs, err := firestoreClient.Collection("assessmentQuestions").Where("type", "==", question.Type).Documents(r.Context()).GetAll()
            if err == nil {
                totalQuestionsByType[question.Type] = len(docs)
            }
        }
    }

    resultType := ""
    if len(submission.Submissions) > 0 {
        question, err := services.GetQuestionByID(r.Context(), submission.Submissions[0].QuestionID)
        if err == nil {
            resultType = question.Type
        }
    }

    result := models.AssessmentResult{
        Type:           resultType,
        CorrectAnswers: totalCorrect,
        TotalQuestions: totalQuestionsByType[resultType],
        Status:         "completed",
    }

    log.Printf("Successfully processed %d submissions for userID: %s, type: %s, correct: %d/%d", len(submission.Submissions), userID, resultType, totalCorrect, totalQuestionsByType[resultType])
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "result": result,
    })
}