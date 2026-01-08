package main

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iotames/easyconf"
	"github.com/iotames/easyserver/httpsvr"
)

var WebPort int

func init() {
	cf := easyconf.NewConf()
	cf.IntVar(&WebPort, "WEB_PORT", 8081, "web服务端口")
	cf.Parse(true)
}

func main() {
	s := newHttpServer()
	// 静态文件目录
	s.AddMiddleHead(httpsvr.NewMiddleStatic("/thumbnails", "./thumbnails"))

	// 健康检查接口
	s.AddGetHandler("/health", func(ctx httpsvr.Context) {
		// easyserver.ResponseJsonOk(ctx, "ok")
		responseData(ctx, map[string]any{
			"time": time.Now().Format(time.RFC3339),
		})
	})

	// 图片缩放接口
	s.AddGetHandler("/resizeimg", handleResizeImage)

	// 确保必要目录存在
	dirs := []string{"./thumbnails", "./downimgs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录 %s 失败: %v\n", dir, err)
		}
	}

	fmt.Println("==============================================")
	fmt.Println("图片缩放服务启动成功!")
	fmt.Println("监听端口: ", WebPort)
	fmt.Println("==============================================")
	fmt.Println("使用示例:")
	fmt.Printf("1. 网络图片: http://127.0.0.1:%d/resizeimg?size=100&imguri=https://xxxxx.com/myname.jpg\n", WebPort)
	fmt.Printf("2. 本地图片: http://127.0.0.1:%d/resizeimg?size=100&imguri=./test.jpg\n", WebPort)
	fmt.Printf("3. 本地文件: http://127.0.0.1:%d/resizeimg?size=100&imguri=/Users/username/image.jpg\n", WebPort)
	fmt.Println("==============================================")
	fmt.Printf("健康检查: http://127.0.0.1:%d/health\n", WebPort)
	fmt.Println("==============================================")

	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}

func handleResizeImage(ctx httpsvr.Context) {
	// 原图文件名：使用URI的sha256
	// 裁剪后的缩略图文件名：使用原图文件的字节sha256

	imgURI := strings.TrimSpace(ctx.GetQueryValue("imguri", ""))
	sizeStr := strings.TrimSpace(ctx.GetQueryValue("size", ""))

	width, err := checkRequest(imgURI, sizeStr)
	if err != nil {
		responseFail(ctx, err, http.StatusBadRequest)
		return
	}

	if width > 2000 {
		width = 2000
	}

	fmt.Printf("处理请求: imguri=%s, size=%d\n", imgURI, width)

	img, format, imageData, err := getOriginalImage(imgURI)
	if err != nil {
		responseFail(ctx, err, http.StatusInternalServerError)
		return
	}

	hashStr, fileName := getThumbnailFilename(imgURI, format, imageData, width)

	resizedImg, thumbnailBase64, msg, err := getResizeImg(img, format, fileName, width)
	if err != nil {
		responseFail(ctx, fmt.Errorf("图片base64编码失败:%w", err), http.StatusInternalServerError)
	}

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()
	// 返回结果
	responseData(ctx, map[string]any{
		"message":  msg,
		"url":      fmt.Sprintf("/thumbnails/%s", fileName), // 缩略图的URL地址
		"filename": fileName,
		"sha256":   hashStr,
		"original_size": map[string]any{
			"width":  originalWidth,
			"height": originalHeight,
		},
		"resized_size": map[string]any{
			"width":  resizedImg.Bounds().Dx(),
			"height": resizedImg.Bounds().Dy(),
		},
		"requested_size": width,
		"actual_size":    resizedImg.Bounds().Dx(),
		"format":         strings.ToLower(format),
		"image_data":     thumbnailBase64,
	})

}

func getOriginalImage(imgURI string) (image.Image, string, []byte, error) {
	var img image.Image
	var format string
	var imageData []byte
	var err error

	// 判断URI类型
	if strings.HasPrefix(imgURI, "http") {
		// 网络图片，先检查缓存
		img, format, imageData, err = downloadImageWithCache(imgURI)
	} else {
		// 本地文件
		img, format, imageData, err = loadLocalImage(imgURI)
	}
	return img, format, imageData, err
}
