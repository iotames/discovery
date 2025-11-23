#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifdef _WIN32
#include <winsock2.h>
#pragma comment(lib, "ws2_32.lib")
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#endif

#define BUFFER_SIZE 4096

// 启动HTTP服务器，监听指定端口
typedef struct {
    int server_fd;
    int port;
} HttpServer;

void start_server(int port);

#endif // HTTP_SERVER_H
