package services

import (
    "context"
    "fmt"
    "log"
    "time"

    "cloud.google.com/go/firestore"
    "github.com/dzuura/neurodyx-be/models"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// SaveTherapyQuestion saves a new therapy question to Firestore.
func SaveTherapyQuestion(ctx context.Context, question models.TherapyQuestion, userID string) (string, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef, _, err := firestoreClient.Collection("therapyQuestions").Doc(question.Type).Collection(question.Category).Add(ctx, map[string]interface{}{
        "type":            question.Type,
        "category":        question.Category,
        "content":         question.Content,
        "description":     question.Description,
        "imageURL":        question.ImageURL,
        "soundURL":        question.SoundURL,
        "options":         question.Options,
        "correctAnswer":   question.CorrectAnswer,
        "correctSequence": question.CorrectSequence,
        "correctPairs":    question.CorrectPairs,
        "timestamp":       firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save therapy question: %w", err)
    }

    log.Printf("Saved therapy question with ID: %s, type: %s, category: %s", docRef.ID, question.Type, question.Category)
    return docRef.ID, nil
}

// GetTherapyQuestions retrieves therapy questions by type and category.
func GetTherapyQuestions(ctx context.Context, questionType, category string, userID string) ([]models.TherapyQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docs, err := firestoreClient.Collection("therapyQuestions").Doc(questionType).Collection(category).Documents(ctx).GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve therapy questions: %w", err)
    }

    questions := make([]models.TherapyQuestion, 0, len(docs))
    for _, doc := range docs {
        var q models.TherapyQuestion
        data := doc.Data()
        q.ID = doc.Ref.ID
        q.Type = data["type"].(string)
        q.Category = data["category"].(string)
        if content, ok := data["content"].(string); ok {
            q.Content = content
        }
        if description, ok := data["description"].(string); ok {
            q.Description = description
        }
        if imageURL, ok := data["imageURL"].(string); ok {
            q.ImageURL = imageURL
        }
        if soundURL, ok := data["soundURL"].(string); ok {
            q.SoundURL = soundURL
        }
        if options, ok := data["options"].([]interface{}); ok {
            q.Options = make([]string, len(options))
            for i, opt := range options {
                q.Options[i] = opt.(string)
            }
        }
        if correctAnswer, ok := data["correctAnswer"].(string); ok {
            q.CorrectAnswer = correctAnswer
        }
        if correctSeq, ok := data["correctSequence"].([]interface{}); ok {
            q.CorrectSequence = make([]string, len(correctSeq))
            for i, seq := range correctSeq {
                q.CorrectSequence[i] = seq.(string)
            }
        }
        if correctPairs, ok := data["correctPairs"].(map[string]interface{}); ok {
            q.CorrectPairs = make(map[string]string)
            for k, v := range correctPairs {
                q.CorrectPairs[k] = v.(string)
            }
        }
        questions = append(questions, q)
    }

    log.Printf("Retrieved %d therapy questions for type: %s, category: %s", len(questions), questionType, category)
    return questions, nil
}

// GetTherapyQuestionByID retrieves a therapy question by its ID.
func GetTherapyQuestionByID(ctx context.Context, questionType, category, questionID string) (models.TherapyQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.TherapyQuestion{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    doc, err := firestoreClient.Collection("therapyQuestions").Doc(questionType).Collection(category).Doc(questionID).Get(ctx)
    if err != nil {
        return models.TherapyQuestion{}, fmt.Errorf("failed to retrieve therapy question: %w", err)
    }

    var q models.TherapyQuestion
    data := doc.Data()
    q.ID = doc.Ref.ID
    q.Type = data["type"].(string)
    q.Category = data["category"].(string)
    if content, ok := data["content"].(string); ok {
        q.Content = content
    }
    if description, ok := data["description"].(string); ok {
        q.Description = description
    }
    if imageURL, ok := data["imageURL"].(string); ok {
        q.ImageURL = imageURL
    }
    if soundURL, ok := data["soundURL"].(string); ok {
        q.SoundURL = soundURL
    }
    if options, ok := data["options"].([]interface{}); ok {
        q.Options = make([]string, len(options))
        for i, opt := range options {
            q.Options[i] = opt.(string)
        }
    }
    if correctAnswer, ok := data["correctAnswer"].(string); ok {
        q.CorrectAnswer = correctAnswer
    }
    if correctSeq, ok := data["correctSequence"].([]interface{}); ok {
        q.CorrectSequence = make([]string, len(correctSeq))
        for i, seq := range correctSeq {
            q.CorrectSequence[i] = seq.(string)
        }
    }
    if correctPairs, ok := data["correctPairs"].(map[string]interface{}); ok {
        q.CorrectPairs = make(map[string]string)
        for k, v := range correctPairs {
            q.CorrectPairs[k] = v.(string)
        }
    }

    log.Printf("Retrieved therapy question with ID: %s, type: %s, category: %s", questionID, questionType, category)
    return q, nil
}

// GetTherapyCategories retrieves all available categories for a given type with descriptions.
func GetTherapyCategories(ctx context.Context, questionType string) ([]models.TherapyCategory, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    categoriesIter := firestoreClient.Collection("therapyQuestions").Doc(questionType).Collections(ctx)
    categories, err := categoriesIter.GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve categories for type %s: %w", questionType, err)
    }

    categoryList := make([]models.TherapyCategory, 0)
    for _, category := range categories {
        categoryName := category.ID
        docs, err := category.Limit(1).Documents(ctx).GetAll()
        if err != nil || len(docs) == 0 {
            continue
        }
        data := docs[0].Data()
        if description, ok := data["description"].(string); ok {
            categoryList = append(categoryList, models.TherapyCategory{
                Category:    categoryName,
                Description: description,
            })
        }
    }

    log.Printf("Retrieved %d categories for type: %s", len(categoryList), questionType)
    return categoryList, nil
}

// SaveTherapyResult saves the user's therapy result with flexible answer validation.
func SaveTherapyResult(ctx context.Context, userID string, submission models.TherapySubmission, firebaseToken string) (models.TherapyResult, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.TherapyResult{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    var question models.TherapyQuestion
    types := []string{"visual", "auditory", "kinesthetic", "tactile"}
    for _, t := range types {
        categoriesIter := firestoreClient.Collection("therapyQuestions").Doc(t).Collections(ctx)
        categories, err := categoriesIter.GetAll()
        if err != nil {
            continue
        }
        for _, category := range categories {
            doc, err := category.Doc(submission.QuestionID).Get(ctx)
            if err == nil && doc.Exists() {
                data := doc.Data()
                question.ID = doc.Ref.ID
                question.Type = data["type"].(string)
                question.Category = data["category"].(string)
                if options, ok := data["options"].([]interface{}); ok {
                    question.Options = make([]string, len(options))
                    for i, opt := range options {
                        question.Options[i] = opt.(string)
                    }
                }
                if correctAnswer, ok := data["correctAnswer"].(string); ok {
                    question.CorrectAnswer = correctAnswer
                }
                if correctSeq, ok := data["correctSequence"].([]interface{}); ok {
                    question.CorrectSequence = make([]string, len(correctSeq))
                    for i, seq := range correctSeq {
                        question.CorrectSequence[i] = seq.(string)
                    }
                }
                if correctPairs, ok := data["correctPairs"].(map[string]interface{}); ok {
                    question.CorrectPairs = make(map[string]string)
                    for k, v := range correctPairs {
                        question.CorrectPairs[k] = v.(string)
                    }
                }
                break
            }
        }
        if question.ID != "" {
            break
        }
    }

    if question.ID == "" {
        return models.TherapyResult{}, fmt.Errorf("therapy question with ID %s not found", submission.QuestionID)
    }

    isCorrect := false
    switch question.Category {
    case "word_repetition", "word_recognition_by_touch", "complete_the_word_by_touch":
        answerStr, ok := submission.Answer.(string)
        if ok {
            isCorrect = ValidateStringMatch(answerStr, question.CorrectAnswer)
        }
    case "letter_matching":
        answerSeq, ok := submission.Answer.([]interface{})
        if ok {
            seq := make([]string, len(answerSeq))
            for i, v := range answerSeq {
                seq[i], _ = v.(string)
            }
            isCorrect = ValidateSequence(seq, question.CorrectSequence)
        }
    case "number_letter_similarity":
        answerPairsRaw, ok := submission.Answer.([]interface{})
        if ok {
            answerPairs := make([]map[string]string, len(answerPairsRaw))
            for i, pairRaw := range answerPairsRaw {
                pairMap, pairOk := pairRaw.(map[string]interface{})
                if !pairOk {
                    log.Printf("Invalid pair format for questionID: %s", submission.QuestionID)
                    break
                }
                left, leftOk := pairMap["left"].(string)
                right, rightOk := pairMap["right"].(string)
                if !leftOk || !rightOk {
                    log.Printf("Invalid left or right value for questionID: %s", submission.QuestionID)
                    break
                }
                answerPairs[i] = map[string]string{
                    "left":  left,
                    "right": right,
                }
            }
            isCorrect = ValidatePairs(answerPairs, question.CorrectPairs)
        } else {
            log.Printf("Expected []interface{} for number_letter_similarity, got %T for questionID: %s", submission.Answer, submission.QuestionID)
        }
    default:
        answerStr, ok := submission.Answer.(string)
        if ok && len(question.Options) > 0 {
            for _, opt := range question.Options {
                if opt == answerStr {
                    isCorrect = true
                    break
                }
            }
        }
    }

    result := models.TherapyResult{
        Type:           question.Type,
        Category:       question.Category,
        CorrectAnswers: 1,
        TotalQuestions: 1,
    }
    if !isCorrect {
        result.CorrectAnswers = 0
    }

    _, err = firestoreClient.Collection("users").Doc(userID).Collection("therapy").Doc(question.Type).Collection(question.Category).Doc(submission.QuestionID).Set(ctx, map[string]interface{}{
        "type":           question.Type,
        "category":       question.Category,
        "questionID":     submission.QuestionID,
        "correctAnswers": result.CorrectAnswers,
        "answer":         submission.Answer,
        "status":         "completed",
        "timestamp":      firestore.ServerTimestamp,
    }, firestore.MergeAll)
    if err != nil {
        return models.TherapyResult{}, fmt.Errorf("failed to save therapy result: %w", err)
    }

    log.Printf("Saved therapy result for userID: %s, questionID: %s, type: %s, category: %s, isCorrect: %v", userID, submission.QuestionID, question.Type, question.Category, isCorrect)
    return result, nil
}

// GetTherapyResults retrieves therapy results for a user by type and category.
func GetTherapyResults(ctx context.Context, userID, questionType, category string) (models.TherapyResult, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.TherapyResult{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    totalQuestions := 0
    docs, err := firestoreClient.Collection("therapyQuestions").Doc(questionType).Collection(category).Documents(ctx).GetAll()
    if err == nil {
        totalQuestions = len(docs)
    }

    result := models.TherapyResult{
        Type:           questionType,
        Category:       category,
        CorrectAnswers: 0,
        TotalQuestions: totalQuestions,
        Status:         "not started",
    }

    submissionDocs, err := firestoreClient.Collection("users").Doc(userID).Collection("therapy").Doc(questionType).Collection(category).Documents(ctx).GetAll()
    if err != nil {
        if status.Code(err) == codes.NotFound {
            log.Printf("No submissions found for type %s, category %s for userID: %s", questionType, category, userID)
            return result, nil
        }
        return result, fmt.Errorf("failed to fetch submissions: %w", err)
    }

    correctAnswers := 0
    for _, subDoc := range submissionDocs {
        data := subDoc.Data()
        if correct, ok := data["correctAnswers"].(int64); ok {
            correctAnswers += int(correct)
        } else if correct, ok := data["correctAnswers"].(int); ok {
            correctAnswers += correct
        }
    }

    result.CorrectAnswers = correctAnswers
    if len(submissionDocs) > 0 {
        result.Status = "completed"
    }

    log.Printf("Retrieved therapy result for userID: %s, type: %s, category: %s, correct: %d/%d", userID, questionType, category, result.CorrectAnswers, result.TotalQuestions)
    return result, nil
}

// UpdateDailyProgress updates the user's daily progress for therapy activities.
func UpdateDailyProgress(ctx context.Context, userID, questionType, category string) (*models.DailyProgress, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    now := time.Now().UTC()
    date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    docID := date.Format("20060102")
    docRef := firestoreClient.Collection("users").Doc(userID).Collection("progress").Doc(docID)

    var progress models.DailyProgress
    err = firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        doc, err := tx.Get(docRef)
        if err != nil && !doc.Exists() {
            progress = models.DailyProgress{
                UserID:         userID,
                Date:           date,
                TherapyCount:   1,
                StreakAchieved: false,
            }
        } else if err != nil {
            return fmt.Errorf("failed to fetch progress: %w", err)
        } else {
            if err := doc.DataTo(&progress); err != nil {
                return fmt.Errorf("failed to parse progress data: %w", err)
            }
            progress.TherapyCount++
        }

        progress.StreakAchieved = progress.TherapyCount >= 5

        progressData := map[string]interface{}{
            "userID":         progress.UserID,
            "date":           progress.Date,
            "therapyCount":   progress.TherapyCount,
            "streakAchieved": progress.StreakAchieved,
        }
        return tx.Set(docRef, progressData)
    })
    if err != nil {
        return nil, fmt.Errorf("failed to update daily progress: %w", err)
    }

    log.Printf("Updated daily progress for userID: %s on %s, therapyCount: %d, streak: %v", userID, date.Format("2006-01-02"), progress.TherapyCount, progress.StreakAchieved)
    return &progress, nil
}

// GetWeeklyProgress retrieves the user's progress for the last 7 days.
func GetWeeklyProgress(ctx context.Context, userID string) ([]models.DailyProgress, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    now := time.Now().UTC()
    endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    startDate := endDate.AddDate(0, 0, -6)

    docs, err := firestoreClient.Collection("users").Doc(userID).Collection("progress").
        Where("date", ">=", startDate).
        Where("date", "<=", endDate).
        Documents(ctx).GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve weekly progress: %w", err)
    }

    progressMap := make(map[string]models.DailyProgress)
    for _, doc := range docs {
        var p models.DailyProgress
        if err := doc.DataTo(&p); err != nil {
            log.Printf("Failed to parse progress data for doc %s: %v", doc.Ref.ID, err)
            continue
        }
        progressMap[p.Date.Format("20060102")] = p
    }

    result := make([]models.DailyProgress, 0, 7)
    for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
        docID := d.Format("20060102")
        if p, exists := progressMap[docID]; exists {
            result = append(result, p)
        } else {
            result = append(result, models.DailyProgress{
                UserID:         userID,
                Date:           d,
                TherapyCount:   0,
                StreakAchieved: false,
            })
        }
    }

    log.Printf("Retrieved weekly progress for userID: %s, entries: %d", userID, len(result))
    return result, nil
}

