package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Directory struct {
	Path     string      `json:"path"`
	Name     string      `json:"name"`
	IsDir    bool        `json:"isDir"`
	Children []Directory `json:"children,omitempty"`
}

func getDirectoriesHandler(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		// 在Windows上，获取所有驱动器
		drives, err := getWindowsDrives()
		if err != nil {
			// 如果获取驱动器失败（可能是在非Windows系统上），返回根目录
			dirs, err := getDirectoryContents("/")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"directories": dirs})
			return
		}
		
		// 如果获取到了驱动器（Windows系统），返回驱动器列表
		if len(drives) > 0 {
			c.JSON(http.StatusOK, gin.H{"directories": drives})
			return
		}
		
		// 如果没有获取到驱动器，返回根目录
		dirs, err := getDirectoryContents("/")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"directories": dirs})
		return
	}

	// 遍历指定路径
	dirs, err := getDirectoryContents(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"directories": dirs})
}

func getWindowsDrives() ([]Directory, error) {
	var drives []Directory
	// 检查从A到Z的驱动器
	for i := 'A'; i <= 'Z'; i++ {
		drive := string(i) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, Directory{
				Path:  drive,
				Name:  drive,
				IsDir: true,
			})
		}
	}
	return drives, nil
}

func getDirectoryContents(path string) ([]Directory, error) {
	var contents []Directory

	// 读取目录内容
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePath := filepath.Join(path, file.Name())
		contents = append(contents, Directory{
			Path:  filePath,
			Name:  file.Name(),
			IsDir: file.IsDir(),
		})
	}

	return contents, nil
}
