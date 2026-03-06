package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

type ScheduleRequest struct {
	SourceDir   string `json:"sourceDir" binding:"required"`
	OutputDir   string `json:"outputDir" binding:"required"`
	CronExpr    string `json:"cronExpr" binding:"required"` // cron表达式
	TaskName    string `json:"taskName" binding:"required"` // 任务名称
	KeepCopies  int    `json:"keepCopies"`                  // 保留备份份数
	OSSConfigID string `json:"ossConfigId"`                 // OSS配置ID
}

type Schedule struct {
	ID          string    `json:"id"`
	SourceDir   string    `json:"sourceDir"`
	OutputDir   string    `json:"outputDir"`
	CronExpr    string    `json:"cronExpr"`
	TaskName    string    `json:"taskName"`    // 任务名称
	KeepCopies  int       `json:"keepCopies"`  // 保留备份份数
	OSSConfigID string    `json:"ossConfigId"` // OSS配置ID
	FileName    string    `json:"fileName"`    // 自动生成的文件名
	CreatedAt   time.Time `json:"createdAt"`
}

type BackupRecord struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	SourceDir  string    `json:"sourceDir"`
	OutputDir  string    `json:"outputDir"`
	ScheduleId string    `json:"scheduleId"`
	CreatedAt  time.Time `json:"createdAt"`
}

const (
	schedulesFile  = "config/schedules.json"
	ossConfigsFile = "config/oss_configs.json"
)

// OSS配置结构体
type OSSConfig struct {
	ID              string    `json:"id"`
	Name            string    `json:"name" binding:"required"`
	Endpoint        string    `json:"endpoint" binding:"required"`
	AccessKeyID     string    `json:"accessKeyId" binding:"required"`
	AccessKeySecret string    `json:"accessKeySecret" binding:"required"`
	BucketName      string    `json:"bucketName" binding:"required"`
	Prefix          string    `json:"prefix"`
	CreatedAt       time.Time `json:"createdAt"`
}

// OSS配置请求结构体
type OSSConfigRequest struct {
	Name            string `json:"name" binding:"required"`
	Endpoint        string `json:"endpoint" binding:"required"`
	AccessKeyID     string `json:"accessKeyId" binding:"required"`
	AccessKeySecret string `json:"accessKeySecret" binding:"required"`
	BucketName      string `json:"bucketName" binding:"required"`
	Prefix          string `json:"prefix"`
}

var (
	schedules      = make(map[string]Schedule)
	scheduleMutex  sync.Mutex
	cronScheduler  = cron.New()
	backupRecords  = make(map[string]BackupRecord)
	backupMutex    sync.Mutex
	ossConfigs     = make(map[string]OSSConfig)
	ossConfigMutex sync.Mutex
)

func init() {
	// 加载计划任务
	loadSchedules()
	// 加载OSS配置
	loadOSSConfigs()
	// 启动cron调度器
	cronScheduler.Start()
}

// 保存OSS配置到文件
func saveOSSConfigs() error {
	ossConfigMutex.Lock()
	defer ossConfigMutex.Unlock()

	// 转换为切片
	configList := make([]OSSConfig, 0, len(ossConfigs))
	for _, config := range ossConfigs {
		configList = append(configList, config)
	}

	// 转换为JSON
	data, err := json.MarshalIndent(configList, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return ioutil.WriteFile(ossConfigsFile, data, 0644)
}

// 从文件加载OSS配置
func loadOSSConfigs() error {
	// 检查文件是否存在
	if _, err := os.Stat(ossConfigsFile); os.IsNotExist(err) {
		// 文件不存在，返回
		return nil
	}

	// 读取文件
	data, err := ioutil.ReadFile(ossConfigsFile)
	if err != nil {
		return err
	}

	// 解析JSON
	var configList []OSSConfig
	if err := json.Unmarshal(data, &configList); err != nil {
		return err
	}

	// 加载到内存
	ossConfigMutex.Lock()
	defer ossConfigMutex.Unlock()

	for _, config := range configList {
		ossConfigs[config.ID] = config
	}

	return nil
}

// 获取所有OSS配置
func getOSSConfigsHandler(c *gin.Context) {
	ossConfigMutex.Lock()
	defer ossConfigMutex.Unlock()

	// 转换为切片
	configList := make([]OSSConfig, 0, len(ossConfigs))
	for _, config := range ossConfigs {
		configList = append(configList, config)
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configList,
	})
}

