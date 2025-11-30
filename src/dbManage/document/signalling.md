房间管理交互

# 一、信令格式

*wss://xxx.com/websocket*

### 连接头Headers

```JSON
{
    "userId":"xx" , 
    "sdkToken": "xx"
    "appId":"xx"
} 
```

**备注**:

### 发送格式

```JSON
{
    "event": "xx",
    "data": {
        "xx": "xx"
    },
    "sn": 123,
    "time": 15745567667555, // 任务发送时间
    "valid": 30000 // 单位 ms 
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
    "sn": 123,
    "time": 15745567667555,
    "valid": 30000 // 单位 ms 
}
```

### 通知格式

```JSON
{
    "event": "xx",
    "data": {
        "xx": "xx"
    },
    "time": 15745567667555,
    "valid": 30
}
```

**解释：**

event：表示一个行为事件。

data：表示行为数据。

sn：表示发起行为的序列化，由终端发起方赋值，服务端响应时带回，服务端通知时去除。

time：任务发送时间戳。

valid：任务有效时长，单位秒。

### **错误 Response 格式**

```Plain
enum MessageCode {
    // 正常消息
    success = 200,
    // 验证失败
    authInvalid = 401,
    // 消息过期
    msgExpired = 408,
    // 认证处理异常
    authError = 500,
    // ws正常断开状态码  应用端收到该状态码不重连
    wsClosedSuccess = 1000,
    // 主体异常 客户端提示用户检查主体是否正常启用
    maintainerError = 2001,
    // 应用ID (appId) 被关闭，客户端提示检查应用是否正常启用
    applicationError = 2002,
    // 应用ID不存在
    applicationNotExist = 2003,
    // 过期
    tokenExpired = 2011,
    // userId和token里面的UserId不一致
    illegalUser = 2012,
    // 加入房间错误
    joinError = 2013,
    // 连接CF错误
    cfError = 2015,
    // 发布错误
    pubError = 2016,
    // 订阅错误
    subError = 2017,
    // 关闭轨道错误
    clsError = 2018,
    //直播间通用错误
    liveError = 2019,
    // 其他通用错误
    commonError = 2020,
}
```

# 二、通信信令

### 创建房间

事件名：createRoom

Request：

```JSON
{
    "event": "createRoom",
    "data": {
        "callType": "P2P | Group",
        "creator": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1
        },
        "invitees": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx"
            }
        ]
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "createRoom",
    "data": {
        "roomId": "xx",
        "roomUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callAction": 1
                "callState": 1
            }
        ],
        "busyUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callState": 3
            }
        ]
    },
    "sn": 1
}
```

**解释：**

callState：表示一个用户的状态：0（无）、1（已加入）、2（通话中）、3（被邀请中）。

callAction：表示当前通话的行为：

- callAction&1：二进制第1位为0（语音不可用）、二进制第1位为1（语音可用）。
- callAction&2：二进制第2位为0（视频不可用）、二进制第2位为1（视频可用）。
- callAction&4：二进制第3位为0（用户没有说话）、二进制第3位为1（用户正在说话）。
- callAction&8：二进制第4位为0（用户语音没有被禁言）、二进制第4位为1（用户语音被禁言）。
- callAction&16：二进制第5位为0（用户视频没有被禁言）、二进制第5位为1（用户视频被禁言）。

### 加入房间

事件名：joinRoom

Request：

```JSON
{
    "event": "joinRoom",
    "data": {
        "roomId": "xx",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1
        }
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "joinRoom",
    "data": {
        "roomId": "xx",
        "roomUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callAction": 1,
                "callState": 1
            }
        ],
        "callType": 0 | 1 | 2
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onJoinRoom",
    "data": {
        "roomId": "xx",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1,
            "callState": 1
        }
    }
}
```

### 邀请通话

事件名：inviteCall

Request：

