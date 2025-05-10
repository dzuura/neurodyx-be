package models

// AssessmentQuestion represents a single assessment question with various types and answer formats.
type AssessmentQuestion struct {
    ID             string            `json:"id"`
    Type           string            `json:"type"`
    Category       string            `json:"category"`
    Content        string            `json:"content,omitempty"`
    ImageURL       string            `json:"imageURL,omitempty"`
    SoundURL       string            `json:"soundURL,omitempty"`
    Options        []string          `json:"options,omitempty"`
    LeftItems      []string          `json:"leftItems,omitempty"`
    RightItems     []string          `json:"rightItems,omitempty"`
    CorrectAnswer  string            `json:"correctAnswer,omitempty"`
    CorrectSequence []string         `json:"correctSequence,omitempty"`
    CorrectPairs   map[string]string `json:"correctPairs,omitempty"`
}

// AssessmentSubmission represents a user's submission for an assessment question.
type AssessmentSubmission struct {
    QuestionID string      `json:"questionID"`
    Answer     interface{} `json:"answer"`
}

// AssessmentResult represents the result of a user's assessment for a specific type.
type AssessmentResult struct {
    Type           string `json:"type"`
    CorrectAnswers int    `json:"correctAnswers"`
    TotalQuestions int    `json:"totalQuestions"`
    Status         string `json:"status"`
}