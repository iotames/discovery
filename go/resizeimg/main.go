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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 静态文件目录
	r.Static("/thumbnails", "./thumbnails")

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 图片缩放接口
	r.GET("/resizeimg", handleResizeImage)

	// 确保必要目录存在
	dirs := []string{"./thumbnails", "./downimgs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录 %s 失败: %v\n", dir, err)
		}
	}

	fmt.Println("==============================================")
	fmt.Println("图片缩放服务启动成功!")
	fmt.Println("监听端口: 8081")
	fmt.Println("==============================================")
	fmt.Println("使用示例:")
	fmt.Println("1. 网络图片: http://127.0.0.1:8081/resizeimg?size=100&imguri=https://xxxxx.com/myname.jpg")
	fmt.Println("2. 本地图片: http://127.0.0.1:8081/resizeimg?size=100&imguri=./test.jpg")
	fmt.Println("3. 本地文件: http://127.0.0.1:8081/resizeimg?size=100&imguri=/Users/username/image.jpg")
	fmt.Println("==============================================")
	fmt.Println("健康检查: http://127.0.0.1:8081/health")
	fmt.Println("==============================================")

	if err := r.Run(":8081"); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}

func handleResizeImage(c *gin.Context) {
	imgURI := strings.TrimSpace(c.Query("imguri"))
	sizeStr := strings.TrimSpace(c.Query("size"))

	if imgURI == "" || sizeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "参数缺失，请提供 imguri 和 size 参数",
			"example": "http://127.0.0.1:8081/resizeimg?imguri=https://example.com/image.jpg&size=100",
			"tip":     "本地文件可以使用 ./image.jpg 或 /path/to/image.jpg",
		})
		return
	}

	width, err := strconv.Atoi(sizeStr)
	if err != nil || width <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "size 参数必须是正整数",
		})
		return
	}

	if width > 2000 {
		width = 2000
	}

	fmt.Printf("处理请求: imguri=%s, size=%d\n", imgURI, width)

	var img image.Image
	var format string
	var imageData []byte

	// 判断URI类型
	if strings.HasPrefix(imgURI, "http://") || strings.HasPrefix(imgURI, "https://") {
		// 网络图片，先检查缓存
		img, format, imageData, err = downloadImageWithCache(imgURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "下载图片失败",
				"detail":     err.Error(),
				"imguri":     imgURI,
				"suggestion": "请检查URL是否正确，或者尝试使用本地文件",
				"local_tip":  "本地文件格式: ./image.jpg 或 /path/to/image.jpg",
				"example":    "http://127.0.0.1:8081/resizeimg?imguri=./test.jpg&size=100",
			})
			return
		}
	} else {
		// 本地文件
		img, format, imageData, err = loadLocalImage(imgURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "加载本地图片失败",
				"detail":     err.Error(),
				"imguri":     imgURI,
				"suggestion": "请检查文件路径是否正确，或者使用绝对路径",
			})
			return
		}
	}

	// 计算原始图片的SHA256
	hash := sha256.Sum256(imageData)
	hashStr := hex.EncodeToString(hash[:])
	ext := getExtension(format, imgURI)

	// 重要修改：在文件名中加入尺寸信息，确保不同尺寸有不同文件名
	// 格式：原图hash_宽度.扩展名
	fileName := fmt.Sprintf("%s_%d%s", hashStr, width, ext)
	thumbnailPath := filepath.Join("./thumbnails", fileName)

	// 检查缩略图是否已存在
	if fileInfo, err := os.Stat(thumbnailPath); err == nil && fileInfo.Size() > 0 {
		// 读取已存在的缩略图转换为base64
		existingImg, existingFormat, existingData, err := loadLocalImage(thumbnailPath)
		if err == nil {
			base64Str := base64.StdEncoding.EncodeToString(existingData)
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"url":      fmt.Sprintf("/thumbnails/%s", fileName),
				"message":  "缩略图已存在，直接返回",
				"filename": fileName,
				"sha256":   hashStr,
				"original_size": gin.H{
					"width":  img.Bounds().Dx(),
					"height": img.Bounds().Dy(),
				},
				"resized_size": gin.H{
					"width":  existingImg.Bounds().Dx(),
					"height": existingImg.Bounds().Dy(),
				},
				"requested_size": width,
				"actual_size":    existingImg.Bounds().Dx(),
				"format":         strings.ToLower(existingFormat),
				"image_data":     base64Str,
			})
			return
		}
	}

	// 等比缩放图片
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	if originalWidth < width {
		width = originalWidth
	}

	resizedImg := imaging.Resize(img, width, 0, imaging.Lanczos)

	// 转换为base64
	base64Str, err := imageToBase64(resizedImg, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "图片base64编码失败",
			"detail": err.Error(),
		})
		return
	}

	// 保存缩略图
	err = saveImage(thumbnailPath, resizedImg, format)
	if err != nil {
		fmt.Printf("保存缩略图失败: %v\n", err)
	}

	fmt.Printf("生成缩略图成功: %s (%dx%d -> %dx%d)\n",
		fileName, originalWidth, originalHeight,
		resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy())

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     fmt.Sprintf("/thumbnails/%s", fileName),
		"original_size": gin.H{
			"width":  originalWidth,
			"height": originalHeight,
		},
		"resized_size": gin.H{
			"width":  resizedImg.Bounds().Dx(),
			"height": resizedImg.Bounds().Dy(),
		},
		"requested_size": width,
		"actual_size":    resizedImg.Bounds().Dx(),
		"filename":       fileName,
		"sha256":         hashStr,
		"format":         strings.ToLower(format),
		"message":        "缩放成功",
		"image_data":     base64Str,
	})
}

