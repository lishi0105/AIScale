package chunkImport

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"hdzk.cn/foodapp/internal/server/middleware"
)

func UploadChunk(c *gin.Context, uploadDir string) (int, error) {
	const errTitle = "上传文件切片失败"
	if uploadDir == "" {
		return 0, fmt.Errorf("上传路径不能为空")
	}
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	// 获取参数
	filename := c.PostForm("filename")
	chunkIndexStr := c.PostForm("chunk_index")

	if filename == "" {
		return 0, fmt.Errorf("缺少文件名")
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		return 0, fmt.Errorf("切片索引格式错误")
	}

	// 获取文件数据
	file, err := c.FormFile("file")
	if err != nil {
		return 0, fmt.Errorf("获取文件失败")
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return 0, fmt.Errorf("打开文件失败" + err.Error())
	}
	defer src.Close()

	// 读取文件数据
	chunkData := make([]byte, file.Size)
	if _, err := src.Read(chunkData); err != nil {
		return 0, fmt.Errorf("读取文件数据失败" + err.Error())
	}

	// 保存切片
	if err := SaveUploadedChunk(uploadDir, filename, chunkIndex, chunkData); err != nil {
		return 0, fmt.Errorf("保存切片失败: " + err.Error())
	}
	return chunkIndex, nil
}

func MergeChunks(c *gin.Context, uploadDir string) (string, error) {
	const errTitle = "合并文件切片失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		return "", fmt.Errorf("账户已删除，禁止操作")
	}
	if act.Role != middleware.RoleAdmin {
		return "", fmt.Errorf("仅管理员可操作")
	}

	var req struct {
		Filename    string `json:"filename" binding:"required"`
		TotalChunks int    `json:"total_chunks" binding:"required,min=1"`
		MD5         string `json:"md5" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return "", fmt.Errorf("输入格式非法: " + err.Error())
	}

	// 合并切片
	finalPath, err := MergeChunkFiles(uploadDir, req.Filename, req.TotalChunks)
	if err != nil {
		return "", fmt.Errorf("合并切片失败: " + err.Error())
	}

	// 校验MD5
	if err := ValidateFile(finalPath, req.MD5); err != nil {
		// 删除文件
		os.Remove(finalPath)
		return "", fmt.Errorf("校验 MD5 失败: " + err.Error())
	}
	return finalPath, nil
}

// SaveUploadedChunk 保存上传的文件切片
func SaveUploadedChunk(uploadDir, filename string, chunkIndex int, chunkData []byte) error {
	// 创建临时目录
	tmpDir := filepath.Join(uploadDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 保存切片
	chunkPath := filepath.Join(tmpDir, fmt.Sprintf("%s.part%d", filename, chunkIndex))
	if err := os.WriteFile(chunkPath, chunkData, 0644); err != nil {
		return fmt.Errorf("保存切片失败: %w", err)
	}

	return nil
}

// MergeChunks 合并文件切片
func MergeChunkFiles(uploadDir, filename string, totalChunks int) (string, error) {
	tmpDir := filepath.Join(uploadDir, "tmp")
	finalPath := filepath.Join(uploadDir, filename)

	// 创建最终文件
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("创建最终文件失败: %w", err)
	}
	defer finalFile.Close()

	// 合并所有切片
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("%s.part%d", filename, i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return "", fmt.Errorf("读取切片 %d 失败: %w", i, err)
		}

		if _, err := finalFile.Write(chunkData); err != nil {
			return "", fmt.Errorf("写入数据失败: %w", err)
		}

		// 删除切片文件
		os.Remove(chunkPath)
	}

	return finalPath, nil
}

func ValidateFile(filePath string, expectedMD5 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("计算MD5失败: %w", err)
	}

	actualMD5 := hex.EncodeToString(hash.Sum(nil))
	if actualMD5 != expectedMD5 {
		return fmt.Errorf("文件MD5校验失败: 期望 %s, 实际 %s", expectedMD5, actualMD5)
	}

	return nil
}
