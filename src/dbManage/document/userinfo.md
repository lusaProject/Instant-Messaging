用户管理交互

# 一、信令格式

wss://xxx.com/ws

### 发送格式

```JSON
{
    "event": "xx",
    "data": {
        "xx": "xx"
    },
    "sn": 123
}
```

### 响应正确格式

```JSON
{
    "event": "xx",
    "code": 200,
    "data": {
        "xx": "xx"
    },
    "sn": 123
}
```

### 响应错误格式

```JSON
{
    "event": "xx",
    "code": 500,
    "desc": "xx",
    "data": {},
    "sn": 123
}
```

### 通知格式

```JSON
{
    "event": "xx",
    "data": {
        "xx": "xx"
    },
}
```

**解释：**

event：表示一个行为事件。

data：表示行为数据。

sn：表示发起行为的序列化。

# 二、通信信令

### 查询全部用户

事件名：queryAllUsers

Request：

```JSON
{
    "event": "queryAllUsers",
    "data": {

    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "queryAllUsers",
    "data": {
        "users": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "isOnline": ture
            }
        ]
    },
    "sn": 1
}
```

### 添加好友

事件名：addFriend

Request：

```JSON
{
    "event": "addFriend",
    "data": {
        "myId": "xx",
        "friendId": "xx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "addFriend",
    "data": {
        "id": "xx",
        "name": "xx",
        "avatar": "xx"
    },
    "sn": 1
}
```

### 查询好友

事件名：queryFriends

Request：

```JSON
{
    "event": "queryFriends",
    "data": {
        "myId": "xx",
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "queryFriends",
    "data": {
        "applyNotifyNum": 1,
        "users": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "isOnline": ture
            }
        ]
    },
    "sn": 1
}
```

### 发送消息

事件名：message

Request：

```JSON
{
    "event": "message",
    "data": {
        "senderId": "xx",
        "receiverIds": ["a", "b"],
        "type": "xx",
        "content": "xx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "message",
    "data": {
        "senderId": "xx",
        "receiverIds": ["a", "b"],
        "type": "xx",
        "content": "xx"
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "message",
    "data": {
        "senderId": "xx",
        "receiverIds": ["a", "b"],
        "type": "xx",
        "content": "xx"
    }
}
```

### 更新用户

事件名：updateUser

Notify：

```JSON
{
    "event": "updateUser",
    "data": {
        "id": "xx",
        "name": "xx",
        "avatar": "xx",
        "isOnline": ture
    }
}
```

### 更新用户名称

事件名：updateUsername

Request：

```JSON
{
    "event": "updateUsername",
    "data": {
        "myId": "xx",
        "name": "xx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "updateUsername",
    "data": {},
    "sn": 1
}
```

### 更新应用

事件名：updateApp

Request：

```JSON
{
    "event": "updateApp",
    "data": {
        "version": "1.0.1"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "updateApp",
    "data": {
        "hasNewVersion": true,
        "isForceUpdate": true,
        "newVersion": "1.0.5",
        "desc": "xxxxx",
        "url": "https://xxx",
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "updateApp",
    "data": {
        "hasNewVersion": true,
        "isForceUpdate": true,
        "newVersion": "1.0.5",
        "desc": "xxxxxx",
        "url": "https://xxx",
    }
}
```

# 三、基于消息的信令

### 邀请通话

事件名：inviteCall

Request、Response和Notify：

```JSON
{
    "event": "message",
    "data": {
        "content": "",
        "receiverIds": [
            "36817cd458f2162e"
        ],
        "senderId": "96c6b83deeb39049",
        "type": "inviteCall"
    },
    "sn": "5"
}
```

### 退出通话

事件名：quitCall

Request、Response和Notify：

```JSON
{
    "event": "message",
    "data": {
        "content": "",
        "receiverIds": [
            "36817cd458f2162e"
        ],
        "senderId": "96c6b83deeb39049",
        "type": "quitCall"
    },
    "sn": "8"
}
```

### 更新通话

事件名：updateCall

Request、Response和Notify：

