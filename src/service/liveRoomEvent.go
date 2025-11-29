package httpService

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
)

func liveRoomEvent(userid string, data map[string]interface{}) {
	cmd, _ := data["event"].(string)

	switch {

	case cmd == "createLiveRoom":
		dataMap, _ := data["data"].(map[string]interface{})
		liveid, _ := dataMap["liveID"].(string)
		name, _ := dataMap["name"].(string)
		cover, _ := dataMap["cover"].(string)
		secret, _ := dataMap["secret"].(string)

		//检查房间
		{
			user := getUserInfo(userid)

			if user.LiveRoom != "" {
				asyncCloseRoom(user.LiveRoom)

				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       user.LiveRoom,
						"liveRoomStatus": Close,
					},
				}

				//清除状态
				{
					setList, _ := clientDB.SMembers(user.LiveRoom).Result()
					for _, id := range setList {
						var user UserInfo
						user.ViewerState = Clean
						user.UserState = Clean
						user.ArtistState = Clean
						user.LiveRoomStatus = Clean
						userCleanHSet(id, user)
					}
				}

				asyncAllNotify(user.LiveRoom, notifyData)

				clientDB.HDel("ALLLIVEROOM", user.LiveRoom)
				clientDB.Del(user.LiveRoom)

				roomCount := user.LiveRoom + "_count"
				clientDB.Del(roomCount)
			}
		}

		var room LiveRoomBase
		if liveid == "" {
			// room.LiveID = uuid.New().String()
			liveRoomNum += 1
			room.LiveID = strconv.Itoa(liveRoomNum)
		} else {
			room.LiveID = liveid
		}

		room.Name = name
		room.Cover = cover
		room.Secret = secret
		room.Artist = userid

		writeDB, _ := json.Marshal(room)
		clientDB.HSet("ALLLIVEROOM", room.LiveID, writeDB)

		//创建加入
		clientDB.SAdd(room.LiveID, userid)
		clientDB.Expire(room.LiveID, time.Second*86400)

		//房间计数
		roomCount := room.LiveID + "_count"
		clientDB.HSet(roomCount, "roomCount", 0)

		//初始状态
		setLiveStatus(room.LiveID, Nomal)

		//记录房间
		{
			var user UserInfo
			user.LiveRoom = room.LiveID
			userInfoHSet(userid, user)
		}

		//响应数据
		var roomres LiveRoom
		{
			artist := getUserBase(userid)
			user := getLiveStatus(room.LiveID)

			roomres.LiveID = room.LiveID
			roomres.Name = room.Name
			roomres.Cover = room.Cover
			roomres.Secret = room.Secret
			roomres.LiveRoomStatus = user.LiveRoomStatus
			roomres.Artist = artist
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "createLiveRoom",
			"sn":    int(sn),
			"data":  roomres,
		}

		writeJSON(userid, response)

	case cmd == "allLiveRoom":
		var liveList []LiveRoom
		hashGetAll, _ := clientDB.HGetAll("ALLLIVEROOM").Result()
		for _, userJson := range hashGetAll {
			var roomBase LiveRoomBase
			json.Unmarshal([]byte(userJson), &roomBase)

			artist := getUserBase(roomBase.Artist)
			user := getLiveStatus(roomBase.LiveID)

			var room LiveRoom
			room.LiveID = roomBase.LiveID
			room.Name = roomBase.Name
			room.Cover = roomBase.Cover
			room.Secret = roomBase.Secret
			room.LiveRoomStatus = user.LiveRoomStatus
			room.Artist = artist

			liveList = append(liveList, room)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "allLiveRoom",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"liveList": liveList,
			},
		}

		writeJSON(userid, response)

	case cmd == "joinLiveRoom":
		dataMap, _ := data["data"].(map[string]interface{})
		liveRoom, _ := dataMap["liveRoom"].(string)

		liveJson, _ := clientDB.HGet("ALLLIVEROOM", liveRoom).Result()
		if liveJson == "" {
			sn, _ := data["sn"].(float64)
			response := map[string]interface{}{
				"event": "joinLiveRoom",
				"sn":    int(sn),
				"data": map[string]interface{}{
					"liveRoom":       liveRoom,
					"userId":         userid,
					"liveRoomStatus": Close,
				},
			}
			writeJSON(userid, response)
			return
		}

		artist := getLiveArtist(liveRoom)
		//房间状态
		var roomStatus string
		{
			userJson, _ := clientDB.HGet("ALLUSER", artist).Result()
			var userDB UserInfo
			json.Unmarshal([]byte(userJson), &userDB)
			roomStatus = userDB.LiveRoomStatus
		}

		clientDB.SAdd(liveRoom, userid)

		//延时发送
		go func() {
			time.Sleep(2 * time.Second)
			onRoomViewerChange(artist, liveRoom)
		}()

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "joinLiveRoom",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"liveRoom":       liveRoom,
				"userId":         userid,
				"liveRoomStatus": roomStatus,
			},
		}

		writeJSON(userid, response)

	case cmd == "leaveLiveRoom":
		dataMap, _ := data["data"].(map[string]interface{})
		liveRoom, _ := dataMap["liveRoom"].(string)

		artist := getLiveArtist(liveRoom)
		roomCount := liveRoom + "_count"

		//关闭房间
		{
			if userid == artist {

				//清除状态
				{
					setList, _ := clientDB.SMembers(liveRoom).Result()
					for _, id := range setList {
						var user UserInfo
						user.ViewerState = Clean
						user.UserState = Clean
						user.ArtistState = Clean
						user.LiveRoomStatus = Clean
						userCleanHSet(id, user)
					}
				}

				asyncCloseRoom(liveRoom)

				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       liveRoom,
						"liveRoomStatus": Close,
					},
				}

				asyncAllNotify(liveRoom, notifyData)

				clientDB.HDel("ALLLIVEROOM", liveRoom)
				clientDB.Del(liveRoom)

				clientDB.Del(roomCount)

			} else {
				clientDB.SRem(liveRoom, userid)

				//刷新计数
				{
					user := getLiveStatus(liveRoom)
					if user.LiveRoomStatus == Online {
						clientDB.HIncrBy(roomCount, "roomCount", -1)
						count, _ := clientDB.HGet(roomCount, "roomCount").Int64()
						if count < 1 {
							//房间状态
							notifyData := map[string]interface{}{
								"event": "onLiveRoomStatus",
								"data": map[string]interface{}{
									"liveRoom":       liveRoom,
									"liveRoomStatus": Nomal,
								},
							}
							asyncAllNotify(liveRoom, notifyData)
							setLiveStatus(liveRoom, Nomal)
						}
					}
				}

				var user UserInfo
				user.ViewerState = Clean
				user.UserState = Clean
				user.ArtistState = Clean
				user.LiveRoomStatus = Clean
				userCleanHSet(userid, user)

				onRoomViewerChange(artist, liveRoom)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "leaveLiveRoom",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"liveRoom": liveRoom,
				"userId":   userid,
			},
		}

		writeJSON(userid, response)

	case cmd == "sendRoomMessage":
		dataMap, _ := data["data"].(map[string]interface{})
		liveRoom, _ := dataMap["roomId"].(string)
		content, _ := dataMap["content"].(string)

		//发言身份
		var msgType string
		{
			artist := getLiveArtist(liveRoom)

			if userid == artist {
				msgType = Artist
			} else {
				msgType = General
			}
		}

		//通知房间
		{
			userJson, _ := clientDB.HGet("ALLUSER", userid).Result()
			var user UserInfo
			if userJson != "" {
				json.Unmarshal([]byte(userJson), &user)
			}

			notifyData := map[string]interface{}{
				"event": "onRoomMessage",
				"data": map[string]interface{}{
					"roomId":  liveRoom,
					"content": content,
					"sender": map[string]interface{}{
						"id":     user.ID,
						"name":   user.Name,
						"avatar": user.Avatar,
						"type":   msgType,
					},
				},
			}

			asyncAllNotify(liveRoom, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "sendRoomMessage",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "queryRoomViewers":
		dataMap, _ := data["data"].(map[string]interface{})
		liveRoom, _ := dataMap["liveRoom"].(string)

		artist := getLiveArtist(liveRoom)

		//观众列表
		setList, _ := clientDB.SMembers(liveRoom).Result()

		var viewerList []UserListInfo
		for _, viewerId := range setList {
			if viewerId != artist {
				userJson, _ := clientDB.HGet("ALLUSER", viewerId).Result()
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
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "queryRoomViewers",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"users": viewerList,
			},
		}

		writeJSON(userid, response)

	case cmd == "applyConnection":
		dataMap, _ := data["data"].(map[string]interface{})
		liveRoom, _ := dataMap["roomId"].(string)

		//申请连线
		var user UserInfo
		user.UserState = "0"
		userInfoHSet(userid, user)

		userBase := getUserBase(userid)

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onApplyConnection",
				"data": map[string]interface{}{
					"roomId": liveRoom,
					"sender": userBase,
				},
			}

			artist := getLiveArtist(liveRoom)
			writeJSON(artist, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "applyConnection",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "answerConnection":

		var answer AnswerConnection

		if err := mapstructure.Decode(data, &answer); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(answer)

		otherId := answer.Data.UserID
		isAgree := answer.Data.IsAgree
		artist := getLiveArtist(answer.Data.RoomID)

		//记录连线
		var user UserInfo
		if isAgree == "true" {

			user.UserState = "1"

			//请求推流
			{
				if userid == artist {
					asyncSetUserRoomPermission(otherId, answer.Data.RoomID, true)
					setUserStatus(otherId, Online)
					userInfoHSet(otherId, user)
				} else {
					asyncSetUserRoomPermission(userid, answer.Data.RoomID, true)
					setUserStatus(userid, Online)
					userInfoHSet(userid, user)
				}
			}

			//房间计数
			{
				roomCount := answer.Data.RoomID + "_count"
				clientDB.HIncrBy(roomCount, "roomCount", 1)
			}

			//房间状态
			{
				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       answer.Data.RoomID,
						"liveRoomStatus": Online,
					},
				}
				asyncAllNotify(answer.Data.RoomID, notifyData)

				setLiveStatus(answer.Data.RoomID, Online)
			}
		} else {
			user.UserState = "2"

			//重置状态
			var userclean UserInfo
			userclean.UserState = Clean
			userCleanHSet(userid, userclean)
		}

		if userid == artist {
			userInfoHSet(otherId, user)
		} else {
			userInfoHSet(userid, user)
		}

		userBase := getUserBase(userid)

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onAnswerConnection",
				"data": map[string]interface{}{
					"roomId":  answer.Data.RoomID,
					"sender":  userBase,
					"isAgree": answer.Data.IsAgree,
				},
			}

			writeJSON(otherId, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "answerConnection",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "inviteConnection":
		dataMap, _ := data["data"].(map[string]interface{})
		roomId, _ := dataMap["roomId"].(string)
		otherId, _ := dataMap["userId"].(string)

		userBase := getUserBase(userid)

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onInviteConnection",
				"data": map[string]interface{}{
					"roomId": roomId,
					"sender": userBase,
				},
			}

			writeJSON(otherId, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "inviteConnection",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "getRoomOnline":
		dataMap, _ := data["data"].(map[string]interface{})
		stateType, _ := dataMap["type"].(string)
		liveRoom, _ := dataMap["liveRoom"].(string)

		artist := getLiveArtist(liveRoom)

		//连线列表
		setList, _ := clientDB.SMembers(liveRoom).Result()

		var viewerList []UserStateInfo
		for _, viewerId := range setList {
			if viewerId != "" && viewerId != artist {
				userJson, _ := clientDB.HGet("ALLUSER", viewerId).Result()
				if userJson != "" {
					var user UserInfo
					json.Unmarshal([]byte(userJson), &user)

					var userlist UserStateInfo
					userlist.ID = user.ID
					userlist.Name = user.Name
					userlist.Avatar = user.Avatar

					switch stateType {
					case "1":
						if user.UserState == "0" {
							userlist.State = user.ViewerState
							viewerList = append(viewerList, userlist)
						}
					case "2":
						if user.UserState == "0" {
							userlist.State = user.UserState
							viewerList = append(viewerList, userlist)
						}
					case "3":
						if user.UserState != "1" {
							userlist.State = user.ArtistState
							viewerList = append(viewerList, userlist)
						}
					}
				}
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "getRoomOnline",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"users": viewerList,
			},
		}

		writeJSON(userid, response)

	case cmd == "answerPk":

		var answer AnswerConnection

		if err := mapstructure.Decode(data, &answer); err != nil {
			log.Fatalf("Failed to decode: %v", err)
		}

		log.Print(answer)

		otherId := answer.Data.UserID
		isAgree := answer.Data.IsAgree

		if isAgree == "true" {

			asyncStartPk(userid, answer.Data.RoomID, answer.Data.RemoteRoomId)

			{
				//本端通知
				{
					notifyData := map[string]interface{}{
						"event": "onLiveRoomStatus",
						"data": map[string]interface{}{
							"liveRoom":       answer.Data.RoomID,
							"liveRoomStatus": PK,
						},
					}

					asyncAllNotify(answer.Data.RoomID, notifyData)
				}

				//远端通知
				{
					notifyData := map[string]interface{}{
						"event": "onLiveRoomStatus",
						"data": map[string]interface{}{
							"liveRoom":       answer.Data.RemoteRoomId,
							"liveRoomStatus": PK,
						},
					}

					asyncAllNotify(answer.Data.RemoteRoomId, notifyData)
				}

				//保存状态
				setLiveStatus(answer.Data.RoomID, PK)
				setLiveStatus(answer.Data.RemoteRoomId, PK)
			}
		}

		userBase := getUserBase(userid)

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onAnswerPk",
				"data": map[string]interface{}{
					"roomId":  answer.Data.RoomID,
					"sender":  userBase,
					"isAgree": answer.Data.IsAgree,
				},
			}

			writeJSON(otherId, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "answerPk",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "invitePk":
		dataMap, _ := data["data"].(map[string]interface{})
		roomId, _ := dataMap["roomId"].(string)
		otherRoom, _ := dataMap["remoteRoomId"].(string)
		otherId, _ := dataMap["userId"].(string)
		userBase := getUserBase(userid)

		//状态判断
		{
			user := getLiveStatus(otherRoom)
			if user.LiveRoomStatus == Online {
				sn, _ := data["sn"].(float64)
				response := map[string]interface{}{
					"event": "invitePk",
					"sn":    int(sn),
					"code":  500,
					"msg":   "The room is currently connect mic",
					"data":  map[string]interface{}{},
				}

				writeJSON(userid, response)
				return
			}
		}

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onInvitePk",
				"data": map[string]interface{}{
					"roomId": roomId,
					"sender": userBase,
				},
			}

			writeJSON(otherId, notifyData)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "invitePk",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "getPkUserList":
		var liveList []PkUser
		hashGetAll, _ := clientDB.HGetAll("ALLLIVEROOM").Result()
		for _, userJson := range hashGetAll {
			var room LiveRoomBase
			json.Unmarshal([]byte(userJson), &room)

			liveStatus := getLiveStatus(room.LiveID)
			if room.Artist != "" && room.Artist != userid && liveStatus.IsOnline == "true" {
				user := getUserBase(room.Artist)
				var pkuser PkUser
				pkuser.ID = user.ID
				pkuser.Name = user.Name
				pkuser.Avatar = user.Avatar
				pkuser.RoomID = room.LiveID

				setList, _ := clientDB.SMembers(room.LiveID).Result()
				pkuser.ViewerNum = len(setList) - 1

				//State
				{
					if liveStatus.LiveRoomStatus == PK {
						pkuser.State = "1"
					} else {
						pkuser.State = "0"
					}
				}

				liveList = append(liveList, pkuser)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "getPkUserList",
			"sn":    int(sn),
			"data": map[string]interface{}{
				"users": liveList,
			},
		}

		writeJSON(userid, response)

	case cmd == "kickOutLiveRoom":
		dataMap, _ := data["data"].(map[string]interface{})
		roomId, _ := dataMap["liveRoom"].(string)
		userIds, _ := dataMap["userIds"].([]interface{})

		asyncKickOutLiveRoom(userid, roomId, userIds)

		for _, data := range userIds {
			id, _ := data.(string)
			clientDB.SRem(roomId, id)
		}

		//通知变化
		{
			artist := getLiveArtist(roomId)
			onRoomViewerChange(artist, roomId)
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "kickOutLiveRoom",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "quitConnection":
		dataMap, _ := data["data"].(map[string]interface{})
		roomId, _ := dataMap["roomId"].(string)
		otherId, _ := dataMap["userId"].(string)

		asyncSetUserRoomPermission(otherId, roomId, false)

		userBase := getUserBase(otherId)

		//重置状态
		{
			var userclean UserInfo
			userclean.UserState = Clean
			userCleanHSet(userid, userclean)
		}

		//房间通知
		{
			notifyData := map[string]interface{}{
				"event": "onQuitConnection",
				"data": map[string]interface{}{
					"roomId": roomId,
					"user":   userBase,
				},
			}

			asyncNotify(userid, roomId, notifyData)
		}

		//房间计数
		{
			roomCount := roomId + "_count"
			clientDB.HIncrBy(roomCount, "roomCount", -1)
			count, _ := clientDB.HGet(roomCount, "roomCount").Int64()

			if count < 1 {
				//房间状态
				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       roomId,
						"liveRoomStatus": Nomal,
					},
				}
				asyncAllNotify(roomId, notifyData)

				setLiveStatus(roomId, Nomal)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "quitConnection",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)

	case cmd == "quitPk":
		dataMap, _ := data["data"].(map[string]interface{})
		roomId, _ := dataMap["roomId"].(string)
		otherId, _ := dataMap["userId"].(string)
		pKIds, _ := dataMap["pKUserIds"].([]interface{})

		asyncEndPk(userid, otherId, roomId)

		userBase := getUserBase(otherId)

		notifyData := map[string]interface{}{
			"event": "onQuitPk",
			"data": map[string]interface{}{
				"roomId": roomId,
				"user":   userBase,
			},
		}

		asyncNotify(userid, roomId, notifyData)

		//房间状态
		{
			pknum := len(pKIds)
			if pknum < 3 {
				for _, data := range pKIds {
					userid, _ := data.(string)
					user := getUserInfo(userid)

					if user.LiveRoom != "" {
						notifyData := map[string]interface{}{
							"event": "onLiveRoomStatus",
							"data": map[string]interface{}{
								"liveRoom":       user.LiveRoom,
								"liveRoomStatus": Nomal,
							},
						}
						asyncAllNotify(user.LiveRoom, notifyData)
						setLiveStatus(user.LiveRoom, Nomal)
					}
				}
			} else {
				notifyData := map[string]interface{}{
					"event": "onLiveRoomStatus",
					"data": map[string]interface{}{
						"liveRoom":       roomId,
						"liveRoomStatus": Nomal,
					},
				}
				asyncAllNotify(roomId, notifyData)

				setLiveStatus(roomId, Nomal)
			}
		}

		sn, _ := data["sn"].(float64)
		response := map[string]interface{}{
			"event": "quitPk",
			"sn":    int(sn),
			"data":  map[string]interface{}{},
		}

		writeJSON(userid, response)
	}
}
