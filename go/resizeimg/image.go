package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// 解码图片数据
func decodeImage(data []byte) (img image.Image, format string, err error) {
	img, format, err = image.Decode(strings.NewReader(string(data)))
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

// 将图片转换为base64字符串
func imageToBase64(img image.Image, format string) (string, error) {
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		if err != nil {
			return "", fmt.Errorf("JPEG编码失败: %v", err)
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return "", fmt.Errorf("PNG编码失败: %v", err)
		}
	default:
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		if err != nil {
			return "", fmt.Errorf("默认编码失败: %v", err)
		}
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// 从URL获取文件扩展名。默认为 .jpg
func getExtensionFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ".jpg"
	}

	ext := filepath.Ext(parsedURL.Path)
	if ext == "" {
		return ".jpg"
	}
	return strings.ToLower(ext)
}

// 保存图片到缓存
func saveCacheImage(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// 加载本地图片
func loadLocalImage(filePath string) (image.Image, string, []byte, error) {
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			filePath = absPath
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, "", nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	if fileInfo, err := os.Stat(filePath); err == nil && fileInfo.IsDir() {
		return nil, "", nil, fmt.Errorf("路径是目录，不是文件: %s", filePath)
	}

	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("读取文件失败: %v", err)
	}

	if len(imageData) == 0 {
		return nil, "", nil, fmt.Errorf("文件内容为空: %s", filePath)
	}

	img, format, err := decodeImage(imageData)
	if err != nil {
		return nil, "", nil, fmt.Errorf("解码图片失败: %v", err)
	}

	return img, format, imageData, nil
}

func getResizeImg(img image.Image, format string, fileName string, width int) (resizedImg image.Image, thumbnailBase64, msg string, err error) {
	thumbnailPath := filepath.Join("./thumbnails", fileName)
	msg = "缩放成功"

	// 检查缩略图是否已存在
	if fileInfo, err := os.Stat(thumbnailPath); err == nil && fileInfo.Size() > 0 {
		// 读取已存在的缩略图转换为base64
		existingImg, existingFormat, existingData, err := loadLocalImage(thumbnailPath)
		if err == nil {
			thumbnailBase64 = base64.StdEncoding.EncodeToString(existingData)
			msg = "缩略图已存在，直接返回"
			fmt.Println(msg, thumbnailPath)
			resizedImg = existingImg
			format = existingFormat

		}
	}

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	if thumbnailBase64 == "" {
		// 等比缩放图片

		if originalWidth < width {
			width = originalWidth
		}
		fmt.Println("开始生成缩略图：", thumbnailPath)
		resizedImg = imaging.Resize(img, width, 0, imaging.Lanczos)

		// 转换为base64
		thumbnailBase64, err = imageToBase64(resizedImg, format)
		if err != nil {

			return
		}

		// 保存缩略图
		err = saveImage(thumbnailPath, resizedImg, format)
		if err != nil {
			fmt.Printf("保存缩略图失败: %v\n", err)
		}

		fmt.Printf("生成缩略图成功: %s (%dx%d -> %dx%d)\n", fileName, originalWidth, originalHeight, resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy())

	}

	return
}

// 保存图片
func saveImage(path string, img image.Image, format string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case "png":
		return png.Encode(file, img)
	default:
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	}
}

// 获取缩略图文件扩展名
func getThumbnailFilename(imgURI, format string, imageData []byte, width int) (hashStr, fileName string) {
	// 这边是缩略图的扩展名，和原图的扩展名不一样。
	thumbnailExt := getExtension(format, imgURI)
	// 计算原始图片的SHA256
	hash := sha256.Sum256(imageData)
	hashStr = hex.EncodeToString(hash[:])
	// 重要修改：在文件名中加入尺寸信息，确保不同尺寸有不同文件名
	// 格式：原图hash_宽度.扩展名
	fileName = fmt.Sprintf("%s_%d%s", hashStr, width, thumbnailExt)
	return hashStr, fileName
}

// 获取文件扩展名
func getExtension(format, uri string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	case "bmp":
		return ".bmp"
	case "webp":
		return ".webp"
	default:
		parsedURL, err := url.Parse(uri)
		if err == nil && parsedURL.Path != "" {
			ext := strings.ToLower(filepath.Ext(parsedURL.Path))
			if ext != "" {
				return ext
			}
		}

		ext := strings.ToLower(filepath.Ext(uri))
		if ext != "" {
			return ext
		}

		return ".jpg"
	}
}

// // 获取图片MIME类型
// func getMimeType(format string) string {
// 	switch strings.ToLower(format) {
// 	case "jpeg", "jpg":
// 		return "image/jpeg"
// 	case "png":
// 		return "image/png"
// 	case "gif":
// 		return "image/gif"
// 	case "bmp":
// 		return "image/bmp"
// 	case "webp":
// 		return "image/webp"
// 	default:
// 		return "image/jpeg"
// 	}
// }
