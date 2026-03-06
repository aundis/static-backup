package main

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type RestoreRequest struct {
	BackupFile string `json:"backupFile" binding:"required"`
	TargetDir  string `json:"targetDir" binding:"required"`
}

func restoreHandler(c *gin.Context) {
	var req RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 检查备份文件是否存在
	if _, err := os.Stat(req.BackupFile); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Backup file does not exist"})
		return
	}

	// 确保目标目录存在
	if err := os.MkdirAll(req.TargetDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create target directory: " + err.Error()})
		return
	}

	// 打开备份文件
	zipReader, err := zip.OpenReader(req.BackupFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open backup file: " + err.Error()})
		return
	}
	defer zipReader.Close()

	// 解压文件
	for _, file := range zipReader.File {
		if err := extractFile(file, req.TargetDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract file: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Restore completed successfully",
		"targetDir":  req.TargetDir,
		"backupFile": req.BackupFile,
	})
}

func extractFile(file *zip.File, targetDir string) error {
	// 创建目标文件路径
	targetPath := filepath.Join(targetDir, file.Name)

	// 确保目录存在
	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, file.Mode())
	}

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// 打开zip文件
	zipFile, err := file.Open()
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建目标文件
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// 复制文件内容
	if _, err := io.Copy(targetFile, zipFile); err != nil {
		return err
	}

	// 设置文件权限
	return os.Chmod(targetPath, file.Mode())
}
