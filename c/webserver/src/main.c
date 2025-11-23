#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>

// 主程序入口，解析端口参数并启动服务器
int main(int argc, char *argv[]) {
    int port = 8080; // 默认端口
    if (argc > 1) {
        port = atoi(argv[1]);
        if (port <= 0) {
            printf("端口号无效，使用默认8080\n");
            port = 8080;
        }
    }
    printf("[INFO] 启动HTTP服务器，监听端口: %d\n", port);
    start_server(port);
    return 0;
}
