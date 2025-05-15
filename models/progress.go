package models

import "time"

// DailyProgress represents a user's daily progress for therapy activities.
type DailyProgress struct {
    UserID         string    `json:"userID"`
    Date           time.Time `json:"date"`
    TherapyCount   int       `json:"therapyCount"`
    StreakAchieved bool      `json:"streakAchieved"`
}

// ProgressDetail represents the detailed progress for a specific month.
type ProgressDetail struct {
    Date           time.Time `json:"date"`
    Status         string    `json:"status"`
}