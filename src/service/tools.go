package httpService

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/robfig/cron/v3"
)

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	if len(parts1) < len(parts2) {
		return -1
	} else if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}

func getSdkToken(appId string, appKey string, roomid string, userid string) (sdkToken string) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	now := time.Now().Unix()
	expirationTime := now + 86400
	claims["exp"] = expirationTime * 1000
	claims["room_id"] = roomid
	claims["user_id"] = userid
	claims["app_id"] = appId

	signingKey := []byte(appKey)
	sdkToken, _ = token.SignedString(signingKey)
	log.Print("sdkToken ===> ", sdkToken)
	return sdkToken
}

func saveUserInfo(userRecv string, userid string, temp bool) {
	var userDataJson UserBaseInfo
	json.Unmarshal([]byte(userRecv), &userDataJson)

	var nameData string
	if temp {
		nameData, _ = url.QueryUnescape(userDataJson.Name)
	} else {
		name, _ := base64.StdEncoding.DecodeString(userDataJson.Name)
		nameData = string(name)
	}

	var user UserInfo
	user.ID = userid
	user.IsOnline = "true"

	if userid == "228414032e7be3a36ae7c98cf7f3f6690" {
		user.Name = "postman"
	} else {
		user.Name = nameData
	}

	//头像获取
	var urlAvatar string
	{
		rand.Seed(time.Now().UnixNano())
		randNum := rand.Intn(36) + 147258
		numStr := strconv.Itoa(randNum)
		urlAvatar = "https://data.devplay.cc/avatar/" + numStr + ".jpg"
	}

	existJson, _ := clientDB.HGet("ALLUSER", userid).Result()
	if existJson == "" {
		user.Avatar = urlAvatar
		userJson, _ := json.Marshal(user)
		clientDB.HSet("ALLUSER", userid, userJson)
	} else {
		var userDB UserInfo
		json.Unmarshal([]byte(existJson), &userDB)

		var user UserInfo
		user.IsOnline = "true"
		user.Name = nameData

		if userDB.Avatar == "" {
			user.Avatar = urlAvatar
		}

		userInfoHSet(userid, user)
	}

	clientDB.SRem("BUSYUSERS", userid)

	// 更新通知
	if !temp {
		userBase := getUserBase(userid)

		userState := UserListInfo{
			ID:       userBase.ID,
			Name:     userBase.Name,
			Avatar:   userBase.Avatar,
			IsOnline: "true",
		}

		status := map[string]interface{}{
			"event": "updateUser",
			"data":  userState,
		}

		asyncBroadcast(userid, status)
	}

	//主播上线
	{
		user := getUserInfo(userid)
		if user.LiveRoom != "" && user.LiveRoomStatus != "" {
			notifyData := map[string]interface{}{
				"event": "onLiveRoomStatus",
				"data": map[string]interface{}{
					"liveRoom":       user.LiveRoom,
					"liveRoomStatus": user.LiveRoomStatus,
				},
			}
			asyncAllNotify(user.LiveRoom, notifyData)
		}
	}

	//记录时间戳
	{
		var user UserInfo
		user.Timestamp = "0"
		userInfoHSet(userid, user)
	}
}

func closeConnect(temp bool, userid string) {
	//读写加锁
	mu_ws.Lock()
	delete(clients, userid)
	mu_ws.Unlock()

	clientDB.SRem("BUSYUSERS", userid)

	if temp {
		clientDB.HDel("ALLUSER", userid)
	} else {

		//离线状态
		var user UserInfo
		user.IsOnline = "false"
		userInfoHSet(userid, user)

		// 更新通知
		{
			userBase := getUserBase(userid)

			userState := UserListInfo{
				ID:       userBase.ID,
				Name:     userBase.Name,
				Avatar:   userBase.Avatar,
				IsOnline: "false",
			}

			status := map[string]interface{}{
				"event": "updateUser",
				"data":  userState,
			}

			asyncBroadcast(userid, status)
		}

		//主播离开
		{
			user := getUserInfo(userid)
			if user.LiveRoom != "" {
				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       user.LiveRoom,
						"liveRoomStatus": Offline,
					},
				}
				asyncAllNotify(user.LiveRoom, notifyData)
			}
		}

		//记录时间戳
		{
			timestamp := time.Now().Unix()
			timestampStr := strconv.FormatInt(timestamp, 10)

			var user UserInfo
			user.Timestamp = timestampStr
			userInfoHSet(userid, user)
		}

	}
}

func asyncNotify(userid string, key string, notifyData map[string]interface{}) {
	go func() {
		setList, _ := clientDB.SMembers(key).Result()
		for _, ID := range setList {
			if userid != ID {
				writeJSON(ID, notifyData)
			}
		}
	}()
}

func asyncAllNotify(key string, notifyData map[string]interface{}) {
	go func() {
		setList, _ := clientDB.SMembers(key).Result()
		for _, ID := range setList {
			writeJSON(ID, notifyData)
		}
	}()
}

