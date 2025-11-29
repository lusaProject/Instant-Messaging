package httpService

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"io"
	"mime/multipart"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		fileType := r.FormValue("type")

		resultDir := "/var/data/" + fileType + "/"

		if _, err := os.Stat(resultDir); os.IsNotExist(err) {
			err := os.Mkdir(resultDir, 0755)
			if err != nil {
				log.Print("创建目录失败: ", err)
			} 
		} 

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving music file", http.StatusBadRequest)
			fmt.Println("Error retrieving music file", err)
			return
		}
		defer file.Close()

		fileName := resultDir + fileHeader.Filename
		log.Printf("Uploaded file name: %s\n", fileName)

		err = saveFile(file, fileName)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			fmt.Println("Error saving file", err)
			return
		}

		response := map[string]interface{}{
			"code": 0,
			"msg":  "success",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func saveFile(file multipart.File, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	return err
}