```JSON
{
    "event": "inviteCall",
    "data": {
        "roomId": "xx",
        "callType": "P2P | Group"
        "inviter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1,
        },
        "invitees": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx"
            }
        ]
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "inviteCall",
    "data": {
        "roomId": "xx",
        "roomUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callAction": 1,
                "callState": 1
            }
        ],
        "busyUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callState": 3
            }
        ]
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onInviteCall",
    "data": {
        "roomId": "xx",
        "callType": "P2P | Group",
        "inviter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1
        },
        "invitees": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx"
            }
        ]
        "roomUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callAction": 1,
                "callState": 1
            }
        ]
    }
}
```

### 邀请过期

事件名：onInviteExpired

Notify：

```JSON
{
    "event": "onInviteExpired",
    "data": {
        "roomId": "xx",
        "callType": "P2P | Group",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
        }
    }
}
```

### 拒绝通话

事件名：rejectCall

Request：

```JSON
{
    "event": "rejectCall",
    "data": {
        "roomId": "xx",
        "inviter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
        },
        "rejecter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
        }
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "rejectCall",
    "data": {},
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onRejectCall",
    "data": {
        "roomId": "xx",
        "inviter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
        },
        "rejecter": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx"
        }
    }
}
```

### 更新通话

事件名：updateCall

Request：

```JSON
{
    "event": "updateCall",
    "data": {
        "roomId": "xx",
        "actionType": "AudioEnable | VideoEnable | Speaking | AudioMute | VideoMute",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1
        } 
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "updateCall",
    "data": {},
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onUpdateCall",
    "data": {
        "roomId": "xx",
        "actionType": "AudioEnable | VideoEnable | Speaking | AudioMute | VideoMute",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1
        }
    }
}
```

### 连接Cloudflare

事件名：connectCF

Request：

```JSON
{
    "event": "connectCF",
    "data": {
        "sdp": "xx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "connectCF",
    "data": {
        "remoteSdp": "xx",
        "subscribeUsers": [
            {
                "id": "xx",
                "name": "xx",
                "avatar": "xx",
                "callAction": 1,
                "callState": 1
            }
        ]
    },
    "sn": 1
}
```

### 发布流

事件名：publish

Request：

```JSON
{
    "event": "publish",
    "data": {
        "sdp": "xx",
        "tracks": [
            {
                "mid": "xx",
                "trackName": "xx",
                "location": "local"
            }
        ]
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "publish",
    "data": {
        "remoteSdp": "xx"
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onPublish",
    "data": {
        "roomId": "xx",
        "user": {
            "id": "xx",
            "name": "xx",
            "avatar": "xx",
            "callAction": 1,
            "callState": 2
        }
    }
}
```

### 订阅流

事件名：subscribe

Request：

```JSON
{
    "event": "subscribe",
    "data": {
        "users": [{
        id:"用户id",
        tracks:[
          {mid:"mid1",trackName:"轨道1",type:"类型必传"}
        ]}]
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "subscribe",
    "data": {
        "remoteSdp": "xx",
        "tracks": [
            {
                "mid": "xx",
                "trackName": "xx",
                "location": "local | remote",
                "userId": "xx"
            }
        ]
    },
    "sn": 1
}
```

### **停止订阅流（关闭轨道）**

事件名：closeTrack

Request：

```JSON
{
    "event": "closeTrack",
    "data": {
        "tracks": [{mid:"必传",userId:"必传",trackName:"必传"}],
        "sdp":"xxx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "closeTrack",
    "data": {
        "remoteSdp": "xx",
        "tracks": [
            {
                "mid": "xx",
                "trackName": "xx",
                "location": "local | remote",
            }
        ]
    },
    "sn": 1
}
```

Notify:

```JSON
{
    "compress": false,
    "sn": 0,
    "event":"onCloseTrack",
    "data": {
        "user": {
            "id": "",
            "callAction": 1,
            "callState":1,
            "joinTime":0,
            "tracks":[
                {
                "mid":"客户端的原始mid",
                "type":0, // 1 | 2 
                "trackName":"原始name"
                }
            ]
        }
    }
}
```

### 协商流

事件名：renegotiate

Request：

```JSON
{
    "event": "renegotiate",
    "data": {
        "sdp": "xx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "renegotiate",
    "data": {},
    "sn": 1
}
```

