package models

// TherapyQuestion represents a therapy question with unique fields.
type TherapyQuestion struct {
    ID             string            `json:"id"`
    Type           string            `json:"type"`
    Category       string            `json:"category"`
    Content        string            `json:"content,omitempty"`
    Description    string            `json:"description,omitempty"`
    ImageURL       string            `json:"imageURL,omitempty"`
    SoundURL       string            `json:"soundURL,omitempty"`
    Options        []string          `json:"options,omitempty"`
    LeftItems      []string          `json:"leftItems,omitempty"`
    RightItems     []string          `json:"rightItems,omitempty"`
    CorrectAnswer  string            `json:"correctAnswer,omitempty"`
    CorrectSequence []string         `json:"correctSequence,omitempty"`
    CorrectPairs   map[string]string `json:"correctPairs,omitempty"`
}

// TherapySubmission represents a user's submission for a therapy question.
type TherapySubmission struct {
    QuestionID string      `json:"questionID"`
    Answer     interface{} `json:"answer"`
}

// TherapyResult represents the result of a user's therapy session.
type TherapyResult struct {
    Type           string `json:"type"`
    Category       string `json:"category"`
    CorrectAnswers int    `json:"correctAnswers"`
    TotalQuestions int    `json:"totalQuestions"`
    Status         string `json:"status"`
}

// TherapyCategory represents a therapy category with its description.
type TherapyCategory struct {
    Category    string `json:"category"`
    Description string `json:"description"`
}