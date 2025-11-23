#ifndef DATABASE_H
#define DATABASE_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "sqlite3.h"  // 第三方库: SQLite3 v3.51.0

// 商品结构体
typedef struct product_s {
    int id;
    char name[100];
    char description[200];
    double price;
    int stock;
    char create_time[20];
} product_t;

// 数据库初始化
int database_init(const char* db_path);

// 关闭数据库
void database_close();

// 商品CRUD操作
int product_create(const char* name, const char* description, double price, int stock);
int product_update(int id, const char* name, const char* description, double price, int stock);
int product_delete(int id);
product_t* product_get(int id);
product_t** product_list(int* count);

// 释放商品列表内存
void product_list_free(product_t** products, int count);

#endif