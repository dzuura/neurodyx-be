package middleware

import (
    "net/http"

    "github.com/dzuura/neurodyx-be/config"
)

// AdminMiddleware ensures that the user has admin privileges.
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := r.Context().Value(UserIDKey).(string)
        if !ok {
            http.Error(w, "User ID missing", http.StatusUnauthorized)
            return
        }

        firestoreClient, err := config.App.Firestore(r.Context())
        if err != nil {
            http.Error(w, "Failed to connect to Firestore", http.StatusInternalServerError)
            return
        }
        defer firestoreClient.Close()

        doc, err := firestoreClient.Collection("users").Doc(userID).Get(r.Context())
        if err != nil {
            http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
            return
        }

        var userData map[string]interface{}
        if err := doc.DataTo(&userData); err != nil {
            http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
            return
        }

        isAdmin, ok := userData["isAdmin"].(bool)
        if !ok || !isAdmin {
            http.Error(w, "Admin access required", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    }
}