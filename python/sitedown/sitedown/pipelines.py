# pipelines.py
import os
import json
import hashlib
from pathlib import Path
from urllib.parse import urlparse, urljoin
from scrapy.pipelines.files import FilesPipeline
from scrapy.pipelines.images import ImagesPipeline
from scrapy.http import Request
from scrapy.exceptions import DropItem
import html
import re

# 在 pipelines.py 中，添加到 HtmlFilePipeline 类之前

class CustomFilesPipeline(FilesPipeline):
    """自定义文件管道，按类型和域名存储"""
    
    def file_path(self, request, response=None, info=None, *, item=None):
        # 调用父类方法获取原始路径（一个哈希文件名）
        original_path = super().file_path(request, response=response, info=info, item=item)
        
        # 从原始路径提取扩展名
        ext = original_path.split('.')[-1] if '.' in original_path else ''
        
        # 确定资源类型目录
        type_to_dir = {
            'css': 'css', 'js': 'js',
            'woff': 'fonts', 'woff2': 'fonts', 'ttf': 'fonts', 
            'eot': 'fonts', 'otf': 'fonts',
            'jpg': 'images', 'jpeg': 'images', 'png': 'images', 
            'gif': 'images', 'webp': 'images', 'bmp': 'images', 
            'ico': 'images', 'svg': 'images',
            'mp4': 'media', 'webm': 'media', 'ogg': 'media', 
            'mp3': 'media', 'wav': 'media', 'flac': 'media',
            'pdf': 'files', 'zip': 'files', 'rar': 'files', 
            'txt': 'files', 'doc': 'files', 'docx': 'files',
            'xls': 'files', 'xlsx': 'files', 'ppt': 'files', 'pptx': 'files',
        }
        
        # 默认目录
        resource_dir = type_to_dir.get(ext.lower(), 'files')
        
        # 获取域名目录
        from urllib.parse import urlparse
        parsed = urlparse(request.url)
        domain = parsed.netloc.replace('www.', '').replace('.', '_').lower()
        
        # 构建新路径：类型/域名/原始文件名
        new_path = f"{resource_dir}/{domain}/{original_path}"
        return new_path


class CustomImagesPipeline(ImagesPipeline):
    """自定义图片管道，按域名存储"""
    
    def file_path(self, request, response=None, info=None, *, item=None):
        # 调用父类方法获取原始路径
        original_path = super().file_path(request, response=response, info=info, item=item)
        
        # 获取域名目录
        from urllib.parse import urlparse
        parsed = urlparse(request.url)
        domain = parsed.netloc.replace('www.', '').replace('.', '_').lower()
        
        # 构建新路径：images/域名/原始文件名
        new_path = f"images/{domain}/{original_path}"
        return new_path