### 退出房间

事件名：quitRoom

Request：

```JSON
{
    "event": "quitRoom",
    "data": {
        "roomId": "xx",
        "userId":"xxxx"
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "quitRoom",
    "data": {},
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onQuitRoom",
    "data": {
        "roomId": "xx",
        "user": {
            "id": "xx", 
        },
        "code": QuitCode 
        "desc": "提示描述"
    }
}
```

onQuitRoom 的退出原因以及描述 (提示用)

```JSON
Enum QuitCode = {
    0 = "正常退出" // 直播模式下,为非主播（有发流权限）的人正常退出 
    2021 = "心跳超时", // 直播模式下,为非主播（有发流权限）的人心跳超时
    2011 = "房间Token过期" 
    2022 = "被剔出房间(直播模式下)"
    2023 = "PK结束对方主播退出(直播模式)"
    2024 = "应用端销毁直播间" // endLive
}
```

### 销毁房间

事件名：destroyRoom (直播模式下,主播心跳超时，退出, token过期 的通知事件)

销毁事件通知 和 onQuitRoom 通知用的是同一个状态码（QuitCode）

Request：

```JSON
{
    "event": "destroyRoom",
    "data": {
        "roomId": "xx",
        "code": QuitCode, // 2011 | 2024 
        "desc" :""
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "destroyRoom",
    "data": {
    },
    "sn": 1
}
```

Notify：

```JSON
{
    "event": "onDestroyRoom",
    "data": {
        "roomId": "xx",
        "code": QuitCode, // 2011 | 2024 
        "desc" :""  
    }
}
```

### 通话心跳

事件名：callHeartbeat

Request：

```JSON
{
    "event": "heartbeat"
}
```

**解释：**

通话心跳用户在加入房间后，退出房间前的期间，定期（10s）向服务器发送通话心跳，服务器超时（30s）后，通知用户退出事件。服务器通过维护在通话中的状态，拒绝他人的通话邀请等。

### 调整通话音量

事件名: **updateVolume**

备注: **目前不需要该接口**

Request：

```JSON
{
    "event": "updateVolume",
    "data": {
        "roomId": "xx",
        "userId": ""  ,
        "volume": 80 //音量大小
    },
    "sn": 1,
    //时间戳...
}
```

Response:

```JSON
{
    "event": "updateVolume",
    "data": {},
    "sn": 1,
    //时间戳...
}
```

Notify:

```JSON
{
    "event": "onUpdateVolume",
    "data": {
        "roomId": "xx",
        "userId": ""  ,
        "volume": 80 //音量大小
    }
    //时间戳...
}
```

### 网络质量监测

事件名: **networkQualityChange**

Request:

```JSON
{
    "event": "networkQualityChange",
    "data": {
        "roomId": "xx",
        "egress": 0 // 上行网络质量  quality 的值 0 - 5 
        "ingress": 0 // 下行网络质量
    },
    "sn": 1,
    //时间戳...
}
```

Response:

```JSON
{
    "event": "networkQualityChange",
    "data": {},
    "sn":1,
    "code":200
    //时间戳
}
```

Notify:

```JSON
{
    "event": "onNetworkQualityChange",
    "data": {
         "roomId": "xx",
         "userId": ""  , //谁的质量转换了
         "egress": 0 // 上行网络质量 
         "ingress": 0 // 下行网络质量
    },
    //时间戳
}
```

### 多人网络质量监测

事件名: **onNetQuality**

Notify:

```JSON
{
    "event": "onNetQuality",
    "data": {
          "qualities": [
             {
                 "roomId": "xx",
                 "userId": ""  , //谁的质量转换了
                 "egress": 0 // 上行网络质量 
                 "ingress": 0 // 下行网络质量
             }
          ]
    },
    //时间戳
}
```

### 房间 Token 即将过期

事件名: **onBeforeTokenExpire**

Notify: 通知的是连接对象本身

备注：即将退出和重新获取token的操作提示

