package models

type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

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

type AssessmentSubmission struct {
    QuestionID string      `json:"questionID"`
    Answer     interface{} `json:"answer"`
}

type AssessmentResult struct {
    Type           string `json:"type"`
    CorrectAnswers int    `json:"correctAnswers"`
    TotalQuestions int    `json:"totalQuestions"`
    Status         string `json:"status"`
}