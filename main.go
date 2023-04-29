package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type user struct {
	name  string
	pwd   string
	token string
}
type userMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	Token    string `json:"token"`
}

var users []*user = []*user{
	{
		name: "yxlr",
		pwd:  "114514",
	},
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var history string

const d = "{\n    \"users\": {\n        \"用户名\":\"密码\",\n        \"用户名2\" : \"密码2\"\n    }\n}"

func init() {
	_, err := os.Lstat("./users.json")
	if err != nil {
		defer os.Exit(100)
		f, err := os.Create("./users.json")
		if err != nil {
			log.Println("创建配置文件错误：", err)
		}
		_, err = f.Write([]byte(d))
		if err != nil {

			log.Println("写入默认配置错误：", err)
		}
		log.Println("请填写配置文件: " + "./users.json")
	}
	data, err := os.ReadFile("./users.json")
	if err != nil {
		log.Println("配置文件读取失败：", err)
		os.Exit(200)
	}
	json := gjson.Get(string(data), "users").Map()
	users = make([]*user, len(json)+1)
	var i = 1
	users[0] = &user{name: "yxlr",pwd: "1145141919810homo"}
	for v , s := range json {
        users[i] = &user{name: v,pwd: s.String()}
		i++
	}
}

func handleWebSocket(c *gin.Context) {
	// 升级HTTP连接为WebSocket协议
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// 将新连接添加到客户端列表中
	clients[conn] = true

	// 无限循环读取从客户端发来的消息
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			return
		}
		x, err := tokenBool(p)
		if err != nil {
			log.Println(err)
			continue
		}

		// 将消息广播给所有连接的客户端
		for c := range clients {
			err := c.WriteMessage(messageType, x)
			if err != nil {
				log.Println(err)
				delete(clients, c)
				return
			}
		}
	}
}

func serveHome(c *gin.Context) {
	// 渲染HTML模板
	c.HTML(200, "index.html", gin.H{
		"history": history,
	})
}

func login(c *gin.Context) {
	userName := c.Query("name")
	userPwd := c.Query("pwd")
	if userName == "" || userPwd == "" {
		c.JSON(200, gin.H{
			"err":  "密码或账号为空",
			"uuid": nil,
		})
		return
	}
	for i, v := range users {
		if userName == v.name && userPwd == v.pwd {
			users[i].token = uuid.New().String()

			c.JSON(200, gin.H{
				"err":  nil,
				"uuid": users[i].token,
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"err":  "账号或密码错误",
		"uuid": nil,
	})
}

func main() {
	router := gin.Default()

	// 挂载WebSocket处理函数
	router.GET("/ws", handleWebSocket)

	// 挂载HTML文件
	router.LoadHTMLGlob("public/*")
	router.GET("/", serveHome)

	router.GET("/login", login)

	router.GET("/history", func(c *gin.Context) {
		c.String(200, history)
	})

	router.GET("/clear", func(c *gin.Context) {
		history = ""
		c.String(200, "聊天记录已全部删除")
	})

	err := router.Run(":8888")
	if err != nil {
		log.Fatal(err)
	}
}

func tokenBool(s []byte) ([]byte, error) {
	var data *userMessage
	var p string
	err := json.Unmarshal(s, &data)
	log.Println(data)
	if err != nil {
		goto erraaa
	}
	for _, b := range users {
		if b.token == data.Token {
			p = fmt.Sprintf("<div class=\"message\"><div class=\"message-info\"><span class=\"sender\">%v</span></div><p>%v</p></div>", data.Username, data.Message)
			history += p
			return []byte(p), nil
		}
	}
	erraaa:
	p = fmt.Sprintf("<div class=\"message\"><div class=\"message-info\"><span class=\"sender\">%v</span></div><p>有人使用了错误的token或使用奇奇怪怪的方法发送消息，信息为：%v</p></div>", "系统", string(s))
	history += p
	return []byte(p), nil
}
