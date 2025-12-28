// internal/api/schema_handler.go
package api

import (
    "encoding/json"
    "net/http"
    "vax/internal/schema"
)

func HandleGetSchema(w http.ResponseWriter, r *http.Request) {
    action := r.URL.Query().Get("action")

    var schemaData map[string]interface{}

    switch action {
    case "update_profile":
        schemaData = schema.GetUpdateProfileSchema()
    default:
        http.Error(w, "Unknown action", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(schemaData)
}
