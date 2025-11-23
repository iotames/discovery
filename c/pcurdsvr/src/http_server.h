#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H

#include "mongoose.h"  // 第三方库: Mongoose v7.19

// 服务器管理函数
void http_server_init(const char* url);
void http_server_run();
void http_server_cleanup();

// 事件处理函数
void event_handler(struct mg_connection* c, int ev, void* ev_data);

#endif