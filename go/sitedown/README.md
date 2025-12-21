# Sitedown

Sitedown 是一个用 Go 语言编写的网站镜像工具，可以将整个网站下载到本地，创建一个可离线浏览的静态版本。

## 功能特点

- 递归下载整个网站
- 按资源类型分类存储（HTML、CSS、JS、图片等）
- 支持跨域资源下载
- CDN路径转换
- URL重写为相对路径以便离线浏览
- 设置网络代理加速下载

## 目录结构

```
sitedown/
├── main.go                 # 主程序
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 依赖校验和
├── downloads/              # 下载文件存储目录
│   ├── css/                # CSS 文件
│   ├── js/                 # JavaScript 文件
│   ├── images/             # 图片文件
│   ├── fonts/              # 字体文件
│   ├── media/              # 音视频文件
│   ├── fetch_xhr/          # API 请求数据
│   └── www_dsite_com/      # HTML 页面
└── README.md               # 项目说明
```

## 工作原理

1. 程序从指定的基础URL开始下载页面
2. 解析HTML文档，提取所有链接
3. 根据资源类型将文件保存到不同的目录中
4. 重写HTML中的URL为相对路径
5. 递归处理提取到的所有链接
6. 对于CDN图片路径进行特殊处理

## 注意事项

- 程序只会爬取指定域名下的HTML页面
- 其他类型的资源（CSS、JS、图片等）不受域名限制
- 所有下载的文件都会保存在 [downloads](file:///c%3A/projects/devops/discovery/go/sitedown/downloads) 目录中
- CDN图片路径会被转换为本地 [/files/](file:///c%3A/projects/devops/discovery/go/sitedown/downloads/files/) 路径
- API请求会被保存为JSON文件