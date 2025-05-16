package handlers

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/mux"
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

// GetTherapyQuestionByIDHandler retrieves a therapy question by ID.
func UpdateTherapyQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionID := mux.Vars(r)["questionID"]
    if questionID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required parameter: questionID"})
        return
    }

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

    firestoreClient, err := config.App.Firestore(r.Context())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to connect to Firestore"})
        return
    }
    defer firestoreClient.Close()

    var originalType, originalCategory string
    types := []string{"visual", "auditory", "kinesthetic", "tactile"}
    found := false
    for _, t := range types {
        categoriesIter := firestoreClient.Collection("therapyQuestions").Doc(t).Collections(r.Context())
        categories, err := categoriesIter.GetAll()
        if err != nil {
            log.Printf("Failed to retrieve categories for type %s: %v", t, err)
            continue
        }
        for _, category := range categories {
            doc, err := category.Doc(questionID).Get(r.Context())
            if err == nil && doc.Exists() {
                originalType = t
                originalCategory = category.ID
                found = true
                break
            }
        }
        if found {
            break
        }
    }

    if !found {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Therapy question not found"})
        return
    }

    if question.Type != originalType || question.Category != originalCategory {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Cannot change type or category. Update must be within the original type: " + originalType + " and category: " + originalCategory})
        return
    }

    err = services.UpdateTherapyQuestion(r.Context(), questionID, question, userID)
    if err != nil {
        log.Printf("Error updating therapy question %s: %v", questionID, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to update therapy question: " + err.Error()})
        return
    }

    updatedQuestion, err := services.GetTherapyQuestionByID(r.Context(), question.Type, question.Category, questionID)
    if err != nil {
        log.Printf("Error retrieving updated therapy question %s: %v", questionID, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve updated question: " + err.Error()})
        return
    }

    cacheKey := question.Type + ":" + question.Category
    config.TherapyQuestionCache.Delete(cacheKey)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(updatedQuestion)
}

// DeleteTherapyQuestionHandler deletes a therapy question by ID.
func DeleteTherapyQuestionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    questionID := mux.Vars(r)["questionID"]
    if questionID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Missing required parameter: questionID"})
        return
    }

    firestoreClient, err := config.App.Firestore(r.Context())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to connect to Firestore"})
        return
    }
    defer firestoreClient.Close()

    var questionType, category string
    types := []string{"visual", "auditory", "kinesthetic", "tactile"}
    found := false
    for _, t := range types {
        categoriesIter := firestoreClient.Collection("therapyQuestions").Doc(t).Collections(r.Context())
        categories, err := categoriesIter.GetAll()
        if err != nil {
            log.Printf("Failed to retrieve categories for type %s: %v", t, err)
            continue
        }
        for _, cat := range categories {
            doc, err := cat.Doc(questionID).Get(r.Context())
            if err == nil && doc.Exists() {
                questionType = t
                category = cat.ID
                found = true
                break
            }
        }
        if found {
            break
        }
    }

    if !found {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Therapy question not found"})
        return
    }

    err = services.DeleteTherapyQuestion(r.Context(), questionID, questionType, category, userID)
    if err != nil {
        log.Printf("Error deleting therapy question %s: %v", questionID, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to delete therapy question: " + err.Error()})
        return
    }

    cacheKey := questionType + ":" + category
    config.TherapyQuestionCache.Delete(cacheKey)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Therapy question deleted successfully"})
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

    _, err := services.UpdateDailyProgress(r.Context(), userID, submission.Type, submission.Category)
    if err != nil {
        log.Printf("Error updating daily progress for userID %s: %v", userID, err)
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