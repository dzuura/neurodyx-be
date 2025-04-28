package models

type ScreeningQuestion struct {
    AgeGroup string `json:"ageGroup"`
    Question string `json:"question"`
}

type ScreeningSubmission struct {
    AgeGroup string `json:"ageGroup"`
    Answers  []bool `json:"answers"`
}