func asyncBroadcast(myid string, data map[string]interface{}) {
	go func() {
		for userid := range clients {
			if userid != myid {
				writeJSON(userid, data)
			}
		}
	}()
}

func writeJSON(ID string, notifyData map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	//读写加锁
	mu_ws.Lock()
	client, ok := clients[ID]
	mu_ws.Unlock()

	if !ok {
		log.Printf("client not found: %s", ID)
	} else {
		err := client.WriteJSON(notifyData)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}
}

func onRoomViewerChange(artist string, liveRoom string) {
	setList, _ := clientDB.SMembers(liveRoom).Result()

	var viewerList []UserListInfo = []UserListInfo{}
	for count, id := range setList {

		if id != artist {
			userJson, _ := clientDB.HGet("ALLUSER", id).Result()
			if userJson != "" {
				var user UserInfo
				json.Unmarshal([]byte(userJson), &user)

				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				viewerList = append(viewerList, userlist)
			}
		}

		if count > 1 {
			break
		}
	}

	notifyData := map[string]interface{}{
		"event": "onRoomViewerChange",
		"data": map[string]interface{}{
			"roomId":     liveRoom,
			"total":      len(setList) - 1,
			"recentList": viewerList,
		},
	}

	asyncAllNotify(liveRoom, notifyData)
}

func userInfoHSet(userid string, user UserInfo) {
	userJson, _ := clientDB.HGet("ALLUSER", userid).Result()

	if userJson != "" {

		var userDB UserInfo
		json.Unmarshal([]byte(userJson), &userDB)

		var userModify UserInfo

		if user.ID == "" {
			userModify.ID = userDB.ID
		} else {
			userModify.ID = user.ID
		}

		if user.Name == "" {
			userModify.Name = userDB.Name
		} else {
			userModify.Name = user.Name
		}

		if user.Avatar == "" {
			userModify.Avatar = userDB.Avatar
		} else {
			userModify.Avatar = user.Avatar
		}

		if user.IsOnline == "" {
			userModify.IsOnline = userDB.IsOnline
		} else {
			userModify.IsOnline = user.IsOnline
		}

		if user.ViewerState == "" {
			userModify.ViewerState = userDB.ViewerState
		} else {
			userModify.ViewerState = user.ViewerState
		}

		if user.UserState == "" {
			userModify.UserState = userDB.UserState
		} else {
			userModify.UserState = user.UserState
		}

		if user.ArtistState == "" {
			userModify.ArtistState = userDB.ArtistState
		} else {
			userModify.ArtistState = user.ArtistState
		}

		if user.LiveRoom == "" {
			userModify.LiveRoom = userDB.LiveRoom
		} else {
			userModify.LiveRoom = user.LiveRoom
		}

		if user.LiveRoomStatus == "" {
			userModify.LiveRoomStatus = userDB.LiveRoomStatus
		} else {
			userModify.LiveRoomStatus = user.LiveRoomStatus
		}

		if user.Timestamp == "" {
			userModify.Timestamp = userDB.Timestamp
		} else {
			userModify.Timestamp = user.Timestamp
		}

		writeDB, _ := json.Marshal(userModify)
		clientDB.HSet("ALLUSER", userid, writeDB)
	}
}

func userCleanHSet(userid string, user UserInfo) {
	userJson, _ := clientDB.HGet("ALLUSER", userid).Result()

	if userJson != "" {

		var userDB UserInfo
		json.Unmarshal([]byte(userJson), &userDB)

		var userModify UserInfo

		if user.ID != Clean {
			userModify.ID = userDB.ID
		}

		if user.Name != Clean {
			userModify.Name = userDB.Name
		}

		if user.Avatar != Clean {
			userModify.Avatar = userDB.Avatar
		}

		if user.IsOnline != Clean {
			userModify.IsOnline = userDB.IsOnline
		}

		if user.ViewerState != Clean {
			userModify.ViewerState = userDB.ViewerState
		}

		if user.UserState != Clean {
			userModify.UserState = userDB.UserState
		}

		if user.ArtistState != Clean {
			userModify.ArtistState = userDB.ArtistState
		}

		if user.LiveRoom != Clean {
			userModify.LiveRoom = userDB.LiveRoom
		}

		if user.LiveRoomStatus != Clean {
			userModify.LiveRoomStatus = userDB.LiveRoomStatus
		}

		if user.Timestamp != Clean {
			userModify.Timestamp = userDB.Timestamp
		}

		writeDB, _ := json.Marshal(userModify)
		clientDB.HSet("ALLUSER", userid, writeDB)
	}
}