// 从URL获取文件扩展名
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

// 下载网络图片（带缓存）
func downloadImageWithCache(urlStr string) (image.Image, string, []byte, error) {
	// 1. 计算URL的sha256作为缓存文件名
	hash := sha256.Sum256([]byte(urlStr))
	hashStr := hex.EncodeToString(hash[:])
	ext := getExtensionFromURL(urlStr)
	cacheFileName := hashStr + ext
	cachePath := filepath.Join("./downimgs", cacheFileName)

	// 2. 检查缓存文件是否存在
	if fileInfo, err := os.Stat(cachePath); err == nil && fileInfo.Size() > 0 {
		fmt.Printf("使用缓存文件: %s (URL: %s)\n", cacheFileName, urlStr)
		// 直接从缓存加载
		return loadLocalImage(cachePath)
	}

	// 3. 如果不存在，则下载
	fmt.Printf("下载图片: %s\n", urlStr)
	img, format, data, err := downloadImageWithRetry(urlStr, 2)
	if err != nil {
		return nil, "", nil, err
	}

	// 4. 下载成功后保存到缓存
	if err := saveCacheImage(cachePath, data); err != nil {
		fmt.Printf("保存缓存失败: %v\n", err)
	} else {
		fmt.Printf("已保存到缓存: %s\n", cacheFileName)
	}

	return img, format, data, nil
}

// 保存图片到缓存
func saveCacheImage(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// 下载网络图片（带重试机制）
func downloadImageWithRetry(url string, maxRetries int) (image.Image, string, []byte, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			fmt.Printf("第 %d 次重试下载: %s\n", i, url)
			time.Sleep(time.Duration(i) * time.Second)
		}

		img, format, data, err := downloadImage(url)
		if err == nil {
			return img, format, data, nil
		}
		lastErr = err
	}

	return nil, "", nil, fmt.Errorf("下载失败，尝试 %d 次后放弃: %v", maxRetries+1, lastErr)
}

// 下载网络图片
func downloadImage(url string) (image.Image, string, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置完整的请求头，模拟真实浏览器
	req.Header = http.Header{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Language":           {"zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"},
		"Cache-Control":             {"max-age=0"},
		"Priority":                  {"u=0, i"},
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"},
		"Sec-Ch-Ua":                 {`"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`},
		"Sec-Ch-Ua-Mobile":          {"?0"},
		"Sec-Ch-Ua-Platform":        {`"Windows"`},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"Connection":                {"keep-alive"},
		"Accept-Encoding":           {"gzip, deflate, br, zstd"},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if len(imageData) == 0 {
		return nil, "", nil, fmt.Errorf("图片数据为空")
	}

	img, format, err := decodeImage(imageData)
	if err != nil {
		return nil, "", nil, fmt.Errorf("解码图片失败: %v", err)
	}

	return img, format, imageData, nil
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

// 解码图片数据
func decodeImage(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(strings.NewReader(string(data)))
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
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

// 获取图片MIME类型
func getMimeType(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
