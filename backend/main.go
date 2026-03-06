package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 验证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		token := c.GetHeader("Authorization")

		// 简单验证，实际项目中应该使用更安全的方式
		if token != "Bearer admin123" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	// 初始化Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 注册API路由
	registerRoutes(r)

	// 提供前端静态文件
	r.Static("/", "./frontend")

	// 启动服务器
	serverAddr := ":8080"
	log.Printf("Server starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func registerRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 登录相关路由
	r.POST("/api/login", loginHandler)

	// 需要验证的路由组
	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		// 目录相关路由
		api.GET("/directories", getDirectoriesHandler)

		// 备份相关路由
		api.POST("/backup", backupHandler)
		api.POST("/restore", restoreHandler)

		// 计划任务相关路由
		api.GET("/schedules", getSchedulesHandler)
		api.POST("/schedule", createScheduleHandler)
		api.POST("/schedule/:id/trigger", triggerScheduleHandler)
		api.DELETE("/schedule/:id", deleteScheduleHandler)
		api.PUT("/schedule/:id", updateScheduleHandler)

		// 备份记录相关路由
		api.GET("/backup-records", getBackupRecordsHandler)
		api.DELETE("/backup/:fileName", deleteBackupHandler)
		api.GET("/backup/:fileName", downloadBackupHandler)

		// OSS配置相关路由
		api.GET("/oss-configs", getOSSConfigsHandler)
		api.POST("/oss-config", createOSSConfigHandler)
		api.PUT("/oss-config/:id", updateOSSConfigHandler)
		api.DELETE("/oss-config/:id", deleteOSSConfigHandler)
		api.POST("/oss-config/test", testOSSConfigHandler)
	}
}