// 创建OSS配置
func createOSSConfigHandler(c *gin.Context) {
	var req OSSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 生成唯一ID
	configID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 创建OSS配置
	config := OSSConfig{
		ID:              configID,
		Name:            req.Name,
		Endpoint:        req.Endpoint,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		BucketName:      req.BucketName,
		Prefix:          req.Prefix,
		CreatedAt:       time.Now(),
	}

	// 保存到内存
	ossConfigMutex.Lock()
	ossConfigs[configID] = config
	ossConfigMutex.Unlock()

	// 保存到文件
	if err := saveOSSConfigs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OSS config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OSS config created successfully",
		"config":  config,
	})
}

// 更新OSS配置
func updateOSSConfigHandler(c *gin.Context) {
	configID := c.Param("id")

	// 检查OSS配置是否存在
	ossConfigMutex.Lock()
	existingConfig, exists := ossConfigs[configID]
	ossConfigMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "OSS config not found"})
		return
	}

	var req OSSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 更新OSS配置
	updatedConfig := OSSConfig{
		ID:              configID,
		Name:            req.Name,
		Endpoint:        req.Endpoint,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		BucketName:      req.BucketName,
		Prefix:          req.Prefix,
		CreatedAt:       existingConfig.CreatedAt,
	}

	// 保存到内存
	ossConfigMutex.Lock()
	ossConfigs[configID] = updatedConfig
	ossConfigMutex.Unlock()

	// 保存到文件
	if err := saveOSSConfigs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OSS config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OSS config updated successfully",
		"config":  updatedConfig,
	})
}

// 删除OSS配置
func deleteOSSConfigHandler(c *gin.Context) {
	configID := c.Param("id")

	// 检查OSS配置是否存在
	ossConfigMutex.Lock()
	_, exists := ossConfigs[configID]
	if !exists {
		ossConfigMutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "OSS config not found"})
		return
	}

	// 从内存中删除
	delete(ossConfigs, configID)
	ossConfigMutex.Unlock()

	// 保存到文件
	if err := saveOSSConfigs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OSS config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OSS config deleted successfully",
	})
}