```JSON
{
    "event": "message",
    "data": {
        "content": "",
        "receiverIds": [
            "36817cd458f2162e"
        ],
        "senderId": "96c6b83deeb39049",
        "type": "updateCall"
    },
    "sn": "16"
}
```

# 四、用户鉴权

### 获取用户token

curl --location 'https://xxx.com/login' \

--header 'Content-Type: application/json' \

--header 'Authorization: Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq' \

--data '{

    "account":"DE9F09F1-6ACE-4F45-913E-268A6850888D",

    "passwd":"123456"

}'

响应数据

{

    "code": "0",

    "data": {

    "userToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50IjoiREU5RjA5RjEtNkFDRS00RjQ1LTkxM0UtMjY4QTY4NTA4ODhEIiwiZXhwaXJlc0RhdGUiOiIyNzI2NjQzNDk2In0.nPTVM-bOEgnXGNVuOEnDKer5G4h9Bl-pEG7GuKFvm3w"

    },

    "msg": "success"

}

### 获取用户SDK-Token

curl --location 'https://xxx.com/getSdkToken' \

--header 'Content-Type: application/json' \

--data '{

    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50IjoiREU5RjA5RjEtNkFDRS00RjQ1LTkxM0UtMjY4QTY4NTA4ODhEIiwiZXhwaXJlc0RhdGUiOiIyNzI2NjQzNDk2In0.nPTVM-bOEgnXGNVuOEnDKer5G4h9Bl-pEG7GuKFvm3w",

    "roomId": "1234"

}'

响应数据

{

    "code": 200,

    "data": {

    "appId": "888BCCD572D9ED023F69D479FFDB48C4",

    "sdkToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjczNTg5ODUsInVzZXJfaWQiOiJERTlGMDlGMS02QUNFLTRGNDUtOTEzRS0yNjhBNjg1MDg4OEQiLCJhcHBfaWQiOiI4ODhCQ0NENTcyRDlFRDAyM0Y2OUQ0NzlGRkRCNDhDNCJ9.eFl24KOr1cMX9SgvGpq3XOgH8rS_8R--orO9NZTRo8A"

    }

}

### 查询SDK-Token

事件名：querySdkToken

Request：

```JSON
{
    "event": "querySdkToken",
    "data": {
        "userToken": "xx",
        "roomName": "xx",
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "querySdkToken",
    "data": {
        "appId": "xx",
        "sdkToken": "xx",
        "roomName": "xx",
    },
    "sn": 1
}
```

### 好友申请列表

事件名：friendRequest

Request：

```JSON
{
  "event": "friendRequest",
  "data": {
        "myId": "xx",
  },
  "sn": 1
}
```

Response：

```JSON
{
  "event": "friendRequest",
  "data": {
    "requestList": [
      {
        "type": "invite(0)-invitees(1)",
        "state": "request(0)-agree(1)-reject(2)",
        "id": "xx",
        "name": "xx",
        "avatar": "xx"
      }
    ]
  },
  "sn": 1
}
```

### 申请状态更新

事件名：requestUpdate

Request：

```JSON
{
  "event": "requestUpdate",
   "data": {
        "myId": "xx",
        "friendId": "xx"
        "state": "request(0)-agree(1)-reject(2)",
   },
  "sn": 1
}
```

Response：

```JSON
{
  "event": "requestUpdate",
   "data": {
        "myId": "xx",
        "friendId": "xx"
        "state": "agree-reject",
   },
  "sn": 1
}
```

Notify：

```JSON
{
   "event": "onRequestUpdate",
   "data": {
        "myId": "xx",
        "friendId": "xx"
        "state": "agree-reject",
   }
}
```

### 上报日志

事件名：onReportLog

Notify：

```JSON
{
   "event": "onReportLog",
   "data": {
        "taskId": 1,
        "startTime": 123,
        "endTime": 456
   }
}
```

### 事件确认

事件名：eventAck

Notify：

```JSON
{
   "event": "eventAck",
   "data": {
        "event": "onReportLog",
        "taskId": 1
   }
}
```

# 五、直播业务

### 创建直播房间

