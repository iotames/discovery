# spiders/dsite.py
import scrapy
from scrapy.spiders import CrawlSpider, Rule
from scrapy.linkextractors import LinkExtractor
from urllib.parse import urlparse, urljoin
import re
from sitedown.items import DsiteMirrorItem

class DsiteSpider(CrawlSpider):
    name = 'dsite'
    
    def __init__(self, start_url='https://www.santic-oem.com/', *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.start_urls = [start_url]
        self.allowed_domains = ['www.santic-oem.com']
        self.downloaded_resources = {}  # 记录已下载的资源
        
        # 配置LinkExtractor
        self.rules = (
            Rule(LinkExtractor(
                allow_domains=self.allowed_domains,
                deny_extensions=[],
                unique=True,
                restrict_css=None,
                restrict_xpaths=None,
            ), callback='parse_page', follow=True),
        )
        
        # 初始化规则
        self._compile_rules()
    
    def parse_start_url(self, response):
        """处理起始URL"""
        return self.parse_page(response)
    
    def parse_page(self, response):
        """解析页面，提取资源链接"""
        item = DsiteMirrorItem()
        item['url'] = response.url
        item['html_content'] = response.body
        item['content_type'] = response.headers.get('Content-Type', b'').decode('utf-8', 'ignore')
        item['depth'] = response.meta.get('depth', 0)
        
        # 提取所有资源链接
        item['css_links'] = self.extract_css_links(response)
        item['js_links'] = self.extract_js_links(response)
        item['image_links'] = self.extract_image_links(response)
        item['font_links'] = self.extract_font_links(response)
        item['media_links'] = self.extract_media_links(response)
        item['fetch_xhr_links'] = self.extract_fetch_xhr_links(response)

        # 收集所有资源URL，分别放入对应字段
        all_links = []
        all_links.extend(item['css_links'])
        all_links.extend(item['js_links'])
        all_links.extend(item['font_links'])
        all_links.extend(item['media_links'])
        all_links.extend(item['fetch_xhr_links'])
        item['file_urls'] = all_links
        item['image_urls'] = item['image_links']
        
        yield item
    
    def extract_css_links(self, response):
        """提取CSS链接"""
        css_links = []
        
        # 从link标签提取
        css_links.extend(response.css('link[rel="stylesheet"]::attr(href)').getall())
        
        # 处理相对路径
        css_links = [self.make_absolute(url, response) for url in css_links]
        
        return css_links
    
    def extract_js_links(self, response):
        """提取JS链接"""
        js_links = []
        
        # 从script标签提取
        js_links.extend(response.css('script[src]::attr(src)').getall())
        
        # 处理相对路径
        js_links = [self.make_absolute(url, response) for url in js_links]
        
        return js_links
    
    def extract_image_links(self, response):
        """提取图片链接"""
        image_links = []
        
        # 从img标签提取
        image_links.extend(response.css('img::attr(src)').getall())
        image_links.extend(response.css('img::attr(data-src)').getall())
        image_links.extend(response.css('img::attr(data-lazy-src)').getall())
        
        # 从picture/source标签提取
        image_links.extend(response.css('source::attr(srcset)').getall())
        image_links.extend(response.css('source::attr(src)').getall())
        
        # 处理srcset（包含多个图片）
        srcset_links = response.css('img::attr(srcset)').getall()
        for srcset in srcset_links:
            for part in srcset.split(','):
                url = part.strip().split(' ')[0]
                if url:
                    image_links.append(url)
        
        # 处理特殊路径
        processed_links = []
        for url in image_links:
            if url:
                # 处理cdn-cgi/image/路径
                if '/cdn-cgi/image/' in url:
                    url = url.replace('/cdn-cgi/image/w=1920,format=auto/files/', '/files/')
                    url = url.replace('/cdn-cgi/image/', '/files/')
                processed_links.append(self.make_absolute(url, response))
        
        return processed_links
    
    def extract_font_links(self, response):
        """提取字体链接"""
        font_links = []
        
        # 从CSS中提取字体链接
        css_text = response.text
        font_patterns = [
            r'url\(["\']?([^"\')]+\.(?:woff|woff2|ttf|eot|otf|svg)[^"\')]*)["\']?\)',
            r'src:\s*url\(["\']?([^"\')]+\.(?:woff|woff2|ttf|eot|otf|svg)[^"\')]*)["\']?\)',
        ]
        
        for pattern in font_patterns:
            matches = re.findall(pattern, css_text, re.IGNORECASE)
            font_links.extend(matches)
        
        # 从link标签提取
        font_links.extend(response.css('link[rel*="font"][href*=".woff"], '
                                      'link[rel*="font"][href*=".woff2"], '
                                      'link[rel*="font"][href*=".ttf"], '
                                      'link[rel*="font"][href*=".eot"], '
                                      'link[rel*="font"][href*=".otf"]::attr(href)').getall())
        
        # 处理相对路径
        font_links = [self.make_absolute(url, response) for url in font_links]
        
        return font_links
    
    def extract_media_links(self, response):
        """提取媒体文件链接"""
        media_links = []
        
        # 从video/audio/source标签提取
        media_links.extend(response.css('video::attr(src)').getall())
        media_links.extend(response.css('audio::attr(src)').getall())
        media_links.extend(response.css('source::attr(src)').getall())
        
        # 从track标签提取
        media_links.extend(response.css('track::attr(src)').getall())
        
        # 处理相对路径
        media_links = [self.make_absolute(url, response) for url in media_links]
        
        return media_links
    
    def extract_fetch_xhr_links(self, response):
        """提取XHR/Fetch请求链接"""
        xhr_links = []
        
        # 从script标签中提取可能的API请求
        scripts = response.css('script:not([src])').getall()
        for script in scripts:
            # 查找常见的API模式
            patterns = [
                r'["\'](https?://[^"\']+/api/[^"\']+)["\']',
                r'["\'](https?://[^"\']+/ajax/[^"\']+)["\']',
                r'["\'](https?://[^"\']+/json/[^"\']+)["\']',
                r'["\'](https?://[^"\']+/data/[^"\']+)["\']',
                r'fetch\(["\']([^"\']+)["\']',
                r'\.get\(["\']([^"\']+)["\']',
                r'\.post\(["\']([^"\']+)["\']',
                r'\.ajax\(["\']([^"\']+)["\']',
                r'url:\s*["\']([^"\']+)["\']',
            ]
            
            for pattern in patterns:
                matches = re.findall(pattern, script, re.IGNORECASE)
                xhr_links.extend(matches)
        
        # 处理相对路径
        xhr_links = [self.make_absolute(url, response) for url in xhr_links]
        
        return xhr_links
    
    def make_absolute(self, url, response):
        """将相对URL转换为绝对URL"""
        if not url:
            return url
        
        # 处理data URL
        if url.startswith('data:'):
            return url
        
        # 处理javascript URL
        if url.startswith(('javascript:', 'mailto:', 'tel:', '#')):
            return url
        
        # 处理协议相对URL
        if url.startswith('//'):
            return f'{response.url.split(":")[0]}:{url}'
        
        # 处理相对路径
        if not url.startswith(('http://', 'https://')):
            return urljoin(response.url, url)
        
        return url
    
    def process_fetch_xhr(self, response):
        """处理XHR响应，保存为JSON文件"""
        item = DsiteMirrorItem()
        item['url'] = response.url
        item['content_type'] = response.headers.get('Content-Type', b'').decode('utf-8', 'ignore')
        
        # 生成文件名
        from urllib.parse import urlparse
        import hashlib
        
        parsed = urlparse(response.url)
        path = parsed.path
        if not path:
            path = '/api'
        
        # 清理路径并生成文件名
        safe_path = re.sub(r'[^a-zA-Z0-9_\-/]', '_', path)
        if safe_path.endswith('/'):
            safe_path = safe_path[:-1]
        
        if not safe_path.endswith('.json'):
            safe_path = f"{safe_path}.json"
        
        # 创建目录结构
        domain = parsed.netloc.replace('www.', '').replace('.', '_').lower()
        file_path = f"downloads/fetch_xhr/{domain}/{safe_path}"
        
        import os
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        
        # 保存JSON文件
        try:
            content = response.body
            if content:
                with open(file_path, 'wb') as f:
                    f.write(content)
                self.logger.info(f"Saved XHR response: {response.url} -> {file_path}")
        except Exception as e:
            self.logger.error(f"Failed to save XHR response {response.url}: {e}")