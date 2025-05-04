package services


// ValidateStringMatch checks if the user's string answer matches the correct answer
func ValidateStringMatch(userAnswer, correctAnswer string) bool {
    return userAnswer == correctAnswer
}

// ValidateSequence checks if the user's sequence matches the correct sequence exactly
func ValidateSequence(userSequence, correctSequence []string) bool {
    if len(userSequence) != len(correctSequence) {
        return false
    }
    for i, char := range userSequence {
        if char != correctSequence[i] {
            return false
        }
    }
    return true
}

// ValidatePairs checks if the user's pairs match the correct pairs
func ValidatePairs(userPairs []map[string]string, correctPairs map[string]string) bool {
    userPairMap := make(map[string]string)
    for _, pair := range userPairs {
        left, leftOk := pair["left"]
        right, rightOk := pair["right"]
        if !leftOk || !rightOk {
            return false
        }
        userPairMap[left] = right
    }

    if len(userPairMap) != len(correctPairs) {
        return false
    }

    for left, correctRight := range correctPairs {
        userRight, exists := userPairMap[left]
        if !exists || userRight != correctRight {
            return false
        }
    }

    return true
}