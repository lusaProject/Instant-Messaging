package httpService

import (
	db "demo/dbManage"
	"net/http"

	"log"
)

func ServiceInit() {
	clientDB = db.RedisInit()
	asyncCronScheduler()

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/getSdkToken", getSdkTokenHandler)
	http.HandleFunc("/updateVersion", updateVersionHandler)

	log.Println("Server started on 8099")
	certFile := "/etc/nginx/ssl/xxxxx.crt"
	keyFile := "/etc/nginx/ssl/xxxxxx.key"
	err := http.ListenAndServeTLS(":8099", certFile, keyFile, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	clientDB.Close()
}
