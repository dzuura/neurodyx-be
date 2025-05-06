package services

import (
    "context"
    "fmt"
    "log"

    "cloud.google.com/go/firestore"
    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/models"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// SaveScreeningQuestion saves a new screening question to Firestore
func SaveScreeningQuestion(ctx context.Context, question models.ScreeningQuestion, userID string) (string, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef, _, err := firestoreClient.Collection("screeningQuestions").Add(ctx, map[string]interface{}{
        "ageGroup":  question.AgeGroup,
        "question":  question.Question,
        "timestamp": firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save screening question: %w", err)
    }

    log.Printf("Saved screening question with ID: %s", docRef.ID)
    return docRef.ID, nil
}

// GetScreeningQuestions retrieves screening questions, optionally filtered by ageGroup
func GetScreeningQuestions(ctx context.Context, ageGroup string, userID string) ([]models.ScreeningQuestion, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    var docs []*firestore.DocumentSnapshot
    if ageGroup == "" {
        docs, err = firestoreClient.Collection("screeningQuestions").Documents(ctx).GetAll()
    } else {
        docs, err = firestoreClient.Collection("screeningQuestions").Where("ageGroup", "==", ageGroup).Documents(ctx).GetAll()
    }
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve screening questions: %w", err)
    }

    questions := make([]models.ScreeningQuestion, 0, len(docs))
    for _, doc := range docs {
        var q models.ScreeningQuestion
        data := doc.Data()
        q.AgeGroup = data["ageGroup"].(string)
        q.Question = data["question"].(string)
        questions = append(questions, q)
    }

    log.Printf("Retrieved %d screening questions for ageGroup: %s", len(questions), ageGroup)
    return questions, nil
}

// SaveScreeningResult saves the user's screening submission and risk level
func SaveScreeningResult(ctx context.Context, userID, ageGroup string, answers []bool, riskLevel string, firebaseToken string) error {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    _, _, err = firestoreClient.Collection("users").Doc(userID).Collection("screenings").Add(ctx, map[string]interface{}{
        "ageGroup":  ageGroup,
        "answers":   answers,
        "riskLevel": riskLevel,
        "timestamp": firestore.ServerTimestamp,
    })
    if err != nil {
        return fmt.Errorf("failed to save screening result: %w", err)
    }

    log.Printf("Saved screening result for userID: %s, ageGroup: %s", userID, ageGroup)
    return nil
}

// GetQuestionByID retrieves a question by its ID
func GetQuestionByID(ctx context.Context, questionID string) (models.AssessmentQuestion, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return models.AssessmentQuestion{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    doc, err := firestoreClient.Collection("assessmentQuestions").Doc(questionID).Get(ctx)
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

    log.Printf("Retrieved question with ID: %s", questionID)
    return q, nil
}

// SaveAssessmentQuestion saves a new assessment question to Firestore
func SaveAssessmentQuestion(ctx context.Context, question models.AssessmentQuestion, userID string) (string, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef, _, err := firestoreClient.Collection("assessmentQuestions").Add(ctx, map[string]interface{}{
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
        "pathData":        question.PathData,
        "timestamp":       firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save assessment question: %w", err)
    }

    log.Printf("Saved assessment question with ID: %s", docRef.ID)
    return docRef.ID, nil
}

// GetAssessmentQuestions retrieves assessment questions by type
func GetAssessmentQuestions(ctx context.Context, questionType string, userID string) ([]models.AssessmentQuestion, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    var docs []*firestore.DocumentSnapshot
    if questionType != "" {
        docs, err = firestoreClient.Collection("assessmentQuestions").Where("type", "==", questionType).Documents(ctx).GetAll()
    } else {
        docs, err = firestoreClient.Collection("assessmentQuestions").Documents(ctx).GetAll()
    }
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve assessment questions: %w", err)
    }

    questions := make([]models.AssessmentQuestion, 0, len(docs))
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

    log.Printf("Retrieved %d assessment questions for type: %s", len(questions), questionType)
    return questions, nil
}

// SaveAssessmentResult saves the user's assessment result with flexible answer validation
func SaveAssessmentResult(ctx context.Context, userID string, submission models.AssessmentSubmission, firebaseToken string) (models.AssessmentResult, error) {
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return models.AssessmentResult{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    question, err := GetQuestionByID(ctx, submission.QuestionID)
    if err != nil {
        return models.AssessmentResult{}, fmt.Errorf("failed to fetch question: %w", err)
    }

    isCorrect := false
    switch question.Category {
    case "word_repetition":
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
    case "complete_the_word_by_touch", "word_recognition_by_touch":
        answerStr, ok := submission.Answer.(string)
        if ok {
            isCorrect = ValidateStringMatch(answerStr, question.CorrectAnswer)
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

// GetAssessmentResults retrieves all assessment results for a user
func GetAssessmentResults(ctx context.Context, userID string, firebaseToken string) ([]models.AssessmentResult, error) {
    firestoreClient, err := config.App.Firestore(ctx)
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
        docs, err := firestoreClient.Collection("assessmentQuestions").Where("type", "==", t).Documents(ctx).GetAll()
        if err == nil {
            totalQuestions[t] = len(docs)
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