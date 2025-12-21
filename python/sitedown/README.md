## 介绍

项目已废弃，仅供研习。不可用。

## 运行爬虫


```bash
scrapy crawl dsite

scrapy crawl dsite -s HTTPS_PROXY='http://127.0.0.1:7890' -s HTTP_PROXY='http://127.0.0.1:7890' 
```

## AI Prompt

[角色]

你是Python爬虫工程师，擅长网站页面抓取，整站镜像克隆下载。

技能：
1. 精通Python语法。精通HTTP协议的请求和响应。
2. 熟悉文件操作，比如前端html，css, js文件处理。
3. 精通Scrapy爬虫框架。官方文档：https://docs.scrapy.org/en/latest/

[目标]

把 https://www.dsite.com/ 整站的 html,CSS,JS,Font,Img,Media 等资源全部下载到本地。做一个本地镜像站，所有资源可离线访问。
最终效果和以下命令行相识，但要做一些改动，我会在后续的步骤详细说明。

```
wget -m -k -p -e robots=off --header="User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36" https://example.com
```

[步骤]

1. 使用scrapy爬虫框架，在限定域名 www.dsite.com 下抓取整个站点的所有网页。
2. 【重要】html即Doc网页文档限定抓取域名，Fetch/XHR,CSS,JS,Font,Img,Media等依赖资源请求不限定域名。
3. 本次爬虫的所有下载文件都保存到本地的指定目录。
4. 网站Doc首页保存为根目录的index.html文件，其他Doc网页也均以html文件格式保存。
5. 【重要】依赖的Fetch/XHR,CSS,JS,Font,Img,Media等资源文件，必须新建目录存放，和Doc资源隔开。文件夹名全部转英文小写。Fetch/XHR则为fetch_xhr。
6. 【重要】依赖的Fetch/XHR资源请求，一律转换为静态资源。如保存为json文件格式。
7. 【重要】对于跨域名资源的请求，新建域名子目录存放资源。
8. 【重要】所有链接的请求路径，一律转换为本地相对路径，便以定位到实际已保存的资源文件。
9. 【重要】依赖资源文件，有一些要酌情进行转换。如图片资源请求路径，出现 `/cdn-cgi/image/w=1920,format=auto/files/` 路径开头的，一律替换成 `/files/`。实际请求路径和保存路径都要替换。http请求获取的返回内容，记得遵照规则保存在img目录。

[输出]

输出scrapy爬虫代码，使得 `scrapy crawl dsite` 命令行能完成目标。
