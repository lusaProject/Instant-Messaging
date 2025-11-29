package httpService

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

func signaling(userid string, data map[string]interface{}) {
	cmd, _ := data["event"].(string)

	switch {
	case cmd == "queryAllUsers":
		var onlineList []UserListInfo
		hashGetAll, _ := clientDB.HGetAll("ALLUSER").Result()
		for _, userJson := range hashGetAll {
			var user UserInfo
			json.Unmarshal([]byte(userJson), &user)

			var userlist UserListInfo
			userlist.ID = user.ID
			userlist.Name = user.Name
			userlist.Avatar = user.Avatar
			userlist.IsOnline = user.IsOnline
			onlineList = append(onlineList, userlist)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "queryAllUsers",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"users": onlineList,
			},
		}

		writeJSON(userid, response)

	case cmd == "allowAddUsers":

		dataMap, _ := data["data"].(map[string]interface{})
		id, _ := dataMap["myId"].(string)

		var onlineList []UserListInfo
		hashGetAll, _ := clientDB.HGetAll("ALLUSER").Result()
		for _, userJson := range hashGetAll {
			var user UserInfo
			json.Unmarshal([]byte(userJson), &user)

			ok, _ := clientDB.SIsMember(id, user.ID).Result()
			if !ok {
				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				onlineList = append(onlineList, userlist)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "allowAddUsers",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"users": onlineList,
			},
		}
		writeJSON(userid, response)

	case cmd == "updateUsername":
		dataMap, _ := data["data"].(map[string]interface{})
		name, _ := dataMap["name"].(string)

		//离线状态
		var user UserInfo
		user.IsOnline = "true"
		user.Name = name
		userInfoHSet(userid, user)

		//更新通知
		{
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

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "updateUsername",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}
		writeJSON(userid, response)

	case cmd == "requestUpdate":
		dataMap, _ := data["data"].(map[string]interface{})
		id, _ := dataMap["myId"].(string)
		friend, _ := dataMap["friendId"].(string)
		state, _ := dataMap["state"].(float64)

		var friendDB RequestList
		var myDB RequestList

		myKey := id + "_request"
		friendKey := friend + "_request"
		{
			myJson, _ := clientDB.HGet(myKey, friend).Result()
			json.Unmarshal([]byte(myJson), &myDB)

			friendJson, _ := clientDB.HGet(friendKey, id).Result()
			json.Unmarshal([]byte(friendJson), &friendDB)

			myVal := RequestList{
				myDB.Type,
				int(state),
				myDB.ID,
				myDB.Name,
				myDB.Avatar,
			}

			friendVal := RequestList{
				friendDB.Type,
				int(state),
				friendDB.ID,
				friendDB.Name,
				friendDB.Avatar,
			}

			writeMyDB, _ := json.Marshal(myVal)
			clientDB.HSet(myKey, friend, writeMyDB)

			writefriendDB, _ := json.Marshal(friendVal)
			clientDB.HSet(friendKey, id, writefriendDB)

			//同意添加好友
			if int(state) == 1 {
				clientDB.SAdd(id, friend)
				clientDB.SAdd(friend, id)
			}
		}

		//通知对方
		{
			notifyData := map[string]interface{}{
				"event": "onRequestUpdate",
				"data": map[string]interface{}{
					"myId":     id,
					"friendId": friend,
					"state":    int(state),
				},
			}

			writeJSON(friend, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "requestUpdate",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"myId":     id,
				"friendId": friend,
				"state":    int(state),
			},
		}
		writeJSON(userid, response)

	case cmd == "friendRequest":
		dataMap, _ := data["data"].(map[string]interface{})
		id, _ := dataMap["myId"].(string)

		myKey := id + "_request"
		var requestList []RequestList
		hashGetAll, _ := clientDB.HGetAll(myKey).Result()
		for _, userJson := range hashGetAll {
			var user RequestList
			json.Unmarshal([]byte(userJson), &user)
			requestList = append(requestList, user)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "friendRequest",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"requestList": requestList,
			},
		}
		writeJSON(userid, response)

	case cmd == "addFriend":
		dataMap, _ := data["data"].(map[string]interface{})
		id, _ := dataMap["myId"].(string)
		friend, _ := dataMap["friendId"].(string)

		var friendDB UserInfo
		var myDB UserInfo

		if id != friend {

			friendJson, _ := clientDB.HGet("ALLUSER", friend).Result()
			json.Unmarshal([]byte(friendJson), &friendDB)

			myJson, _ := clientDB.HGet("ALLUSER", id).Result()
			json.Unmarshal([]byte(myJson), &myDB)

			myVal := RequestList{
				1,
				0,
				myDB.ID,
				myDB.Name,
				myDB.Avatar,
			}

			friendVal := RequestList{
				0,
				0,
				friendDB.ID,
				friendDB.Name,
				friendDB.Avatar,
			}

			myKey := id + "_request"
			writeMyDB, _ := json.Marshal(myVal)

			friendKey := friend + "_request"
			writefriendDB, _ := json.Marshal(friendVal)

			clientDB.HSet(myKey, friend, writefriendDB)
			clientDB.HSet(friendKey, id, writeMyDB)

			//通知对方
			{
				notifyData := map[string]interface{}{
					"event": "onAddFriend",
					"data": map[string]interface{}{
						"myId":     id,
						"friendId": friend,
					},
				}
				writeJSON(friend, notifyData)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "addFriend",
			"data":  friendDB,
			"sn":    int(sn),
		}

		writeJSON(userid, response)

	case cmd == "querySdkToken":
		dataMap, _ := data["data"].(map[string]interface{})
		tokenString, _ := dataMap["userToken"].(string)
		roomName, _ := dataMap["roomName"].(string)

		if tokenString == "" {
			response := map[string]interface{}{"msg": "invalid input"}
			writeJSON(userid, response)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("XLC8J22HOmTT2zZyeRVq3RKQ3juNkZR6OJUgIW2Y"), nil
		})

		if err != nil {
			response := map[string]interface{}{
				"code": 1005,
				"msg":  "Failed to decode or validate token",
			}
			writeJSON(userid, response)
			return
		}

		if !token.Valid {
			response := map[string]interface{}{
				"code": 1006,
				"msg":  "Invalid token",
			}

			writeJSON(userid, response)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userid := fmt.Sprintf("%v", claims["account"])

		var roomid string
		if roomName == "" {
			roomid = uuid.New().String()
		} else {
			roomid = roomName
		}

		sdkToken := getSdkToken(AppId, AppKey, roomid, userid)

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "querySdkToken",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"sdkToken": sdkToken,
				"appId":    AppId,
				"roomName": roomid,
			},
		}

		writeJSON(userid, response)

	case cmd == "queryFriends":

		//刷新上线
		{
			var user UserInfo
			user.IsOnline = "true"
			userInfoHSet(userid, user)

			// 更新通知
			{
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
		}

		//请求数量
		var sum int = 0
		{
			myKey := userid + "_request"
			hashGetAll, _ := clientDB.HGetAll(myKey).Result()
			for _, userJson := range hashGetAll {
				var user RequestList
				json.Unmarshal([]byte(userJson), &user)
				if user.Type == 1 && user.State == 0 {
					sum++
				}
			}
		}

		//好友列表
		setList, _ := clientDB.SMembers(userid).Result()

		var onlineList []UserListInfo
		for _, idFriend := range setList {
			userJson, _ := clientDB.HGet("ALLUSER", idFriend).Result()
			if userJson != "" {
				var user UserInfo
				json.Unmarshal([]byte(userJson), &user)

				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				onlineList = append(onlineList, userlist)
			} else {
				clientDB.SRem(userid, idFriend)
			}
		}

		//加上本机
		{
			userJson, _ := clientDB.HGet("ALLUSER", userid).Result()
			if userJson != "" {
				var user UserInfo
				json.Unmarshal([]byte(userJson), &user)

				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				onlineList = append(onlineList, userlist)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "queryFriends",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"applyNotifyNum": sum,
				"users":          onlineList,
			},
		}

		writeJSON(userid, response)

	case cmd == "ping":
		response := map[string]interface{}{
			"event": "pong",
		}
		writeJSON(userid, response)

	case cmd == "quitRoom":
		var quitRoom QuitRoom

		if err := mapstructure.Decode(data, &quitRoom); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(quitRoom)

		roomid := quitRoom.Data.RoomID

		clientDB.SRem(roomid, quitRoom.Data.User.ID)
		clientDB.SRem("BUSYUSERS", quitRoom.Data.User.ID)

		setList, _ := clientDB.SMembers(roomid).Result()
		if len(setList) == 0 {
			clientDB.Del(roomid)
		}

		//通知房间
		{
			var event string
			if len(setList) > 1 {
				event = "onQuitRoom"
			} else {
				event = "onDestroyRoom"
			}

			notifyData := map[string]interface{}{
				"event": event,
				"data": map[string]interface{}{
					"roomId": roomid,
					"user":   quitRoom.Data.User,
				},
			}

			asyncNotify(userid, roomid, notifyData)
		}

		//临时房间
		{
			roomTemp := roomid + "_temp"
			clientDB.SRem(roomTemp, quitRoom.Data.User.ID)

			notifyData := map[string]interface{}{
				"event": "onQuitRoom",
				"data": map[string]interface{}{
					"roomId": roomid,
					"user":   quitRoom.Data.User,
				},
			}

			asyncNotify(userid, roomTemp, notifyData)
		}

		sn := quitRoom.Sn
		response := map[string]interface{}{
			"event": "quitRoom",
			"sn":    sn,
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "rejectCall":
		var rejectCall RejectCall

		if err := mapstructure.Decode(data, &rejectCall); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(rejectCall)

		roomid := rejectCall.Data.RoomID
		clientDB.SRem("BUSYUSERS", rejectCall.Data.Rejecter.ID)
		clientDB.SRem("BUSYUSERS", rejectCall.Data.Inviter.ID)

		roomTemp := roomid + "_temp"

		notifyData := map[string]interface{}{
			"event": "onRejectCall",
			"data": map[string]interface{}{
				"roomId":   roomid,
				"inviter":  rejectCall.Data.Inviter,
				"rejecter": rejectCall.Data.Rejecter,
			},
		}

		//通知房间
		writeJSON(rejectCall.Data.Inviter.ID, notifyData)
		writeJSON(rejectCall.Data.Rejecter.ID, notifyData)

		asyncNotify(userid, roomid, notifyData)
		asyncNotify(userid, roomTemp, notifyData)

		sn := rejectCall.Sn
		response := map[string]interface{}{
			"event": "rejectCall",
			"sn":    sn,
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "onInviteExpired":

		var onInviteExpired OnInviteExpired

		if err := mapstructure.Decode(data, &onInviteExpired); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(onInviteExpired)

		roomid := onInviteExpired.Data.RoomID
		//通知房间
		asyncNotify(userid, roomid, data)

	case cmd == "joinRoom":

		var joinRoom JoinRoom

		if err := mapstructure.Decode(data, &joinRoom); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(joinRoom)

		roomid := joinRoom.Data.RoomID

		notifyData := map[string]interface{}{
			"event": "onJoinRoom",
			"data": map[string]interface{}{
				"roomId": roomid,
				"user":   joinRoom.Data.User,
			},
		}

		//通知房间
		asyncNotify(userid, roomid, notifyData)

		clientDB.SAdd(roomid, joinRoom.Data.User.ID)
		clientDB.Expire(roomid, time.Second*86400)

		setList, _ := clientDB.SMembers(roomid).Result()
		var roomList []UserListInfo = []UserListInfo{}
		{
			for _, roomUser := range setList {
				userJson, _ := clientDB.HGet("ALLUSER", roomUser).Result()
				var user UserInfo
				json.Unmarshal([]byte(userJson), &user)

				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				roomList = append(roomList, userlist)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "joinRoom",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"roomId":    roomid,
				"roomUsers": roomList,
			},
		}

		writeJSON(userid, response)

	case cmd == "inviteCall":

		var inviteCall InviteCall

		if err := mapstructure.Decode(data, &inviteCall); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(inviteCall)

		roomid := inviteCall.Data.RoomID

		//房间用户
		var roomList []UserListInfo = []UserListInfo{}
		{
			setList, _ := clientDB.SMembers(roomid).Result()
			for _, roomUser := range setList {
				userJson, _ := clientDB.HGet("ALLUSER", roomUser).Result()
				var user UserInfo
				json.Unmarshal([]byte(userJson), &user)

				var userlist UserListInfo
				userlist.ID = user.ID
				userlist.Name = user.Name
				userlist.Avatar = user.Avatar
				userlist.IsOnline = user.IsOnline

				roomList = append(roomList, userlist)
			}
		}

		//通知对方
		var busyList []UserListInfo = []UserListInfo{}
		{

			roomTemp := roomid + "_temp"
			clientDB.SAdd(roomTemp, inviteCall.Data.Inviter.ID)
			clientDB.SAdd("BUSYUSERS", inviteCall.Data.Inviter.ID)

			for _, data := range inviteCall.Data.Invitees {
				ok, _ := clientDB.SIsMember("BUSYUSERS", data.ID).Result()
				if ok {
					userJson, _ := clientDB.HGet("ALLUSER", data.ID).Result()
					var user UserInfo
					json.Unmarshal([]byte(userJson), &user)

					var userlist UserListInfo
					userlist.ID = user.ID
					userlist.Name = user.Name
					userlist.Avatar = user.Avatar
					userlist.IsOnline = user.IsOnline

					busyList = append(busyList, userlist)
					clientDB.SRem("BUSYUSERS", inviteCall.Data.Inviter.ID)
				} else {
					clientDB.SAdd(roomTemp, data.ID)
					clientDB.SAdd("BUSYUSERS", data.ID)

					notifyData := map[string]interface{}{
						"event": "onInviteCall",
						"data": map[string]interface{}{
							"roomId":    roomid,
							"callType":  inviteCall.Data.CallType,
							"inviter":   inviteCall.Data.Inviter,
							"invitees":  inviteCall.Data.Invitees,
							"roomUsers": roomList,
						},
					}

					//离线繁忙
					{
						_, ok := clients[data.ID]

						if !ok {
							userJson, _ := clientDB.HGet("ALLUSER", data.ID).Result()
							var user UserInfo
							json.Unmarshal([]byte(userJson), &user)

							var userlist UserListInfo
							userlist.ID = user.ID
							userlist.Name = user.Name
							userlist.Avatar = user.Avatar
							userlist.IsOnline = user.IsOnline

							busyList = append(busyList, userlist)
							clientDB.SRem("BUSYUSERS", inviteCall.Data.Inviter.ID)
						} else {
							writeJSON(data.ID, notifyData)
						}
					}

				}
			}

			clientDB.Expire(roomTemp, time.Second*120)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "inviteCall",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"roomId":    roomid,
				"roomUsers": roomList,
				"busyUsers": busyList,
			},
		}

		writeJSON(userid, response)

	case cmd == "updateApp":

		dataMap, _ := data["data"].(map[string]interface{})
		version, _ := dataMap["version"].(string)

		force, _ := clientDB.HGet("version", "isForceUpdate").Int()

		var isForceUpdate bool
		if force == 1 {
			isForceUpdate = true
		} else {
			isForceUpdate = false
		}

		newVersion, _ := clientDB.HGet("version", "newVersion").Result()
		desc, _ := clientDB.HGet("version", "desc").Result()
		url, _ := clientDB.HGet("version", "url").Result()

		num := compareVersions(version, newVersion)

		var hasNewVersion bool
		if num == -1 {
			hasNewVersion = true
		} else {
			hasNewVersion = false
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "updateApp",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"hasNewVersion": hasNewVersion,
				"isForceUpdate": isForceUpdate,
				"newVersion":    newVersion,
				"desc":          desc,
				"url":           url,
			},
		}

		writeJSON(userid, response)

	}
}
