package httpService

import (
	"encoding/json"
	"log"
	"net/http"
)

func updateVersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		decoder.Decode(&data)
		log.Print(data)

		isForceUpdate, _ := data["isForceUpdate"].(bool)
		newVersion, _ := data["newVersion"].(string)
		desc, _ := data["desc"].(string)
		url, _ := data["url"].(string)

		clientDB.HSet("version", "isForceUpdate", isForceUpdate)
		clientDB.HSet("version", "newVersion", newVersion)
		clientDB.HSet("version", "desc", desc)
		clientDB.HSet("version", "url", url)

		response := map[string]interface{}{
			"code": "0",
			"msg":  "success",
		}

		json.NewEncoder(w).Encode(response)

	} else {
		w.WriteHeader(http.StatusOK)
	}
}