class HtmlFilePipeline:
    """处理HTML文件保存和链接替换的管道"""
    
    def __init__(self, settings):
        self.settings = settings
        self.base_dir = settings.get('BASE_DOWNLOAD_DIR', 'downloads')
        self.link_patterns = {
            'css': re.compile(r'(href=["\'])([^"\']+\.css[^"\']*)(["\'])', re.IGNORECASE),
            'js': re.compile(r'(src=["\'])([^"\']+\.js[^"\']*)(["\'])', re.IGNORECASE),
            'img': re.compile(r'(src=["\'])([^"\']+\.(?:jpg|jpeg|png|gif|webp|bmp|ico|svg)[^"\']*)(["\'])', re.IGNORECASE),
            'font': re.compile(r'(url\(["\']?)([^"\')]+\.(?:woff|woff2|ttf|eot|otf)[^"\')]*)(["\']?\))', re.IGNORECASE),
        }
    
    @classmethod
    def from_crawler(cls, crawler):
        return cls(crawler.settings)
    
    def get_relative_path(self, original_url, target_url, spider):
        """计算相对路径 (适配自定义管道)"""
        if not target_url.startswith(('http://', 'https://', '//')):
            return target_url

        from urllib.parse import urlparse
        parsed = urlparse(target_url)
        
        # 确定资源类型目录
        path = parsed.path
        ext = ''
        if '.' in path:
            ext = path.split('.')[-1].lower().split('?')[0]
        
        type_to_dir = {
            'css': 'css', 'js': 'js',
            'woff': 'fonts', 'woff2': 'fonts', 'ttf': 'fonts', 
            'eot': 'fonts', 'otf': 'fonts',
            'jpg': 'images', 'jpeg': 'images', 'png': 'images', 
            'gif': 'images', 'webp': 'images', 'bmp': 'images', 
            'ico': 'images', 'svg': 'images',
            'mp4': 'media', 'webm': 'media', 'ogg': 'media', 
            'mp3': 'media', 'wav': 'media',
        }
        
        resource_dir = type_to_dir.get(ext, 'files')
        domain = parsed.netloc.replace('www.', '').replace('.', '_').lower()
        
        # 构建相对路径：../资源类型/域名/文件名
        # 这里我们假设文件名由管道生成，使用一个简化的占位符
        # 实际中，HtmlFilePipeline 无法知道管道生成的具体文件名
        # 所以我们需要一个更简单的方案：统一路径格式
        
        # 简单方案：使用原始URL的最后部分作为文件名
        if path and '/' in path:
            filename = path.split('/')[-1]
            if '?' in filename:
                filename = filename.split('?')[0]
            if ext and not filename.endswith(f'.{ext}'):
                filename = f"{filename}.{ext}"
        else:
            filename = f"file.{ext}" if ext else "file"
        
        # 构建相对路径
        relative_path = f"../{resource_dir}/{domain}/{filename}"
        return relative_path
    
    def replace_links(self, html_content, item, spider):
        """替换HTML中的链接为相对路径"""
        if not html_content:
            return html_content
        
        html_str = html_content.decode('utf-8', errors='ignore') if isinstance(html_content, bytes) else str(html_content)
        
        # 替换CSS链接
        def replace_css(match):
            prefix = match.group(1)
            url = match.group(2)
            suffix = match.group(3)
            
            # 处理特殊路径
            if '/cdn-cgi/image/' in url:
                url = url.replace('/cdn-cgi/image/w=1920,format=auto/files/', '../files/')
                url = url.replace('/cdn-cgi/image/', '../files/')
            
            relative_url = self.get_relative_path(item['url'], url, spider)
            if relative_url != url:
                return f'{prefix}{relative_url}{suffix}'
            return match.group(0)
        
        html_str = self.link_patterns['css'].sub(replace_css, html_str)
        
        # 替换JS链接
        def replace_js(match):
            prefix = match.group(1)
            url = match.group(2)
            suffix = match.group(3)
            relative_url = self.get_relative_path(item['url'], url, spider)
            if relative_url != url:
                return f'{prefix}{relative_url}{suffix}'
            return match.group(0)
        
        html_str = self.link_patterns['js'].sub(replace_js, html_str)
        
        # 替换图片链接
        def replace_img(match):
            prefix = match.group(1)
            url = match.group(2)
            suffix = match.group(3)
            
            # 处理特殊路径
            if '/cdn-cgi/image/' in url:
                url = url.replace('/cdn-cgi/image/w=1920,format=auto/files/', '../images/')
                url = url.replace('/cdn-cgi/image/', '../images/')
            
            relative_url = self.get_relative_path(item['url'], url, spider)
            if relative_url != url:
                return f'{prefix}{relative_url}{suffix}'
            return match.group(0)
        
        html_str = self.link_patterns['img'].sub(replace_img, html_str)
        
        # 替换字体链接（在CSS中）
        def replace_font(match):
            prefix = match.group(1)
            url = match.group(2)
            suffix = match.group(3) if len(match.groups()) > 2 else ')'
            relative_url = self.get_relative_path(item['url'], url, spider)
            if relative_url != url:
                return f'{prefix}{relative_url}{suffix}'
            return match.group(0)
        
        html_str = self.link_patterns['font'].sub(replace_font, html_str)
        
        # 替换内联样式中的URL
        html_str = re.sub(
            r'url\(["\']?([^"\')]+)["\']?\)',
            lambda m: f'url({self.get_relative_path(item["url"], m.group(1), spider)})',
            html_str
        )
        
        return html_str.encode('utf-8') if isinstance(html_content, bytes) else html_str
    
    def get_html_path(self, url):
        """获取HTML文件的保存路径"""
        parsed = urlparse(url)
        path = parsed.path
        
        # 首页保存为index.html
        if path == '/' or path == '' or path.endswith('/'):
            filename = 'index.html'
            dir_path = parsed.netloc.replace('www.', '').replace('.', '_')
            dir_path = dir_path.lower()
        else:
            # 清理路径
            path = re.sub(r'[^a-zA-Z0-9_\-./]', '_', path)
            if path.startswith('/'):
                path = path[1:]
            
            if path.endswith('/'):
                path = path[:-1]
            
            # 确保以.html结尾
            if not path.endswith('.html'):
                if '.' in path:
                    # 移除原有扩展名
                    path = path.rsplit('.', 1)[0]
                path = f"{path}.html"
            
            filename = path
            dir_path = parsed.netloc.replace('www.', '').replace('.', '_')
            dir_path = dir_path.lower()
        
        # 创建目录
        html_dir = Path(self.base_dir) / dir_path
        html_dir.mkdir(parents=True, exist_ok=True)
        
        return str(html_dir / filename), dir_path
    
    def process_item(self, item, spider):
        if 'html_content' in item and item['html_content']:
            # 替换链接
            processed_html = self.replace_links(item['html_content'], item, spider)
            
            # 获取保存路径
            file_path, dir_name = self.get_html_path(item['url'])
            
            # 保存HTML文件
            try:
                with open(file_path, 'wb') as f:
                    if isinstance(processed_html, str):
                        f.write(processed_html.encode('utf-8'))
                    else:
                        f.write(processed_html)
                
                spider.logger.info(f"Saved HTML: {item['url']} -> {file_path}")
                item['file_path'] = file_path
                
                # 如果是首页，创建根目录的index.html
                if item['url'].endswith('/') or urlparse(item['url']).path in ['', '/']:
                    root_index = Path(self.base_dir) / 'index.html'
                    with open(root_index, 'wb') as f:
                        f.write(processed_html if isinstance(processed_html, bytes) else processed_html.encode('utf-8'))
                    
            except Exception as e:
                spider.logger.error(f"Failed to save HTML {item['url']}: {e}")
        
        return item