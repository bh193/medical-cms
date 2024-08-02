package main

import (
	"log"
	"medical-cms/database"
	"medical-cms/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
    // 初始化資料庫連接
    db, err :=database.InitDB("medical:Z73FX3LvPEQ7KKDh@tcp(ap03.m2x.com.tw:3306)/medical")
    if err != nil {
        log.Fatal("資料庫連線失敗:", err)
    }
    r := gin.Default()

    // 公開的路由
    r.GET("/auth/google/login", handlers.GoogleLogin)
    r.GET("/auth/callback", handlers.GoogleCallback)

    authorized := r.Group("/")
    // 1. token驗證
    authorized.Use(handlers.AuthMiddleware())

    // 2. 權限檢查
    authorized.Use(handlers.RequirePermission(db))
    {
        authorized.GET("/welcome", handlers.Welcome)
    }

    r.Run(":8080")
}