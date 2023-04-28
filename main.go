package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
		_ = tokenBool(p)

		// 将消息广播给所有连接的客户端
		for c := range clients {
			err := c.WriteMessage(messageType, p)
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
	tpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(c.Writer, nil)
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
	router.GET("/", serveHome)

	router.GET("/login", login)

	err := router.Run(":8888")
	if err != nil {
		log.Fatal(err)
	}
}

func tokenBool(s []byte) []byte {
	var data *userMessage
	json.Unmarshal(s,&data)
	log.Println( "nnnn", data)
	return []byte{}
}
