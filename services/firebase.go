package services

import (
    "context"
    "fmt"
    "log"

    "cloud.google.com/go/firestore"
    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/models"
)

// SaveScreeningQuestion saves a new screening question to Firestore
func SaveScreeningQuestion(ctx context.Context, question models.ScreeningQuestion, firebaseToken string) (string, error) {
    // Verify Firebase token
    authClient, err := config.App.Auth(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to initialize auth client: %w", err)
    }

    _, err = authClient.VerifyIDToken(ctx, firebaseToken)
    if err != nil {
        return "", fmt.Errorf("invalid Firebase token: %w", err)
    }

    // Connect to Firestore
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    // Save the question
    docRef, _, err := firestoreClient.Collection("screeningQuestions").Add(ctx, map[string]interface{}{
        "ageGroup": question.AgeGroup,
        "question": question.Question,
        "timestamp": firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save screening question: %w", err)
    }

    log.Printf("Saved screening question with ID: %s", docRef.ID)
    return docRef.ID, nil
}

// GetScreeningQuestions retrieves screening questions, optionally filtered by ageGroup
func GetScreeningQuestions(ctx context.Context, ageGroup string, firebaseToken string) ([]models.ScreeningQuestion, error) {
    // Verify Firebase token
    authClient, err := config.App.Auth(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize auth client: %w", err)
    }

    _, err = authClient.VerifyIDToken(ctx, firebaseToken)
    if err != nil {
        return nil, fmt.Errorf("invalid Firebase token: %w", err)
    }

    // Connect to Firestore
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    // Fetch questions
    var docs []*firestore.DocumentSnapshot
    if ageGroup == "" {
        docs, err = firestoreClient.Collection("screeningQuestions").Documents(ctx).GetAll()
    } else {
        docs, err = firestoreClient.Collection("screeningQuestions").Where("ageGroup", "==", ageGroup).Documents(ctx).GetAll()
    }
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve screening questions: %w", err)
    }

    // Map documents to ScreeningQuestion
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
    // Verify Firebase token
    authClient, err := config.App.Auth(ctx)
    if err != nil {
        return fmt.Errorf("failed to initialize auth client: %w", err)
    }

    _, err = authClient.VerifyIDToken(ctx, firebaseToken)
    if err != nil {
        return fmt.Errorf("invalid Firebase token: %w", err)
    }

    // Connect to Firestore
    firestoreClient, err := config.App.Firestore(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    // Save the screening result
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