事件名：createLiveRoom

Request：

```JSON
{
    "event": "createLiveRoom",
    "sn": 1234,
    "data": {
        "name": "match made in heaven",
        "liveID": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "cover": "https://xxx.com/avatar/147267.jpg",
        "secret": "147258369"
    }
}
```

Response：

```JSON
{
    "data": {
        "liveID": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "name": "match made in heaven",
        "cover": "https://data.devplay.cc/avatar/147267.jpg",
        "secret": "147258369",
        "liveRoomStatus": "nomal",
        "artist": {
            "id": "228414032e7be3a36ae7c98cf7f3f6690",
            "name": "lusa",
            "avatar": "https://xxx.com/avatar/147267.jpg"
        }
    },
    "event": "createLiveRoom",
    "sn": 1234
}
```

### 直播房间列表

事件名：allLiveRoom

Request：

```JSON
{
    "event": "allLiveRoom",
    "sn": 1234,
    "data": {}
}
```

Response：

```JSON
{
    "data": {
        "liveList": [
            {
                "liveID": "ae8b6f2e-e386-4890-a2de-e471d29f5cdd",
                "name": "match made in heaven",
                "secret": "147258369",
                "liveRoomStatus": "nomal",
                "artist": {
                    "id": "228414032e7be3a36ae7c98cf7f3f6690",
                    "name": "1111",
                }
            },
            {
                "liveID": "1d6759a6-0927-415d-8abd-c84a44e4c2d2",
                "name": "match made in heaven",
                "secret": "147258369",
                 "liveRoomStatus": "nomal",
                "artist": {
                    "id": "228414032e7be3a36ae7c98cf7f3f6690",
                    "name": "2222",
                }
            }
        ]
    },
    "event": "allLiveRoom",
    "sn": 1234
}
```

### 加入直播房间

房间状态

```JSON
enum LiveRoomStatus:String {
    case close
    case nomal
    case pk
    case online
    case offline
}
```

事件名：joinLiveRoom

Request：

```JSON
{
    "event": "joinLiveRoom",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "secret": "147258369",
    }
}
```

Response：

```JSON
{
    "data": {
       "liveRoom":  "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
       "userId":    "8bbde7fd-c8be-4938-6634-b8dac26a6bb1",
       "liveRoomStatus":LiveRoomStatus,
    },
    "event": "joinLiveRoom",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onLiveRoomStatus",
   "data": {
        "liveRoom": "xx",
        "isOffline": true,
        "liveRoomStatus": LiveRoomStatus
   }
}
```

### 离开直播房间

事件名：leaveLiveRoom

Request：(关闭房间和离开房间合并为一个接口)

```JSON
{
    "event": "leaveLiveRoom",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "leaveLiveRoom",
    "sn": 1234
}
```

### 房间发消息

消息类型

```JSON
enum msgType:String {
    case general
    case artist
}
```

事件名：sendRoomMessage

Request：

```JSON
{
    "event": "sendRoomMessage",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "content": "chat message",
 
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "sendRoomMessage",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onRoomMessage",
   "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "content": "chat message",
          "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "type": "general",
          }
    }
}
```

### 关闭房间请求

Request：(关闭房间和离开房间合并为一个接口)

```JSON
{
    "event": "closeLiveRoom",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "closeLiveRoom",
    "sn": 1234
}
```

### 踢人直播模式

Request：

```JSON
{
    "event": "kickOutLiveRoom",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "userIds":["uid001","uid002"]
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "kickOutLiveRoom",
    "sn": 1234
}
```

### **权限控制**

Request：

```JSON
{
    "event": "setUserRoomPermission",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "userId": "uid001",
          "permissions": "true | false"
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "setUserRoomPermission",
    "sn": 1234
}
```

### 查询在线观众列表

Request：

```JSON
{
    "event": "queryRoomViewers",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1"
    }
}
```

Response：

```JSON
{
    "data": {
        "users": [
            {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          },
          {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
        ]
    },
    "event": "queryRoomViewers",
    "sn": 1234
}
```

### 直播人数变化通知

