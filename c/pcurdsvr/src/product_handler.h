#ifndef PRODUCT_HANDLER_H
#define PRODUCT_HANDLER_H

#include "mongoose.h"  // 第三方库: Mongoose v7.19
#include "cJSON.h"     // 第三方库: cJSON v1.7.19

// 前向声明，避免循环依赖
typedef struct product_s product_t;

// API路由处理函数
void handle_product_get(struct mg_connection* c, struct mg_http_message* hm);
void handle_product_list(struct mg_connection* c, struct mg_http_message* hm);
void handle_product_create(struct mg_connection* c, struct mg_http_message* hm);
void handle_product_update(struct mg_connection* c, struct mg_http_message* hm);
void handle_product_delete(struct mg_connection* c, struct mg_http_message* hm);

// 工具函数
void send_json_response(struct mg_connection* c, int status_code, const char* message, cJSON* data);
cJSON* product_to_json(product_t* product);

#endif