```JSON
{
   "event": "onBeforeTokenExpire", 
   "data":{
      "roomId":""
      "userId":"" 
   },
   "time":0 //时间戳  
   "valid":30
}
```

### 房间 Token 已过期

事件名：**onTokenExpired**

Notify: 通知的是连接对象本身

```JSON
{
   "event": "onTokenExpired", //
   "data":{
       "roomId":""
       "userId":"" 
   },
   "time":0 //时间戳 时间戳ms
   "valid":30 // s
}
```

### 房间 Token 更新

事件名: updateRoomToken

Request

```JSON
{
    "event":"updateRoomToken",
    "data":{
        roomId:"",
        sdkToken:"", //新的 token
    },
    "time":0,
    "valid": 30, // 单位秒
    "sn":1,
}
```

Response:

```JSON
{
 "event":"updateRoomToken",
  "data":{
      "token":""
  },
  "code"200,
  "time":0,
  "valid": 30 // 单位秒
}
```

### 更新服务器时间

Request

```JSON
{
    "event":"updateSeverTime",
}
```

Response

```JSON
{
    "event":"updateSeverTime",
    "data":{
        "time":1727594658793
    },
    "code":200
}
```

### 房间信息同步

事件名：syncRoomInfo

Request：

```JSON
{
    "event": "syncRoomInfo",
    "data": {
        "roomId": "xx",
        "user": {
            "id": "xx",
            "callAction": 1,
            "callState": 1
        } 
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "syncRoomInfo",
    "data": {
        "isValid": true,
        "roomId": "xx",
        "users": [
            {
                "id": "xx",
                "callAction": 1,
                "callState": 1
            }
        ]
    },
    "sn": 1
}
```

### 统一信息上报

事件名：reportInfo

Request：

```JSON
{
    "event": "reportInfo",
    "data": {
        "actionEvent": "xx",
        "userId": "123456",
        "platform": "android | ios | web",
        "time":1727594658793,
        "actionData": "json数据" 
    },
    "sn": 1
}
```

Response：

```JSON
{
    "event": "reportInfo",
    "data": {},
    "sn": 1
}
```

# 三、信息上报详情

### 上报轨道无数据

actionEvent：trackNoData

actionData

```JSON
{
  "roomId": "5a9c172d-2bbf-45db-a9cd-d01b942d572d",
  "trackType": "Microphone",
  "trackName": "zlEGlaTyEDpybnlp",
  "mid": "3"
}
```

# 四、直播业务

### 房间销毁通知

事件名：onDestoryRoom (应用服务端请求关闭房间后,客户端会收到来自信令的通知)

Notify:

```JSON
{
    "event": "onDestoryRoom",
    "code"：200,
    "data": {
        "roomId": "xx",
        "userId": "xxx" // 关闭人的 id
    },
    "sn": 1
}
```

### 跨房间转流(PK)

事件名: roomPk

Request:

```JSON
{
    event: "roomPk",
    data:{
        callType: CallType, // 共用 JoinReq
        roomId:"new Room Id",
        user: User // pk发起方主播的状态
    }
}
```

Response:

```JSON
{
    event: "roomPk",
    code:200 | MessageCode.liveError,
    data:{
        roomId:"",
        // 对方多个房间主播的状态和轨道
        users:[
        {
           id:"remote1"
           callAction:1
           tracks:[]
        },
        {
           id:"remote2"
           callAction:2
           tracks:[]
        },
        ]
    }
}
```

Notify:

```JSON
//新房间其他用户接收到 onPublish
{
   event:"onJoinRoom & onPublish", //连续收到onjoinRoom和 onPublish事件
   data:{
       //远端进来主播的状态
       user:{
           id:"remote1"
           callAction:1
           tracks:[]
       },
       roomId:"roomId"
   }
}
```

### 切换直播间

事件名: toggleliveRoom

Request:

```JSON
{
    event: "toggleliveRoom",
    data:{
        roomId:"",
    }
}
```

Response:

```JSON
{
    event: "toggleliveRoom",
    code: 200 | 2019
    data:{
        roomId:"新房间ID",
        hash:"" //新房间的会话HASH值,后续重连不用Token认证，需要在原来连接请求头配置基础上添加Hash字段 
        roomUsers:User[] 
    }
}
```

