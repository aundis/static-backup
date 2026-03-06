package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type BackupRequest struct {
	SourceDir string `json:"sourceDir" binding:"required"`
	OutputDir string `json:"outputDir" binding:"required"`
	FileName  string `json:"fileName"`
}

// 执行备份操作
func performBackup(sourceDir, outputDir, fileName string) (string, error) {
	// 检查源目录是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return "", fmt.Errorf("source directory does not exist")
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// 生成备份文件名
	if fileName == "" {
		fileName = fmt.Sprintf("backup_%s.zip", time.Now().Format("20060102_150405"))
	}

	// 完整的备份文件路径
	backupPath := filepath.Join(outputDir, fileName)

	// 创建备份文件
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer zipFile.Close()

	// 创建zip写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历目录并添加到zip
	if err := addFilesToZip(zipWriter, sourceDir, ""); err != nil {
		return "", fmt.Errorf("failed to add files to backup: %v", err)
	}

	return backupPath, nil
}

func backupHandler(c *gin.Context) {
	var req BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 执行备份
	backupPath, err := performBackup(req.SourceDir, req.OutputDir, req.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Backup completed successfully",
		"backupPath": backupPath,
		"fileName":   filepath.Base(backupPath),
	})
}

func addFilesToZip(zipWriter *zip.Writer, basePath, relativePath string) error {
	// 读取目录内容
	files, err := os.ReadDir(filepath.Join(basePath, relativePath))
	if err != nil {
		return err
	}

	for _, file := range files {
		fullPath := filepath.Join(basePath, relativePath, file.Name())
		zipPath := filepath.Join(relativePath, file.Name())

		if file.IsDir() {
			// 递归处理子目录
			if err := addFilesToZip(zipWriter, basePath, zipPath); err != nil {
				return err
			}
		} else {
			// 打开文件
			srcFile, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			// 创建zip文件条目
			zipFile, err := zipWriter.Create(zipPath)
			if err != nil {
				return err
			}

			// 复制文件内容
			if _, err := io.Copy(zipFile, srcFile); err != nil {
				return err
			}
		}
	}

	return nil
}
