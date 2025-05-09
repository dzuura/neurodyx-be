package services

import (
    "context"
    "fmt"
    "log"

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

    docRef, _, err := firestoreClient.Collection("therapyQuestions").Add(ctx, map[string]interface{}{
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
        "pathData":        question.PathData,
        "timestamp":       firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save therapy question: %w", err)
    }

    log.Printf("Saved therapy question with ID: %s", docRef.ID)
    return docRef.ID, nil
}

// GetTherapyQuestions retrieves therapy questions by type and category.
func GetTherapyQuestions(ctx context.Context, questionType, category string, userID string) ([]models.TherapyQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    query := firestoreClient.Collection("therapyQuestions").Query
    if questionType != "" {
        query = query.Where("type", "==", questionType)
    }
    if category != "" {
        query = query.Where("category", "==", category)
    }

    docs, err := query.Documents(ctx).GetAll()
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
        if pathData, ok := data["pathData"].([]interface{}); ok {
            q.PathData = make([]models.Point, len(pathData))
            for i, p := range pathData {
                if pointMap, ok := p.(map[string]interface{}); ok {
                    x, xOk := pointMap["x"]
                    y, yOk := pointMap["y"]
                    if xOk && yOk {
                        q.PathData[i] = models.Point{
                            X: int(x.(float64)),
                            Y: int(y.(float64)),
                        }
                    } else {
                        q.PathData[i] = models.Point{X: 0, Y: 0}
                    }
                }
            }
        } else {
            q.PathData = []models.Point{}
        }
        questions = append(questions, q)
    }

    log.Printf("Retrieved %d therapy questions for type: %s, category: %s", len(questions), questionType, category)
    return questions, nil
}

// GetTherapyQuestionByID retrieves a therapy question by its ID.
func GetTherapyQuestionByID(ctx context.Context, questionID string) (models.TherapyQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.TherapyQuestion{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    doc, err := firestoreClient.Collection("therapyQuestions").Doc(questionID).Get(ctx)
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
    if pathData, ok := data["pathData"].([]interface{}); ok {
        q.PathData = make([]models.Point, len(pathData))
        for i, p := range pathData {
            if pointMap, ok := p.(map[string]interface{}); ok {
                x, xOk := pointMap["x"]
                y, yOk := pointMap["y"]
                if xOk && yOk {
                    q.PathData[i] = models.Point{
                        X: int(x.(float64)),
                        Y: int(y.(float64)),
                    }
                } else {
                    q.PathData[i] = models.Point{X: 0, Y: 0}
                }
            }
        }
    } else {
        q.PathData = []models.Point{}
    }

    log.Printf("Retrieved therapy question with ID: %s", questionID)
    return q, nil
}

// GetTherapyCategories retrieves all available categories for a given type with descriptions.
func GetTherapyCategories(ctx context.Context, questionType string) ([]models.TherapyCategory, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docs, err := firestoreClient.Collection("therapyQuestions").Where("type", "==", questionType).Documents(ctx).GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve therapy questions: %w", err)
    }

    categories := make(map[string]models.TherapyCategory)
    for _, doc := range docs {
        data := doc.Data()
        if category, ok := data["category"].(string); ok {
            if description, ok := data["description"].(string); ok {
                if _, exists := categories[category]; !exists {
                    categories[category] = models.TherapyCategory{
                        Category:    category,
                        Description: description,
                    }
                }
            }
        }
    }

    categoryList := make([]models.TherapyCategory, 0, len(categories))
    for _, cat := range categories {
        categoryList = append(categoryList, cat)
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

    question, err := GetTherapyQuestionByID(ctx, submission.QuestionID)
    if err != nil {
        return models.TherapyResult{}, fmt.Errorf("failed to fetch question: %w", err)
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
    docs, err := firestoreClient.Collection("therapyQuestions").Where("type", "==", questionType).Where("category", "==", category).Documents(ctx).GetAll()
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