### 更新推流权限

事件名: onUpdatePermissions

该事件是应用服务端发送请求，信令给客户端的通知.

Notify:

```JSON
{
    event: "onUpdatePermissions",
    code: 200 
    data:{
        roomId:"",
        user:{
            id:"",
            permissions: 1 | 3, //1没有 3有
            tracks:[]
        }
    }
}
```

# 五、应用服务端请求接口

### 关闭房间请求 (直播模式)

Request: *(应用服务端请求)*

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/${appId}/${userId}/${roomId}/closeRoom' \
--header 'Content-Type: application/json' \
'
```

Response:(应用服务端)

```JSON
{
    "code": 0 | -1,
    "data": null,
    "msg":  "successDescrirption" | "errorDescrirption"
}
```

Notify: (客户端)

```JSON
{
    "event": "onDestoryRoom",
    "data": {
        "roomId":""
    },
    "sn":  0
}
```

### 开始PK请求 (直播模式)

Request:

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/${appId}/${userId}/${roomId}/startPk' \
--header 'Content-Type: application/json' \
--data '{
    remoteRoomId:"",
}'
```

Response:

```JSON
{
    "code": 0 | -1,
    "data": null,
    "msg":  "successDescrirption" | "errorDescrirption"
}
```

Notify(客户端):

```JSON
{
    event: "onPublish",
    code:200 | MessageCode.liveError,
    data:{
        roomId:"",
        users:{
           id:"remote1"
           callAction:1
           tracks:[]
        },
    }
}
```

### 结束PK请求 (直播模式)

Request:

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/${appId}/${userId}/${roomId}/endPk' \
--header 'Content-Type: application/json' \
--data '{
    roomId:"",
    remoteId:"",//结束Pk的用户Id, 如果是自己的ID,结束所有人的对话，如果是其他人的ID,结束和该用户的对话
}'
```

Respone:

```JSON
{
    "code": 0 | -1,
    "data": null,
    "msg":  "successDescrirption" | "errorDescrirption"
}
```

Notify:(客户端)

收到 onQuitRoom 事件，以及对应的退出原因

### 踢人 (直播模式)

Request:(应用服务端)

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/{appId}/${userId}/${roomId}/kickout' \
--header 'Content-Type: application/json' \
--data '{
    "users":["uid001","uid002"] //用户id列表
}'
```

Response:(应用服务端)

```JSON
{
    "code": 0 | -1,
    "data": null,
    "msg":  "successDescrirption" | "errorDescrirption"
}
```

Notify（客户端）: 见 onQuitRoom

### **发流单人权限控制 (直播)**

Request: (应用服务端http请求)

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/{appId}/${userId}/${roomId}/permission' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq' \
--data '{ 
    "pubPermission": true | false
}'
```

Response: (应用服务端)

```JSON
{
    "code": 0 | -1,
    "data": null,
    "msg": "successDescrirption" | "errorDescrirption"
}
```

### **发流多人权限控制 (直播)**

Request:(应用服务端请求)

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/{appId}/sys/${roomId}/permissions' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer cf_qzIFKgPbmvAlfuGdlkpOKvSBnVlrirHeAq' \
--data '{
    users:["uid001","uid002"],
    pubPermission: true | false 
}'
```

Response:(同上)

### 获取直播间列表

Request

```JSON
 curl -X POST --location 'https://workers.devplay.cc/v1/{appId}/live_list' \
--header 'Content-Type: application/json' \
--header 'Authorization: AppSecret' \
```

Response

```JSON
{
    "code": 200,
    "data": [
        {
            "liveId": "ad46f957-a651-4ec1-b934-07a56d4f4d22",
            "creator": "2fdb77f70f1c334709b556c872fdd9f28"
        },
         {
            "liveId": "ad46f957-a651-4ec1-b934-07a56d4f4d22",
            "creator": "2fdb77f70f1c334709b556c872fdd9f28"
        },
    ]
}
```
