package models

// Point represents a coordinate point for touch-based questions.
type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

// AssessmentQuestion defines a flexible structure for various question types.
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
    PathData       []Point           `json:"pathData,omitempty"`
}

// AssessmentSubmission holds a user's answer submission.
type AssessmentSubmission struct {
    QuestionID string      `json:"questionID"`
    Answer     interface{} `json:"answer"`
}

// AssessmentResult stores the result of an assessment by type.
type AssessmentResult struct {
    Type           string `json:"type"`
    CorrectAnswers int    `json:"correctAnswers"`
    TotalQuestions int    `json:"totalQuestions"`
    Status         string `json:"status"`
}