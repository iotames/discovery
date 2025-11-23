#include <stdio.h>
#include <string.h>
#ifdef _WIN32
#include <winsock2.h>
#else
#include <unistd.h>
#endif

// 路由处理函数声明
typedef void (*route_handler_t)(int client_fd);

// 发送HTTP响应
static void send_response(int client_fd, const char *status, const char *content_type, const char *body) {
    char response[2048];
    snprintf(response, sizeof(response),
        "HTTP/1.1 %s\r\nContent-Type: %s\r\nContent-Length: %zu\r\nConnection: close\r\n\r\n%s",
        status, content_type, strlen(body), body);
#ifdef _WIN32
    send(client_fd, response, (int)strlen(response), 0);
#else
    write(client_fd, response, strlen(response));
#endif
}

// 发送二进制文件响应（用于静态资源）
static void send_file_response(int client_fd, const char *filepath, const char *content_type) {
    FILE *fp = fopen(filepath, "rb");
    if (!fp) {
        send_response(client_fd, "404 Not Found", "text/plain", "File Not Found");
        return;
    }
    fseek(fp, 0, SEEK_END);
    long filesize = ftell(fp);
    fseek(fp, 0, SEEK_SET);
    char header[512];
    snprintf(header, sizeof(header),
        "HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %ld\r\nConnection: close\r\n\r\n",
        content_type, filesize);
#ifdef _WIN32
    send(client_fd, header, (int)strlen(header), 0);
#else
    write(client_fd, header, strlen(header));
#endif
    char buf[1024];
    size_t n;
    while ((n = fread(buf, 1, sizeof(buf), fp)) > 0) {
#ifdef _WIN32
        send(client_fd, buf, (int)n, 0);
#else
        write(client_fd, buf, n);
#endif
    }
    fclose(fp);
}

// 获取简单的Content-Type
static const char* get_content_type(const char *filename) {
    const char *ext = strrchr(filename, '.');
    if (!ext) return "application/octet-stream";
    if (strcmp(ext, ".html") == 0) return "text/html";
    if (strcmp(ext, ".css") == 0) return "text/css";
    if (strcmp(ext, ".js") == 0) return "application/javascript";
    if (strcmp(ext, ".png") == 0) return "image/png";
    if (strcmp(ext, ".jpg") == 0 || strcmp(ext, ".jpeg") == 0) return "image/jpeg";
    if (strcmp(ext, ".gif") == 0) return "image/gif";
    if (strcmp(ext, ".txt") == 0) return "text/plain";
    return "application/octet-stream";
}

// 处理GET /
static void handle_home(int client_fd) {
    send_response(client_fd, "200 OK", "text/plain", "this is home");
}

// 处理GET /helloapi
static void handle_helloapi(int client_fd) {
    send_response(client_fd, "200 OK", "application/json", "{\"code\":0,\"msg\":\"this is json api\"}");
}

// 处理POST /update
static void handle_update(int client_fd, const char *request) {
    // 查找请求体（body）
    const char *body = strstr(request, "\r\n\r\n");
    if (body) body += 4; else body = "";
    // 简单回显json
    char resp[512];
    snprintf(resp, sizeof(resp), "{\"code\":0,\"msg\":\"update ok\",\"data\":%s}", body);
    send_response(client_fd, "200 OK", "application/json", resp);
}

// 处理GET /static/xxx 静态资源
static void handle_static(int client_fd, const char *path) {
    // 安全性：不允许访问..等路径
    if (strstr(path, "..")) {
        send_response(client_fd, "403 Forbidden", "text/plain", "Forbidden");
        return;
    }
    char filepath[512];
    snprintf(filepath, sizeof(filepath), "static%s", path + 7); // 跳过/static
    const char *ctype = get_content_type(filepath);
    send_file_response(client_fd, filepath, ctype);
}

// 路由分发
void route_request(int client_fd, const char *request) {
    // 简单解析请求行
    char method[8], path[256];
    sscanf(request, "%7s %255s", method, path);
    if (strcmp(method, "GET") == 0 && strcmp(path, "/") == 0) {
        handle_home(client_fd);
    } else if (strcmp(method, "GET") == 0 && strcmp(path, "/helloapi") == 0) {
        handle_helloapi(client_fd);
    } else if (strcmp(method, "GET") == 0 && strncmp(path, "/static/", 8) == 0) {
        handle_static(client_fd, path);
    } else if (strcmp(method, "POST") == 0 && strcmp(path, "/update") == 0) {
        handle_update(client_fd, request);
    } else {
        send_response(client_fd, "404 Not Found", "text/plain", "404 Not Found");
    }
}
