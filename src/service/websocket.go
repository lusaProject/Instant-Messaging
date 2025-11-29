package httpService

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	userToken := r.Header.Get("User-Token")
	userRecv := r.Header.Get("User")

	var temp bool
	if userToken == "" {
		userToken = r.URL.Query().Get("user_token")
		userRecv = r.URL.Query().Get("user")
		temp = true
	} else {
		temp = false
	}

	//token鉴权
	var userid string
	{
		tokenString := userToken
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
		userid = fmt.Sprintf("%v", claims["account"])
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	//保存数据
	mu_ws.Lock()
	clients[userid] = ws
	saveUserInfo(userRecv, userid, temp)
	mu_ws.Unlock()

	for {
		var data map[string]interface{}

		err := ws.ReadJSON(&data)
		if err != nil {
			user := getUserBase(userid)
			log.Printf("close: %v-%s", err, user.Name)

			//状态更新
			go closeConnect(temp, userid)
			break
		}

		//日志过滤
		event, _ := data["event"].(string)
		if event != "ping" && event != "queryAllUsers" {
			user := getUserBase(userid)
			log.Print(data, "-", user.Name)
		}

		//自动重连
		{
			mu_ws.Lock()
			_, ok := clients[userid]
			if !ok {
				clients[userid] = ws

				//记录时间戳
				var user UserInfo
				user.Timestamp = "0"
				userInfoHSet(userid, user)
			}
			mu_ws.Unlock()
		}

		//会议事件
		go signaling(userid, data)

		//直播事件
		go liveRoomEvent(userid, data)

	}
}
