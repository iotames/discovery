# items.py
import scrapy

class DsiteMirrorItem(scrapy.Item):
    # 基础字段
    url = scrapy.Field()
    html_content = scrapy.Field()
    content_type = scrapy.Field()
    file_path = scrapy.Field()

    # 资源链接
    css_links = scrapy.Field()
    js_links = scrapy.Field()
    image_links = scrapy.Field()
    font_links = scrapy.Field()
    media_links = scrapy.Field()
    fetch_xhr_links = scrapy.Field()
    
    # 用于文件下载
    file_urls = scrapy.Field()
    files = scrapy.Field()
    image_urls = scrapy.Field()
    images = scrapy.Field()
    
    # 元数据
    referer = scrapy.Field()
    depth = scrapy.Field()