// 登录验证
func loginHandler(c *gin.Context) {
	type LoginRequest struct {
		Password string `json:"password" binding:"required"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 这里可以从配置文件或环境变量中获取密码
	// 为了简单起见，我们使用硬编码的密码
	const correctPassword = "admin123"

	if req.Password == correctPassword {
		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid password",
		})
	}
}

// 测试OSS配置连通性
func testOSSConfigHandler(c *gin.Context) {
	var req OSSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 创建OSS客户端
	client, err := oss.New(req.Endpoint, req.AccessKeyID, req.AccessKeySecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create OSS client: " + err.Error()})
		return
	}

	// 测试桶是否存在
	exists, err := client.IsBucketExist(req.BucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence: " + err.Error()})
		return
	}

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bucket does not exist"})
		return
	}

	// 尝试获取存储空间，验证权限
	_, err = client.Bucket(req.BucketName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get bucket: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OSS connection test successful",
		"config":  req,
	})
}

// 保存计划任务到文件
func saveSchedules() error {
	scheduleMutex.Lock()
	defer scheduleMutex.Unlock()

	// 转换为切片
	scheduleList := make([]Schedule, 0, len(schedules))
	for _, schedule := range schedules {
		scheduleList = append(scheduleList, schedule)
	}

	// 转换为JSON
	data, err := json.MarshalIndent(scheduleList, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return ioutil.WriteFile(schedulesFile, data, 0644)
}

// 从文件加载计划任务
func loadSchedules() error {
	// 检查文件是否存在
	if _, err := os.Stat(schedulesFile); os.IsNotExist(err) {
		// 文件不存在，返回
		return nil
	}

	// 读取文件
	data, err := ioutil.ReadFile(schedulesFile)
	if err != nil {
		return err
	}

	// 解析JSON
	var scheduleList []Schedule
	if err := json.Unmarshal(data, &scheduleList); err != nil {
		return err
	}

	// 加载到内存
	scheduleMutex.Lock()
	defer scheduleMutex.Unlock()

	for _, schedule := range scheduleList {
		schedules[schedule.ID] = schedule
		// 添加到cron
		cronScheduler.AddFunc(schedule.CronExpr, func() {
			backupHandlerForSchedule(schedule)
		})
	}

	return nil
}

func createScheduleHandler(c *gin.Context) {
	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 生成唯一ID
	scheduleID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 创建计划任务
	schedule := Schedule{
		ID:          scheduleID,
		SourceDir:   req.SourceDir,
		OutputDir:   req.OutputDir,
		CronExpr:    req.CronExpr,
		TaskName:    req.TaskName,
		KeepCopies:  req.KeepCopies,
		OSSConfigID: req.OSSConfigID,
		FileName:    "", // 文件名将在执行备份时自动生成
		CreatedAt:   time.Now(),
	}

	// 添加到cron
	err := cronScheduler.AddFunc(req.CronExpr, func() {
		// 执行备份
		backupHandlerForSchedule(schedule)
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression: " + err.Error()})
		return
	}

	// 保存到内存
	scheduleMutex.Lock()
	schedules[scheduleID] = schedule
	scheduleMutex.Unlock()

	// 保存到文件
	if err := saveSchedules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Schedule created successfully",
		"schedule": schedule,
	})
}

func getSchedulesHandler(c *gin.Context) {
	scheduleMutex.Lock()
	defer scheduleMutex.Unlock()

	// 转换为切片
	scheduleList := make([]Schedule, 0, len(schedules))
	for _, schedule := range schedules {
		scheduleList = append(scheduleList, schedule)
	}

	c.JSON(http.StatusOK, gin.H{
		"schedules": scheduleList,
	})
}

func deleteScheduleHandler(c *gin.Context) {
	scheduleID := c.Param("id")

	// 检查计划任务是否存在
	scheduleMutex.Lock()
	_, exists := schedules[scheduleID]
	scheduleMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	// 重新创建cron调度器（简单方法，实际项目中可以更优化）
	cronScheduler.Stop()
	cronScheduler = cron.New()
	cronScheduler.Start()

	// 重新添加其他计划任务
	scheduleMutex.Lock()
	for id, schedule := range schedules {
		if id != scheduleID {
			cronScheduler.AddFunc(schedule.CronExpr, func() {
				backupHandlerForSchedule(schedule)
			})
		}
	}
	// 从内存中删除
	scheduleMutex.Lock()
	delete(schedules, scheduleID)
	scheduleMutex.Unlock()

	// 保存到文件
	if err := saveSchedules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schedule deleted successfully",
	})
}

func backupHandlerForSchedule(schedule Schedule) {
	// 生成备份文件名，包含计划任务ID
	fileName := fmt.Sprintf("backup_%s_%s.zip", schedule.ID, time.Now().Format("20060102_150405"))

	// 执行备份
	backupPath, err := performBackup(schedule.SourceDir, schedule.OutputDir, fileName)
	if err != nil {
		fmt.Printf("Backup failed for schedule %s: %v\n", schedule.ID, err)
		return
	}

	// 根据保留份数清理旧的备份文件
	if schedule.KeepCopies > 0 {
		cleanupOldBackups(schedule.OutputDir, schedule.ID, schedule.KeepCopies)
	}

	// 如果关联了OSS配置，上传到OSS
	if schedule.OSSConfigID != "" {
		ossConfigMutex.Lock()
		ossConfig, exists := ossConfigs[schedule.OSSConfigID]
		ossConfigMutex.Unlock()

		if exists {
			if err := uploadToOSS(backupPath, fileName, ossConfig); err != nil {
				fmt.Printf("Failed to upload backup to OSS for schedule %s: %v\n", schedule.ID, err)
			} else {
				fmt.Printf("Backup uploaded to OSS successfully for schedule %s\n", schedule.ID)
			}
		}
	}

	fmt.Printf("Backup completed successfully for schedule %s: %s\n", schedule.ID, backupPath)
}

// 上传文件到OSS
func uploadToOSS(localPath, fileName string, config OSSConfig) error {
	// 创建OSS客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("failed to create OSS client: %v", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to get bucket: %v", err)
	}

	// 构建OSS对象路径
	ossPath := fileName
	if config.Prefix != "" {
		ossPath = config.Prefix + "/" + fileName
	}

	// 上传文件
	if err := bucket.PutObjectFromFile(ossPath, localPath); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// 从文件名中提取计划任务ID
func extractScheduleIDFromFileName(fileName string) string {
	// 文件名格式为: backup_[scheduleId]_[timestamp].zip
	if len(fileName) > 7 && fileName[:7] == "backup_" {
		// 查找第一个下划线后的部分
		parts := strings.Split(fileName[7:], "_")
		if len(parts) > 1 {
			return parts[0]
		}
	}
	return ""
}

// 从OSS删除文件
func deleteFromOSS(fileName string, config OSSConfig) error {
	// 创建OSS客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("failed to create OSS client: %v", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to get bucket: %v", err)
	}

	// 构建OSS对象路径
	ossPath := fileName
	if config.Prefix != "" {
		ossPath = config.Prefix + "/" + fileName
	}

	// 删除文件
	if err := bucket.DeleteObject(ossPath); err != nil {
		return fmt.Errorf("failed to delete file from OSS: %v", err)
	}

	return nil
}

// 清理旧的备份文件
func cleanupOldBackups(outputDir, scheduleID string, keepCopies int) {
	// 读取目录内容
	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		fmt.Printf("Failed to read backup directory: %v\n", err)
		return
	}

	// 筛选出当前计划任务的备份文件
	var backupFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), fmt.Sprintf("backup_%s_", scheduleID)) {
			backupFiles = append(backupFiles, file)
		}
	}

	// 如果文件数量超过保留份数，删除最旧的文件
	if len(backupFiles) > keepCopies {
		// 按修改时间排序，旧的在前
		sort.Slice(backupFiles, func(i, j int) bool {
			return backupFiles[i].ModTime().Before(backupFiles[j].ModTime())
		})

		// 检查计划任务是否关联了OSS配置
		scheduleMutex.Lock()
		schedule, exists := schedules[scheduleID]
		scheduleMutex.Unlock()

		var ossConfig OSSConfig
		ossExists := false
		if exists && schedule.OSSConfigID != "" {
			ossConfigMutex.Lock()
			ossConfig, ossExists = ossConfigs[schedule.OSSConfigID]
			ossConfigMutex.Unlock()
		}

		// 删除多余的文件
		for i := 0; i < len(backupFiles)-keepCopies; i++ {
			fileName := backupFiles[i].Name()
			filePath := filepath.Join(outputDir, fileName)

			// 如果关联了OSS配置，从OSS删除文件
			if ossExists {
				if err := deleteFromOSS(fileName, ossConfig); err != nil {
					// 记录错误但不影响本地删除
					fmt.Printf("Failed to delete old backup from OSS: %v\n", err)
				}
			}

			// 删除本地文件
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Failed to delete old backup file %s: %v\n", filePath, err)
			}
		}
	}
}

func triggerScheduleHandler(c *gin.Context) {
	scheduleID := c.Param("id")

	// 检查计划任务是否存在
	scheduleMutex.Lock()
	schedule, exists := schedules[scheduleID]
	scheduleMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	// 手动触发备份
	go backupHandlerForSchedule(schedule)

	c.JSON(http.StatusOK, gin.H{
		"message": "Schedule triggered successfully",
	})
}

func updateScheduleHandler(c *gin.Context) {
	scheduleID := c.Param("id")

	// 检查计划任务是否存在
	scheduleMutex.Lock()
	existingSchedule, exists := schedules[scheduleID]
	scheduleMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 更新计划任务
	updatedSchedule := Schedule{
		ID:          scheduleID,
		SourceDir:   req.SourceDir,
		OutputDir:   req.OutputDir,
		CronExpr:    req.CronExpr,
		TaskName:    req.TaskName,
		KeepCopies:  req.KeepCopies,
		OSSConfigID: req.OSSConfigID,
		FileName:    existingSchedule.FileName, // 保持原有的文件名
		CreatedAt:   existingSchedule.CreatedAt,
	}

	// 重新创建cron调度器
	cronScheduler.Stop()
	cronScheduler = cron.New()
	cronScheduler.Start()

	// 重新添加所有计划任务
	scheduleMutex.Lock()
	for id, schedule := range schedules {
		if id == scheduleID {
			// 添加更新后的计划任务
			err := cronScheduler.AddFunc(updatedSchedule.CronExpr, func() {
				backupHandlerForSchedule(updatedSchedule)
			})
			if err != nil {
				scheduleMutex.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression: " + err.Error()})
				return
			}
			schedules[id] = updatedSchedule
		} else {
			// 添加其他计划任务
			cronScheduler.AddFunc(schedule.CronExpr, func() {
				backupHandlerForSchedule(schedule)
			})
		}
	}
	scheduleMutex.Unlock()

	// 保存到文件
	if err := saveSchedules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Schedule updated successfully",
		"schedule": updatedSchedule,
	})
}

func getBackupRecordsHandler(c *gin.Context) {
	// 获取查询参数
	scheduleId := c.Query("scheduleId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 扫描 backup 目录
	backupDir := "./backup"
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backup directory: " + err.Error()})
		return
	}

	// 转换为切片
	recordList := make([]BackupRecord, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 从文件名中提取信息
		// 文件名格式为: backup_[scheduleId]_[timestamp].zip
		fileName := file.Name()

		// 提取计划任务ID
		currentScheduleId := ""
		// 简单的解析逻辑
		if len(fileName) > 7 && fileName[:7] == "backup_" {
			// 查找第一个下划线后的部分
			parts := strings.Split(fileName[7:], "_")
			if len(parts) > 1 {
				currentScheduleId = parts[0]
			}
		}

		// 根据 scheduleId 筛选
		if scheduleId != "" && currentScheduleId != scheduleId {
			continue
		}

		// 从计划任务中获取源目录
		sourceDir := ""
		scheduleMutex.Lock()
		if schedule, exists := schedules[currentScheduleId]; exists {
			sourceDir = schedule.SourceDir
		}
		scheduleMutex.Unlock()

		// 创建备份记录
		record := BackupRecord{
			ID:         fmt.Sprintf("%d", file.ModTime().UnixNano()),
			FileName:   fileName,
			SourceDir:  sourceDir,
			OutputDir:  backupDir,
			ScheduleId: currentScheduleId,
			CreatedAt:  file.ModTime(),
		}

		recordList = append(recordList, record)
	}

	// 按创建时间排序，最新的在前
	sort.Slice(recordList, func(i, j int) bool {
		return recordList[i].CreatedAt.After(recordList[j].CreatedAt)
	})

	// 计算总记录数
	total := len(recordList)

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		recordList = []BackupRecord{}
	} else if end > total {
		recordList = recordList[start:]
	} else {
		recordList = recordList[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"records":    recordList,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// 删除备份文件
func deleteBackupHandler(c *gin.Context) {
	fileName := c.Param("fileName")
	backupDir := "./backup"
	filePath := filepath.Join(backupDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup file not found"})
		return
	}

	// 从文件名中提取计划任务ID
	scheduleID := extractScheduleIDFromFileName(fileName)

	// 如果有计划任务ID，检查是否关联了OSS配置
	if scheduleID != "" {
		scheduleMutex.Lock()
		schedule, exists := schedules[scheduleID]
		scheduleMutex.Unlock()

		// 如果关联了OSS配置，从OSS删除文件
		if exists && schedule.OSSConfigID != "" {
			ossConfigMutex.Lock()
			ossConfig, ossExists := ossConfigs[schedule.OSSConfigID]
			ossConfigMutex.Unlock()

			if ossExists {
				if err := deleteFromOSS(fileName, ossConfig); err != nil {
					// 记录错误但不影响本地删除
					fmt.Printf("Failed to delete backup from OSS: %v\n", err)
				}
			}
		}
	}

	// 删除本地文件
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete backup file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup file deleted successfully",
	})
}

// 下载备份文件
func downloadBackupHandler(c *gin.Context) {
	fileName := c.Param("fileName")
	backupDir := "./backup"
	filePath := filepath.Join(backupDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup file not found"})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/zip")

	// 发送文件
	c.File(filePath)
}
