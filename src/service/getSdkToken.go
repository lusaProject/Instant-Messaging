package httpService

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

func getSdkTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		decoder := json.NewDecoder(r.Body)
		var data map[string]string
		decoder.Decode(&data)

		if data["token"] == "" {
			http.Error(w, "invalid input", http.StatusUnauthorized)
			return
		}

		tokenString := data["token"]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("XLC8J22HOmTT2zZyeRVq3RKQ3juNkZR6OJUgIW2Y"), nil
		})

		if err != nil {
			res := map[string]interface{}{
				"code": 1005,
				"msg":  "Failed to decode or validate token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(res)
			return
		}

		if !token.Valid {
			res := map[string]interface{}{
				"code": 1006,
				"msg":  "Invalid token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(res)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userid := fmt.Sprintf("%v", claims["account"])

		var roomid string
		if data["roomId"] == "" {
			roomid = uuid.New().String()
		} else {
			roomid = data["roomId"]
		}

		sdkToken := getSdkToken(AppId, AppKey, roomid, userid)

		response := map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{
				"sdkToken": sdkToken,
				"appId":    AppId,
				"roomName": roomid,
			},
		}

		json.NewEncoder(w).Encode(response)

	} else {
		w.WriteHeader(http.StatusOK)
	}

}
