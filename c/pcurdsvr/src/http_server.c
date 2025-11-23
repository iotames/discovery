#include "http_server.h"
#include "product_handler.h"
#include "database.h"

static struct mg_mgr mgr;

// 修复函数签名，移除第四个参数
void event_handler(struct mg_connection* c, int ev, void* ev_data) {
    if (ev == MG_EV_HTTP_MSG) {
        struct mg_http_message* hm = (struct mg_http_message*)ev_data;
        
        // 使用 mg_strcmp 进行 URI 比较
        if (mg_strcmp(hm->uri, mg_str("/api/product/get")) == 0) {
            handle_product_get(c, hm);
        } else if (mg_strcmp(hm->uri, mg_str("/api/product/list")) == 0) {
            handle_product_list(c, hm);
        } else if (mg_strcmp(hm->uri, mg_str("/api/product/create")) == 0) {
            handle_product_create(c, hm);
        } else if (mg_strcmp(hm->uri, mg_str("/api/product/update")) == 0) {
            handle_product_update(c, hm);
        } else if (mg_strcmp(hm->uri, mg_str("/api/product/delete")) == 0) {
            handle_product_delete(c, hm);
        } else {
            mg_http_reply(c, 404, "Content-Type: text/plain\r\n", "Not Found");
        }
    }
}

void http_server_init(const char* url) {
    mg_mgr_init(&mgr);
    // 修复：使用正确的函数签名
    struct mg_connection *nc = mg_http_listen(&mgr, url, event_handler, &mgr);
    if (nc == NULL) {
        fprintf(stderr, "Failed to start server on %s\n", url);
        return;
    }
    printf("HTTP server listening on %s\n", url);
}

// 将无限循环改为单次轮询，返回到 main 检查信号
void http_server_run() {
    // for (;;) {
    //     mg_mgr_poll(&mgr, 1000);
    // }
    mg_mgr_poll(&mgr, 1000);
}

void http_server_cleanup() {
    mg_mgr_free(&mgr);
}