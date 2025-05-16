package services

import (
    "context"
    "fmt"
    "log"

    "cloud.google.com/go/firestore"
    "github.com/dzuura/neurodyx-be/models"
)

// GetScreeningQuestions retrieves screening questions, optionally filtered by ageGroup.
func GetScreeningQuestions(ctx context.Context, ageGroup string, userID string) ([]models.ScreeningQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    var docs []*firestore.DocumentSnapshot
    if ageGroup == "" {
        ageGroupsIter := firestoreClient.Collection("screeningQuestions").DocumentRefs(ctx)
        ageGroups, err := ageGroupsIter.GetAll()
        if err != nil {
            return nil, fmt.Errorf("failed to retrieve age groups: %w", err)
        }
        for _, ageGroupDoc := range ageGroups {
            ageGroupName := ageGroupDoc.ID
            groupDocs, err := firestoreClient.Collection("screeningQuestions").Doc(ageGroupName).Collection("questions").Documents(ctx).GetAll()
            if err != nil {
                log.Printf("Failed to retrieve questions for ageGroup %s: %v", ageGroupName, err)
                continue
            }
            docs = append(docs, groupDocs...)
        }
    } else {
        docs, err = firestoreClient.Collection("screeningQuestions").Doc(ageGroup).Collection("questions").Documents(ctx).GetAll()
        if err != nil {
            return nil, fmt.Errorf("failed to retrieve screening questions for ageGroup %s: %w", ageGroup, err)
        }
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

// GetScreeningQuestionByID retrieves a screening question by ID.
func GetScreeningQuestionByID(ctx context.Context, questionID, ageGroup string) (models.ScreeningQuestion, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return models.ScreeningQuestion{}, fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    doc, err := firestoreClient.Collection("screeningQuestions").Doc(ageGroup).Collection("questions").Doc(questionID).Get(ctx)
    if err != nil {
        return models.ScreeningQuestion{}, fmt.Errorf("failed to retrieve screening question: %w", err)
    }

    var q models.ScreeningQuestion
    data := doc.Data()
    q.ID = doc.Ref.ID
    q.AgeGroup = data["ageGroup"].(string)
    if question, ok := data["question"].(string); ok {
        q.Question = question
    }

    log.Printf("Retrieved screening question with ID: %s, ageGroup: %s", questionID, ageGroup)
    return q, nil
}

// SaveScreeningQuestion saves a new screening question to Firestore.
func SaveScreeningQuestion(ctx context.Context, question models.ScreeningQuestion, userID string) (string, error) {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef, _, err := firestoreClient.Collection("screeningQuestions").Doc(question.AgeGroup).Collection("questions").Add(ctx, map[string]interface{}{
        "ageGroup":  question.AgeGroup,
        "question":  question.Question,
        "timestamp": firestore.ServerTimestamp,
    })
    if err != nil {
        return "", fmt.Errorf("failed to save screening question: %w", err)
    }

    log.Printf("Saved screening question with ID: %s, ageGroup: %s", docRef.ID, question.AgeGroup)
    return docRef.ID, nil
}

// UpdateScreeningQuestion updates an existing screening question in Firestore.
func UpdateScreeningQuestion(ctx context.Context, questionID string, question models.ScreeningQuestion, userID string) error {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    _, err = firestoreClient.Collection("screeningQuestions").Doc(question.AgeGroup).Collection("questions").Doc(questionID).Set(ctx, map[string]interface{}{
        "ageGroup":  question.AgeGroup,
        "question":  question.Question,
        "timestamp": firestore.ServerTimestamp,
    }, firestore.MergeAll)
    if err != nil {
        return fmt.Errorf("failed to update screening question: %w", err)
    }

    log.Printf("Updated screening question with ID: %s, ageGroup: %s", questionID, question.AgeGroup)
    return nil
}

// DeleteScreeningQuestion deletes a screening question from Firestore.
func DeleteScreeningQuestion(ctx context.Context, questionID, ageGroup, userID string) error {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    _, err = firestoreClient.Collection("screeningQuestions").Doc(ageGroup).Collection("questions").Doc(questionID).Delete(ctx)
    if err != nil {
        return fmt.Errorf("failed to delete screening question: %w", err)
    }

    log.Printf("Deleted screening question with ID: %s, ageGroup: %s", questionID, ageGroup)
    return nil
}

// SaveScreeningResult saves the user's screening submission and risk level.
func SaveScreeningResult(ctx context.Context, userID, ageGroup string, answers []bool, riskLevel string, firebaseToken string) error {
    firestoreClient, err := GetFirestoreClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Firestore: %w", err)
    }
    defer firestoreClient.Close()

    docRef := firestoreClient.Collection("users").Doc(userID).Collection("screenings").Doc("current")

    _, err = docRef.Set(ctx, map[string]interface{}{
        "ageGroup":  ageGroup,
        "answers":   answers,
        "riskLevel": riskLevel,
        "timestamp": firestore.ServerTimestamp,
    }, firestore.MergeAll)
    if err != nil {
        return fmt.Errorf("failed to save screening result: %w", err)
    }

    log.Printf("Updated screening result for userID: %s, ageGroup: %s", userID, ageGroup)
    return nil
}