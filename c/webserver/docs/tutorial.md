# C语言简易HTTP服务器实现教程

## 1. 项目结构与多文件拆分

- `main.c`：程序入口，负责解析端口参数并启动服务器。
- `http_server.c`/`http_server.h`：服务器主循环、socket初始化、连接处理。
- `router.c`：路由分发，处理不同URL路径。
- `static/`：静态资源目录，后续可扩展静态文件服务。
- `include/`：头文件目录，声明跨文件函数和结构体。
- `Makefile`：编译脚本。

## 2. 变量和函数跨文件引用说明

- 头文件（如`http_server.h`）声明函数和结构体。
- 源文件通过`#include "http_server.h"`引用声明。
- 例如：
  - `main.c`调用`start_server`，只需包含`http_server.h`。
  - `http_server.c`实现`start_server`。
    - 路由相关函数在`router.c`实现，`http_server.c`通过`#include "router.h"`声明并调用（避免多重编译问题）。

## 3. 编译与运行

1. 安装gcc（如MinGW）
2. 在webserver目录下执行：
   ```sh
   make
   ```
3. 生成`webserver.exe`，运行：
   ```sh
   ./webserver.exe 8080
   ```

## 4. 代码详解

### main.c
```c
#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char *argv[]) {
    int port = 8080;
    if (argc > 1) {
        port = atoi(argv[1]);
        if (port <= 0) port = 8080;
    }
    start_server(port);
    return 0;
}
```

### http_server.h
```c
#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H
void start_server(int port);
#endif
```


### http_server.c
```c
#include "http_server.h"
#include "router.h"
// ...socket初始化、循环accept、调用route_request...
```

### router.c
```c
void route_request(int client_fd, const char *request) {
    // 解析请求，分发到不同处理函数
}
```


## 5. 路由扩展与增量实现说明

- 只需在`router.c`中添加新的路径判断和处理函数。
- 例如：
  - `/` 返回"this is home"
  - `/helloapi` 返回json

### 新增 `/static` 静态资源路由
1. 在`router.c`中增加如下内容：
   - 新增`handle_static`函数，解析/static/路径，拼接static目录下的文件路径，读取并返回文件内容。
   - 新增`get_content_type`函数，简单判断文件类型。
   - 在`route_request`中添加`/static/`前缀判断，调用`handle_static`。
2. 访问如`/static/test.txt`会返回static目录下的test.txt内容。

### 新增 `/update` POST接口
1. 在`router.c`中增加如下内容：
   - 新增`handle_update`函数，查找HTTP请求体（body），简单回显json。
   - 在`route_request`中添加`POST /update`判断，调用`handle_update`。
2. 用curl等工具POST json到`/update`，会返回带原始数据的json响应。

### 代码片段示例
```c
// 路由分发
void route_request(int client_fd, const char *request) {
    char method[8], path[256];
    sscanf(request, "%7s %255s", method, path);
    if (strcmp(method, "GET") == 0 && strncmp(path, "/static/", 8) == 0) {
        handle_static(client_fd, path);
    } else if (strcmp(method, "POST") == 0 && strcmp(path, "/update") == 0) {
        handle_update(client_fd, request);
    }
    // ...
}
```

详细实现见`src/router.c`，注释已详细说明。

## 6. 后续扩展建议
- 静态文件服务：添加`/static`路由，映射`static/`目录。
- POST/PUT等方法支持。
- 多线程/多进程处理并发。
- 日志、配置文件、错误处理等。

---

详细代码和注释请参考各源文件。
