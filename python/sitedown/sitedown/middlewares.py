# middlewares.py
import logging
from urllib.parse import urlparse
from scrapy import signals
from scrapy.http import HtmlResponse

logger = logging.getLogger(__name__)

class DsiteMirrorDownloaderMiddleware:
    
    @classmethod
    def from_crawler(cls, crawler):
        s = cls()
        crawler.signals.connect(s.spider_opened, signal=signals.spider_opened)
        return s

    def process_request(self, request, spider):
        # 处理特殊路径替换（cdn-cgi/image/... -> files/...）
        url = request.url
        if '/cdn-cgi/image/' in url:
            new_url = url.replace('/cdn-cgi/image/w=1920,format=auto/files/', '/files/')
            new_url = url.replace('/cdn-cgi/image/', '/files/')  # 更通用的替换
            logger.debug(f"Replacing URL: {url} -> {new_url}")
            request = request.replace(url=new_url)
        
        # 设置请求头
        request.headers.setdefault('User-Agent', 
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36')
        request.headers.setdefault('Accept', 
            'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8')
        
        return None

    def process_response(self, request, response, spider):
        # 确保响应是HtmlResponse
        if isinstance(response, HtmlResponse):
            # 处理重定向
            if response.status in [301, 302, 303, 307, 308]:
                return response
            
            # 记录响应信息
            content_type = response.headers.get('Content-Type', b'').decode('utf-8', 'ignore')
            logger.debug(f"Response from {request.url} - Type: {content_type}")
        
        return response

    def process_exception(self, request, exception, spider):
        logger.error(f"Error processing {request.url}: {exception}")
        return None

    def spider_opened(self, spider):
        spider.logger.info(f'Spider opened: {spider.name}')