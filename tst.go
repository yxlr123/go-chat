package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
	router := gin.New()
	router.Use(cors.Default())
	router.MaxMultipartMemory = 1024 << 20
	router.LoadHTMLGlob("public/*")
	router.GET("/",func(ctx *gin.Context) {
		ctx.HTML(200,"1.html",gin.H{})
	})
	router.POST("/upload", func(c *gin.Context) {
		log.Println(c.Request.Header.Get("X-Token") == "")
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "请选择文件"})
			return
		}
		hash := md5.Sum([]byte(file.Filename))
		filename := hex.EncodeToString(hash[:]) + filepath.Ext(file.Filename)
		err = c.SaveUploadedFile(file, "src/"+filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "文件上传失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("文件上传成功，保存为 %s", filename)})
	})
	router.Run(":8888")
}