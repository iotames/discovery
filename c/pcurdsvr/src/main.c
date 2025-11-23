#include <stdio.h>
#include <signal.h>
#include "http_server.h"
#include "database.h"

// 使用可在信号处理函数中安全修改的类型
static volatile sig_atomic_t s_signo = 0;

static void signal_handler(int signo) {
    s_signo = signo;
}

int main() {
    // 设置信号处理
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    printf("Starting Product CRUD Server...\n");
    
    // 初始化数据库
    if (database_init("products.db") != 0) {
        fprintf(stderr, "Failed to initialize database\n");
        return 1;
    }
    
    // 启动HTTP服务器
    http_server_init("http://0.0.0.0:8000");
    
    printf("Server started. Press Ctrl+C to stop.\n");
    
    // 主循环
    while (s_signo == 0) {
        http_server_run();
    }
    
    printf("Received signal %d, shutting down...\n", s_signo);
    
    // 清理资源
    http_server_cleanup();
    database_close();
    
    printf("Server stopped.\n");
    return 0;
}