每隔2秒检查是否有变化，有变化后给终端发通知

Notify：

```JSON
{
   "event": "onRoomViewerChange",
   "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "total": 100,
          // 最近进入房间的2人
          "recentList": [
              {
                "id": "xx",
                "name": "xx",
                "avatar": "xx"
              },
              {
                "id": "xx",
                "name": "xx",
                "avatar": "xx"
              }
          ]
    }
}
```

### 连线列表

Request：

```JSON
{
    "event": "getRoomOnline",
    //type: 1 观众端当前等待连线列表，2 主播端待处理连线列表，3 主播端可以邀请连线列表
    "type": "1",
    "sn": 1234,
    "data": {
          "liveRoom": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1"
    }
}
```

Response：

// type: 1 观众端当前等待连线列表，2 主播端待处理连线列表，3 主播端可以邀请连线列表

// type == 1 时,  state 用户状态,  0为无状态

// type == 2 时,  state 用户状态,  0为申请中，1为已同意，2为已拒绝

// type == 3 时,  state 用户状态,  0为未邀请，1为已邀请

```JSON
{

    "data": {
        "users": [
            {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "state": "1" 
          },
          {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "state": "1" 
          }
        ]
    },
    "event": "getRoomOnline",
    "sn": 1234
}
```

### 发起连线申请

事件名：applyConnection

Request：

```JSON
{
    "event": "applyConnection",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "applyConnection",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onApplyConnection",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
   }
}
```

### 同意或拒绝连线申请

事件名：answerConnection

Request：

```JSON
{
    "event": "answerConnection",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 对方用户Id
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac25645565",
          "isAgree": "true|false"
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "answerConnection",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onAnswerConnection",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          },
        "isAgree": "true|false"
   }
}
```

### 邀请连线

事件名：inviteConnection

Request：

```JSON
{
    "event": "inviteConnection",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac265a5d4",
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "inviteConnection",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onInviteConnection",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
   }
}
```

### 退出连线

事件名：quitConnection

Request：

```JSON
{
    "event": "quitConnection",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 主播操作的是让断开连线的人，观众操作只能是连线的自己，服务端有个权限判断
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac265a5d4", 
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "quitConnection",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onQuitConnection",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
   }
}
```

### 同意或拒绝PK

事件名：answerPk

Request：

```JSON
{
    "event": "answerPk",
    "sn": 1234,
    "data": {
          // 我的房间
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 对端房间
          "remoteRoomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 对方的用户ID
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac265a5d4",
          "isAgree": "true|false"
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "answerPk",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onAnswerPk",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          },
        "isAgree": "true|false"
   }
}
```

### 邀请Pk

事件名：invitePk

Request：

```JSON
{
    "event": "invitePk",
    "sn": 1234,
    "data": {
          // 我的房间
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 对端房间
          "remoteRoomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          // 对方的用户ID
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac265a5d4",
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "invitePk",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onInvitePk",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "sender": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
   }
}
```

### 退出Pk

事件名：quitPk

Request：

```JSON
{
    "event": "quitPk",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
          "userId": "8bbde7fd-c8be-4938-b98a-b8dac265a5d4",
          "pKUserIds":["uid001","uid002"]
    }
}
```

Response：

```JSON
{
    "data": {},
    "event": "quitPk",
    "sn": 1234
}
```

Notify：

```JSON
{
   "event": "onQuitPk",
   "data": {
        "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
          }
   }
}
```

### 获取PK用户列表

事件名：getPkUserList

Request：

```JSON
{
    "event": "getPkUserList",
    "sn": 1234,
    "data": {
          "roomId": "8bbde7fd-c8be-4938-b98a-b8dac26a6bb1"
    }
}
```

Response：

```JSON
{

    "data": {
        "users": [
            {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "roomId": "xx",
            "viewerNum": 99,
            "state": "0"  // 0未邀请； 1已邀请
          },
          {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "roomId": "xx",
            "viewerNum": 99,
            "state": "1" 
          }
        ]
    },
    "event": "getPkUserList",
    "sn": 1234
}
```
