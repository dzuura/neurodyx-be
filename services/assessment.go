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

// GetQuestionByID retrieves a question by its ID, type, and category.
func GetQuestionByID(ctx context.Context, questionType, category, questionID string) (models.AssessmentQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.AssessmentQuestion{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    doc, err := firestoreClient.Collection("assessmentQuestions").Doc(questionType).Collection(category).Doc(questionID).Get(ctx)
    if err != nil {
        return models.AssessmentQuestion{}, fmt.Errorf("failed to retrieve question: %w", err)
    }

    var q models.AssessmentQuestion
    data := doc.Data()
    q.ID = doc.Ref.ID
    q.Type = data["type"].(string)
    q.Category = data["category"].(string)
    if content, ok := data["content"].(string); ok {
        q.Content = content
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

    log.Printf("Retrieved question with ID: %s, type: %s, category: %s", questionID, questionType, category)
    return q, nil
}

// SaveAssessmentQuestion saves a new assessment question to Firestore.
func SaveAssessmentQuestion(ctx context.Context, question models.AssessmentQuestion, userID string) (string, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef, _, err := firestoreClient.Collection("assessmentQuestions").Doc(question.Type).Collection(question.Category).Add(ctx, map[string]interface{}{
        "type":            question.Type,
        "category":        question.Category,
        "content":         question.Content,
        "imageURL":        question.ImageURL,
        "soundURL":        question.SoundURL,
        "options":         question.Options,
        "leftItems":       question.LeftItems,
        "rightItems":      question.RightItems,
        "correctAnswer":   question.CorrectAnswer,
        "correctSequence": question.CorrectSequence,
        "correctPairs":    question.CorrectPairs,
        "timestamp":       firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save assessment question: %w", err)
    }

    log.Printf("Saved assessment question with ID: %s, type: %s, category: %s", docRef.ID, question.Type, question.Category)
    return docRef.ID, nil
}

// GetAssessmentQuestions retrieves assessment questions by type.
func GetAssessmentQuestions(ctx context.Context, questionType string, userID string) ([]models.AssessmentQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    questions := make([]models.AssessmentQuestion, 0)

    categoriesIter := firestoreClient.Collection("assessmentQuestions").Doc(questionType).Collections(ctx)
    categories, err := categoriesIter.GetAll()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve categories for type %s: %w", questionType, err)
    }

    for _, category := range categories {
        categoryName := category.ID
        docs, err := category.Documents(ctx).GetAll()
        if err != nil {
            log.Printf("Failed to retrieve questions for type %s, category %s: %v", questionType, categoryName, err)
            continue
        }

        for _, doc := range docs {
            var q models.AssessmentQuestion
            data := doc.Data()
            q.ID = doc.Ref.ID
            q.Type = data["type"].(string)
            q.Category = data["category"].(string)
            if content, ok := data["content"].(string); ok {
                q.Content = content
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
            if leftItems, ok := data["leftItems"].([]interface{}); ok {
                q.LeftItems = make([]string, len(leftItems))
                for i, item := range leftItems {
                    q.LeftItems[i] = item.(string)
                }
            }
            if rightItems, ok := data["rightItems"].([]interface{}); ok {
                q.RightItems = make([]string, len(rightItems))
                for i, item := range rightItems {
                    q.RightItems[i] = item.(string)
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
    }

    log.Printf("Retrieved %d assessment questions for type: %s", len(questions), questionType)
    return questions, nil
}

// SaveAssessmentResult saves the user's assessment result with flexible answer validation.
func SaveAssessmentResult(ctx context.Context, userID string, submission models.AssessmentSubmission, firebaseToken string) (models.AssessmentResult, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.AssessmentResult{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    var question models.AssessmentQuestion
    types := []string{"visual", "auditory", "kinesthetic", "tactile"}
    for _, t := range types {
        categoriesIter := firestoreClient.Collection("assessmentQuestions").Doc(t).Collections(ctx)
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
                if leftItems, ok := data["leftItems"].([]interface{}); ok {
                    question.LeftItems = make([]string, len(leftItems))
                    for i, item := range leftItems {
                        question.LeftItems[i] = item.(string)
                    }
                }
                if rightItems, ok := data["rightItems"].([]interface{}); ok {
                    question.RightItems = make([]string, len(rightItems))
                    for i, item := range rightItems {
                        question.RightItems[i] = item.(string)
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
        return models.AssessmentResult{}, fmt.Errorf("question with ID %s not found", submission.QuestionID)
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

    result := models.AssessmentResult{
        Type:           question.Type,
        CorrectAnswers: 1,
        TotalQuestions: 1,
    }
    if !isCorrect {
        result.CorrectAnswers = 0
    }

    _, err = firestoreClient.Collection("users").Doc(userID).Collection("assessments").Doc(question.Type).Collection("submissions").Doc(submission.QuestionID).Set(ctx, map[string]interface{}{
        "type":           question.Type,
        "category":       question.Category,
        "questionID":     submission.QuestionID,
        "correctAnswers": result.CorrectAnswers,
        "answer":         submission.Answer,
        "status":         "completed",
        "timestamp":      firestore.ServerTimestamp,
    }, firestore.MergeAll)
    if err != nil {
        return models.AssessmentResult{}, fmt.Errorf("failed to save assessment result: %w", err)
    }

    log.Printf("Saved assessment result for userID: %s, questionID: %s, type: %s, isCorrect: %v", userID, submission.QuestionID, question.Type, isCorrect)
    return result, nil
}

// GetAssessmentResults retrieves all assessment results for a user.
func GetAssessmentResults(ctx context.Context, userID string, firebaseToken string) ([]models.AssessmentResult, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    totalQuestions := map[string]int{
        "visual":      0,
        "auditory":    0,
        "kinesthetic": 0,
        "tactile":     0,
    }
    for _, t := range []string{"visual", "auditory", "kinesthetic", "tactile"} {
        categoriesIter := firestoreClient.Collection("assessmentQuestions").Doc(t).Collections(ctx)
        categories, err := categoriesIter.GetAll()
        if err != nil {
            continue
        }
        for _, category := range categories {
            docs, err := category.Documents(ctx).GetAll()
            if err == nil {
                totalQuestions[t] += len(docs)
            }
        }
    }

    results := []models.AssessmentResult{
        {Type: "visual", CorrectAnswers: 0, TotalQuestions: totalQuestions["visual"], Status: "not started"},
        {Type: "auditory", CorrectAnswers: 0, TotalQuestions: totalQuestions["auditory"], Status: "not started"},
        {Type: "kinesthetic", CorrectAnswers: 0, TotalQuestions: totalQuestions["kinesthetic"], Status: "not started"},
        {Type: "tactile", CorrectAnswers: 0, TotalQuestions: totalQuestions["tactile"], Status: "not started"},
    }

    for _, result := range results {
        typeName := result.Type
        submissionDocs, err := firestoreClient.Collection("users").Doc(userID).Collection("assessments").Doc(typeName).Collection("submissions").Documents(ctx).GetAll()
        if err != nil {
            if status.Code(err) == codes.NotFound {
                log.Printf("No submissions found for type %s for userID: %s", typeName, userID)
                continue
            }
            log.Printf("Failed to fetch submissions for type %s: %v", typeName, err)
            continue
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

        for i, r := range results {
            if r.Type == typeName {
                results[i].CorrectAnswers = correctAnswers
                if len(submissionDocs) > 0 {
                    results[i].Status = "completed"
                }
                break
            }
        }
    }

    log.Printf("Retrieved %d assessment results for userID: %s", len(results), userID)
    return results, nil
}