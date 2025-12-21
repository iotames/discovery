# settings.py
import os
from pathlib import Path

BOT_NAME = 'sitedown'

SPIDER_MODULES = ['sitedown.spiders']
NEWSPIDER_MODULE = 'sitedown.spiders'

# 遵守robots.txt规则（设置为False以爬取所有页面）
ROBOTSTXT_OBEY = False

# 配置并发请求
CONCURRENT_REQUESTS = 16
CONCURRENT_REQUESTS_PER_DOMAIN = 8
DOWNLOAD_DELAY = 0.5

# 启用和配置下载器中间件
DOWNLOADER_MIDDLEWARES = {
    'scrapy.downloadermiddlewares.httpproxy.HttpProxyMiddleware': 400,
    'sitedown.middlewares.DsiteMirrorDownloaderMiddleware': 543,
}
HTTP_PROXY = 'http://127.0.0.1:7890'
HTTPS_PROXY = 'http://127.0.0.1:7890'

# 启用和配置项目管道
ITEM_PIPELINES = {
    'sitedown.pipelines.CustomFilesPipeline': 1,      # 替换 scrapy.pipelines.files.FilesPipeline
    'sitedown.pipelines.CustomImagesPipeline': 2,     # 替换 scrapy.pipelines.images.ImagesPipeline
    'sitedown.pipelines.HtmlFilePipeline': 800,
}

# 自动限速扩展
AUTOTHROTTLE_ENABLED = True
AUTOTHROTTLE_START_DELAY = 1
AUTOTHROTTLE_MAX_DELAY = 60
AUTOTHROTTLE_TARGET_CONCURRENCY = 2.0

# 下载超时
DOWNLOAD_TIMEOUT = 30

# 重试设置
RETRY_ENABLED = True
RETRY_TIMES = 2

# 配置请求头
DEFAULT_REQUEST_HEADERS = {
    'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
    'Accept-Language': 'en',
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36',
}

# 深度限制
DEPTH_LIMIT = 0  # 0表示无限制

# 日志设置
LOG_LEVEL = 'INFO'

# 文件存储设置
IMAGES_STORE = 'downloads/images'
FILES_STORE = 'downloads/files'
FONTS_STORE = 'downloads/fonts'
FETCH_XHR_STORE = 'downloads/fetch_xhr'
CSS_STORE = 'downloads/css'
JS_STORE = 'downloads/js'
MEDIA_STORE = 'downloads/media'

# 创建下载目录
BASE_DOWNLOAD_DIR = 'downloads'
Path(BASE_DOWNLOAD_DIR).mkdir(exist_ok=True)
for folder in ['images', 'files', 'fonts', 'fetch_xhr', 'css', 'js', 'media']:
    Path(BASE_DOWNLOAD_DIR).joinpath(folder).mkdir(exist_ok=True)

# 媒体管道设置
IMAGES_URLS_FIELD = 'image_urls'
IMAGES_RESULT_FIELD = 'images'

# 启用HTTP缓存
HTTPCACHE_ENABLED = True
HTTPCACHE_DIR = 'httpcache'
HTTPCACHE_EXPIRATION_SECS = 0  # 永不过期
HTTPCACHE_IGNORE_HTTP_CODES = []
HTTPCACHE_STORAGE = 'scrapy.extensions.httpcache.FilesystemCacheStorage'

# 允许下载媒体文件（图片、CSS、JS等）的域外请求
MEDIA_ALLOW_REDIRECTS = True
# 重要：告诉Scrapy不要自动过滤掉来自item['file_urls']和item['image_urls']的域外请求
URLLENGTH_LIMIT = 2000  # 可选，防止超长URL被过滤