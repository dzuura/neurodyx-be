package models

// ScreeningQuestion represents a screening question with unique fields.
type ScreeningQuestion struct {
    ID       string `json:"id,omitempty"`
    AgeGroup string `json:"ageGroup"`
    Question string `json:"question"`
}

// ScreeningSubmission represents a user's submission for a screening question.
type ScreeningSubmission struct {
    AgeGroup string  `json:"ageGroup"`
    Answers  []bool  `json:"answers,omitempty"`
}