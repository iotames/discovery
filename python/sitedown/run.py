# run.py
#!/usr/bin/env python3
import sys
import argparse
from scrapy.crawler import CrawlerProcess
from scrapy.utils.project import get_project_settings
from sitedown.spiders.dsite import DsiteSpider

def main():
    parser = argparse.ArgumentParser(description='网站镜像克隆工具')
    parser.add_argument('url', help='要克隆的网站URL')
    parser.add_argument('--output', '-o', default='downloads', 
                       help='输出目录（默认：downloads）')
    parser.add_argument('--delay', '-d', type=float, default=0.5,
                       help='下载延迟（默认：0.5秒）')
    parser.add_argument('--concurrent', '-c', type=int, default=16,
                       help='并发请求数（默认：16）')
    
    args = parser.parse_args()
    
    # 更新设置
    settings = get_project_settings()
    settings.set('BASE_DOWNLOAD_DIR', args.output)
    settings.set('DOWNLOAD_DELAY', args.delay)
    settings.set('CONCURRENT_REQUESTS', args.concurrent)
    
    # 创建并运行爬虫
    process = CrawlerProcess(settings)
    process.crawl(DsiteSpider, start_url=args.url)
    process.start()

if __name__ == '__main__':
    main()