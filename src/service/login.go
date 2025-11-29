package httpService

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer cf_xxxxxxxxxx" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var data map[string]string
		decoder.Decode(&data)

		if data["account"] == "" || data["passwd"] == "" {
			http.Error(w, "Invalid account or password", http.StatusUnauthorized)
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["account"] = data["account"]
		claims["expiresDate"] = "2726643496"
		signingKey := []byte("XLC8J22HOmTT2zZyeRVq3RKQ3juNkZR6OJUgIW2Y")
		tokenString, err := token.SignedString(signingKey)
		if err != nil {
			log.Println("Failed to generate token:", err)
		}

		response := map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{
				"userToken": tokenString,
				"appId":     AppId,
			},
		}

		json.NewEncoder(w).Encode(response)

	} else {
		w.WriteHeader(http.StatusOK)
	}

}