// GetMonthlyProgress retrieves the user's progress for a specific month and year.
func GetMonthlyProgress(ctx context.Context, userID string, year, month int) ([]models.ProgressDetail, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)

    docs, err := firestoreClient.Collection("users").Doc(userID).Collection("progress").
        Where("date", ">=", startDate).
        Where("date", "<=", endDate).
        Documents(ctx).GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve monthly progress: %w", err)
    }

    progressMap := make(map[string]models.DailyProgress)
    for _, doc := range docs {
        var p models.DailyProgress
        if err := doc.DataTo(&p); err != nil {
            log.Printf("Failed to parse progress data for doc %s: %v", doc.Ref.ID, err)
            continue
        }
        progressMap[p.Date.Format("20060102")] = p
    }

    result := make([]models.ProgressDetail, 0, endDate.Day())
    for day := 1; day <= endDate.Day(); day++ {
        date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
        docID := date.Format("20060102")
        status := "inactive"

        if p, exists := progressMap[docID]; exists {
            if p.StreakAchieved {
                status = "streak"
            } else if p.TherapyCount > 0 {
                status = "active"
            }
        }

        result = append(result, models.ProgressDetail{
            Date:   date,
            Status: status,
        })
    }

    log.Printf("Retrieved monthly progress for userID: %s, year: %d, month: %d, entries: %d", userID, year, month, len(result))
    return result, nil
}