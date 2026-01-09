package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func getHttpClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}
	if Proxy != "" {
		// 设置代理URL
		proxyURL, err := url.Parse(Proxy)
		if err != nil {
			log.Fatal("Failed to parse proxy URL:", err)
		}

		// 自定义 http.Transport
		transport.Proxy = http.ProxyURL(proxyURL)
		log.Println("Using proxy:", proxyURL.String())
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
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

	resp, err := getHttpClient().Do(req)
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

// 下载网络图片（带缓存）
func downloadImageWithCache(urlStr string) (image.Image, string, []byte, error) {
	// 1. 计算URL的sha256作为缓存文件名
	hash := sha256.Sum256([]byte(urlStr))
	hashStr := hex.EncodeToString(hash[:])
	// 从URL获取文件扩展名。默认为 .jpg
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

	return nil, "", nil, fmt.Errorf("下载 %s 图片失败，尝试 %d 次后放弃: %v", url, maxRetries+1, lastErr)
}
