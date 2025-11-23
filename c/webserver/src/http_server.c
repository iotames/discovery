#include "http_server.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#ifdef _WIN32
#include <winsock2.h>
#else
#include <unistd.h>
#include <arpa/inet.h>
#endif
#include "router.h" // 路由处理

// 处理单个HTTP请求
static void handle_client(int client_fd) {
    char buffer[BUFFER_SIZE];
    int received = 0;
    memset(buffer, 0, sizeof(buffer));
#ifdef _WIN32
    received = recv(client_fd, buffer, sizeof(buffer) - 1, 0);
#else
    received = read(client_fd, buffer, sizeof(buffer) - 1);
#endif
    if (received > 0) {
        buffer[received] = '\0';
        route_request(client_fd, buffer);
    }
#ifdef _WIN32
    closesocket(client_fd);
#else
    close(client_fd);
#endif
}

// 启动HTTP服务器
void start_server(int port) {
#ifdef _WIN32
    WSADATA wsaData;
    WSAStartup(MAKEWORD(2,2), &wsaData);
#endif
    int server_fd;
    struct sockaddr_in addr;
    server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd < 0) {
        perror("socket error");
        exit(1);
    }
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(port);
    if (bind(server_fd, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
        perror("bind error");
        exit(1);
    }
    if (listen(server_fd, 10) < 0) {
        perror("listen error");
        exit(1);
    }
    printf("[INFO] 服务器已启动，等待连接...\n");
    while (1) {
        struct sockaddr_in client_addr;
#ifdef _WIN32
        int client_len = sizeof(client_addr);
#else
        socklen_t client_len = sizeof(client_addr);
#endif
        int client_fd = accept(server_fd, (struct sockaddr*)&client_addr, &client_len);
        if (client_fd < 0) {
            perror("accept error");
            continue;
        }
        handle_client(client_fd);
    }
#ifdef _WIN32
    closesocket(server_fd);
    WSACleanup();
#else
    close(server_fd);
#endif
}
