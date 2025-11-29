package httpService

import (
	"sync"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var clientDB *redis.Client

var clients = make(map[string]*websocket.Conn)

var mu sync.Mutex

var mu_ws sync.Mutex

var liveRoomNum int = 12345678

const (
	CloudFlare string = "https://x.com"
)

const (
	AppId  string = "*************"
	AppKey string = "*************"
)

const (
	Close   string = "close"
	Nomal   string = "nomal"
	PK      string = "pk"
	Online  string = "online"
	Offline string = "offline"
	Clean   string = "clean"
)

const (
	Artist  string = "artist"
	General string = "general"
)

type UserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	IsOnline string `json:"isOnline"`

	ViewerState string `json:"viewerState"`
	UserState   string `json:"userState"`
	ArtistState string `json:"artistState"`

	LiveRoom       string `json:"liveRoom"`
	LiveRoomStatus string `json:"liveRoomStatus"`

	Timestamp string `json:"timestamp"`
}

type UserListInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	IsOnline string `json:"isOnline"`
}

type UserStateInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	State  string `json:"state"`
}

type UserBaseInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type InviteCall struct {
	Event string `json:"event"`
	Data  struct {
		RoomID   string `json:"roomId"`
		CallType string `json:"callType"`
		Inviter  struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Avatar     string `json:"avatar"`
			CallAction int    `json:"callAction"`
		} `json:"inviter"`
		Invitees []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"invitees"`
	} `json:"data"`
	Sn int `json:"sn"`
}

type JoinRoom struct {
	Event string `json:"event"`
	Data  struct {
		RoomID string `json:"roomId"`
		User   struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Avatar     string `json:"avatar"`
			CallAction int    `json:"callAction"`
		} `json:"user"`
	} `json:"data"`
	Sn int `json:"sn"`
}

type OnInviteExpired struct {
	Event string `json:"event"`
	Data  struct {
		RoomID   string `json:"roomId"`
		CallType string `json:"callType"`
		User     struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"user"`
	} `json:"data"`
}

type RejectCall struct {
	Event string `json:"event"`
	Data  struct {
		RoomID  string `json:"roomId"`
		Inviter struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"inviter"`
		Rejecter struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"rejecter"`
	} `json:"data"`
	Sn int `json:"sn"`
}

type QuitRoom struct {
	Event string `json:"event"`
	Data  struct {
		RoomID string `json:"roomId"`
		User   struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"user"`
	} `json:"data"`
	Sn int `json:"sn"`
}

type RequestList struct {
	Type   int    `json:"type"`
	State  int    `json:"state"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type LiveRoomBase struct {
	LiveID string `json:"liveID"`
	Name   string `json:"name"`
	Cover  string `json:"cover"`
	Secret string `json:"secret"`
	Artist string `json:"artist"`
}

type LiveRoom struct {
	LiveID         string `json:"liveID"`
	Name           string `json:"name"`
	Cover          string `json:"cover"`
	Secret         string `json:"secret"`
	LiveRoomStatus string `json:"liveRoomStatus"`
	Artist         struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"artist"`
}

type AnswerConnection struct {
	Event string `json:"event"`
	Sn    int    `json:"sn"`
	Data  struct {
		RoomID       string `json:"roomId"`
		UserID       string `json:"userId"`
		IsAgree      string `json:"isAgree"`
		RemoteRoomId string `json:"remoteRoomId"`
	} `json:"data"`
}

type PkUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	RoomID    string `json:"roomId"`
	ViewerNum int    `json:"viewerNum"`
	State     string `json:"state"`
}

type CfLiveList struct {
	Code int    `json:"code"`
	Desc string `json:"desc"`
	Data []struct {
		LiveID  string `json:"liveId"`
		Creator string `json:"creator"`
	} `json:"data"`
}