func asyncSetUserRoomPermission(userId string, roomId string, pubPermission bool) {
	go func() {
		url := fmt.Sprintf("%s/v1/%s/%s/%s/permission", CloudFlare, AppId, userId, roomId)
		log.Print(" asyncSetUserRoomPermission ===> url ", url)

		data := map[string]bool{
			"pubPermission": pubPermission,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}()
}

func asyncKickOutLiveRoom(userId string, roomId string, userIds []interface{}) {
	go func() {
		url := fmt.Sprintf("%s/v1/%s/%s/kickout", CloudFlare, userId, roomId)
		log.Print(" asyncKickOutLiveRoom ===> url ", url)

		requestData := map[string]interface{}{
			"users": userIds,
		}

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			fmt.Println("failed to marshal JSON: ", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("failed to create request ", err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("failed to send request: ", err)
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}()
}

func asyncCloseRoom(roomId string) {
	go func() {
		url := fmt.Sprintf("%s/v1/%s/%s/live_close", CloudFlare, AppId, roomId)
		log.Print(" asyncCloseRoom ===> url ", url)

		data := map[string]interface{}{}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", AppKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}()
}

func asyncEndPk(userId string, otherId string, roomId string) {
	go func() {
		url := fmt.Sprintf("%s/v1/%s/%s/%s/endPk", CloudFlare, AppId, userId, roomId)
		log.Print(" asyncEndPk ===> url ", url)

		data := map[string]interface{}{
			"roomId":   roomId,
			"remoteId": otherId,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON: ", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request: ", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request: ", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}()
}

func asyncStartPk(userId string, roomId string, remoteRoomId string) {
	go func() {
		url := fmt.Sprintf("%s/v1/%s/%s/%s/startPk", CloudFlare, AppId, userId, roomId)
		log.Print(" asyncStartPk ===> url ", url)

		data := map[string]interface{}{
			"remoteRoomId": remoteRoomId,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON: ", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request: ", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request: ", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}()
}

func asyncGetLiveList() {
	go func() {
		hashGetAll, _ := clientDB.HGetAll("ALLLIVEROOM").Result()
		for _, userJson := range hashGetAll {
			var roomBase LiveRoomBase
			json.Unmarshal([]byte(userJson), &roomBase)

			user := getUserInfo(roomBase.Artist)

			timestamp, _ := strconv.ParseInt(user.Timestamp, 10, 64)

			currentTime := time.Now().Unix()

			if (currentTime-timestamp) > 90 && user.Timestamp != "0" {

				//关闭房间
				asyncCloseRoom(roomBase.LiveID)

				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       roomBase.LiveID,
						"liveRoomStatus": Close,
					},
				}

				{
					setList, _ := clientDB.SMembers(roomBase.LiveID).Result()
					for _, id := range setList {
						var user UserInfo
						user.ViewerState = Clean
						user.UserState = Clean
						user.ArtistState = Clean
						user.LiveRoomStatus = Clean
						userCleanHSet(id, user)
					}
				}

				asyncAllNotify(roomBase.LiveID, notifyData)

				clientDB.HDel("ALLLIVEROOM", roomBase.LiveID)
				clientDB.Del(roomBase.LiveID)

				roomCount := roomBase.LiveID + "_count"
				clientDB.Del(roomCount)
			}
		}
	}()
}

func asyncCronScheduler() {
	go func() {
		c := cron.New()
		_, err := c.AddFunc("@every 60s", func() {
			asyncGetLiveList()
		})
		if err != nil {
			log.Println("添加任务失败:", err)
			return
		}
		c.Start()
		defer c.Stop()
		select {}
	}()
}

func getLiveStatus(liveRoom string) UserInfo {
	var userData UserInfo
	liveJson, _ := clientDB.HGet("ALLLIVEROOM", liveRoom).Result()

	if liveJson != "" {
		var room LiveRoomBase
		json.Unmarshal([]byte(liveJson), &room)

		userJson, _ := clientDB.HGet("ALLUSER", room.Artist).Result()
		if userJson != "" {
			json.Unmarshal([]byte(userJson), &userData)
		}
	}

	return userData
}

func getLiveArtist(liveR string) string {
	var room LiveRoomBase
	liveJson, _ := clientDB.HGet("ALLLIVEROOM", liveR).Result()

	if liveJson != "" {
		json.Unmarshal([]byte(liveJson), &room)
	}

	return room.Artist
}

func setLiveStatus(live string, status string) {
	artist := getLiveArtist(live)
	var user UserInfo
	user.LiveRoomStatus = status
	userInfoHSet(artist, user)
}

func setUserStatus(userId string, status string) {
	var user UserInfo
	user.LiveRoomStatus = status
	userInfoHSet(userId, user)
}

func getUserBase(userId string) UserBaseInfo {
	var userBase UserBaseInfo
	userJson, _ := clientDB.HGet("ALLUSER", userId).Result()
	if userJson != "" {
		var user UserInfo
		json.Unmarshal([]byte(userJson), &user)

		userBase.ID = user.ID
		userBase.Name = user.Name
		userBase.Avatar = user.Avatar
	}

	return userBase
}

func getUserInfo(userId string) UserInfo {
	var user UserInfo
	userJson, _ := clientDB.HGet("ALLUSER", userId).Result()
	if userJson != "" {
		json.Unmarshal([]byte(userJson), &user)
	}
	return user
}
