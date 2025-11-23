#include "database.h"

static sqlite3* db = NULL;

int database_init(const char* db_path) {
    int rc = sqlite3_open(db_path, &db);
    if (rc != SQLITE_OK) {
        fprintf(stderr, "Cannot open database: %s\n", sqlite3_errmsg(db));
        return -1;
    }

    // 创建商品表
    const char* sql = "CREATE TABLE IF NOT EXISTS products ("
                      "id INTEGER PRIMARY KEY AUTOINCREMENT,"
                      "name TEXT NOT NULL,"
                      "description TEXT,"
                      "price REAL NOT NULL,"
                      "stock INTEGER DEFAULT 0,"
                      "create_time DATETIME DEFAULT CURRENT_TIMESTAMP);";
    
    char* err_msg = NULL;
    rc = sqlite3_exec(db, sql, 0, 0, &err_msg);
    if (rc != SQLITE_OK) {
        fprintf(stderr, "SQL error: %s\n", err_msg);
        sqlite3_free(err_msg);
        return -1;
    }

    printf("Database initialized successfully\n");
    return 0;
}

void database_close() {
    if (db) {
        sqlite3_close(db);
        db = NULL;
    }
}

int product_create(const char* name, const char* description, double price, int stock) {
    const char* sql = "INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?);";
    sqlite3_stmt* stmt;
    
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, NULL) != SQLITE_OK) {
        return -1;
    }
    
    sqlite3_bind_text(stmt, 1, name, -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, description, -1, SQLITE_STATIC);
    sqlite3_bind_double(stmt, 3, price);
    sqlite3_bind_int(stmt, 4, stock);
    
    int rc = sqlite3_step(stmt);
    sqlite3_finalize(stmt);
    
    return (rc == SQLITE_DONE) ? (int)sqlite3_last_insert_rowid(db) : -1;
}

int product_update(int id, const char* name, const char* description, double price, int stock) {
    const char* sql = "UPDATE products SET name=?, description=?, price=?, stock=? WHERE id=?;";
    sqlite3_stmt* stmt;
    
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, NULL) != SQLITE_OK) {
        return -1;
    }
    
    sqlite3_bind_text(stmt, 1, name, -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, description, -1, SQLITE_STATIC);
    sqlite3_bind_double(stmt, 3, price);
    sqlite3_bind_int(stmt, 4, stock);
    sqlite3_bind_int(stmt, 5, id);
    
    int rc = sqlite3_step(stmt);
    sqlite3_finalize(stmt);
    
    return (rc == SQLITE_DONE) ? 0 : -1;
}

int product_delete(int id) {
    const char* sql = "DELETE FROM products WHERE id=?;";
    sqlite3_stmt* stmt;
    
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, NULL) != SQLITE_OK) {
        return -1;
    }
    
    sqlite3_bind_int(stmt, 1, id);
    
    int rc = sqlite3_step(stmt);
    sqlite3_finalize(stmt);
    
    return (rc == SQLITE_DONE) ? 0 : -1;
}

product_t* product_get(int id) {
    const char* sql = "SELECT id, name, description, price, stock, create_time FROM products WHERE id=?;";
    sqlite3_stmt* stmt;
    product_t* product = NULL;
    
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, NULL) != SQLITE_OK) {
        return NULL;
    }
    
    sqlite3_bind_int(stmt, 1, id);
    
    if (sqlite3_step(stmt) == SQLITE_ROW) {
        product = (product_t*)malloc(sizeof(product_t));
        product->id = sqlite3_column_int(stmt, 0);
        strncpy(product->name, (const char*)sqlite3_column_text(stmt, 1), sizeof(product->name)-1);
        strncpy(product->description, (const char*)sqlite3_column_text(stmt, 2), sizeof(product->description)-1);
        product->price = sqlite3_column_double(stmt, 3);
        product->stock = sqlite3_column_int(stmt, 4);
        strncpy(product->create_time, (const char*)sqlite3_column_text(stmt, 5), sizeof(product->create_time)-1);
    }
    
    sqlite3_finalize(stmt);
    return product;
}

product_t** product_list(int* count) {
    const char* sql = "SELECT id, name, description, price, stock, create_time FROM products ORDER BY id;";
    sqlite3_stmt* stmt;
    product_t** products = NULL;
    int capacity = 10;
    int size = 0;
    
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, NULL) != SQLITE_OK) {
        return NULL;
    }
    
    products = (product_t**)malloc(capacity * sizeof(product_t*));
    
    while (sqlite3_step(stmt) == SQLITE_ROW) {
        if (size >= capacity) {
            capacity *= 2;
            products = (product_t**)realloc(products, capacity * sizeof(product_t*));
        }
        
        products[size] = (product_t*)malloc(sizeof(product_t));
        products[size]->id = sqlite3_column_int(stmt, 0);
        strncpy(products[size]->name, (const char*)sqlite3_column_text(stmt, 1), sizeof(products[size]->name)-1);
        strncpy(products[size]->description, (const char*)sqlite3_column_text(stmt, 2), sizeof(products[size]->description)-1);
        products[size]->price = sqlite3_column_double(stmt, 3);
        products[size]->stock = sqlite3_column_int(stmt, 4);
        strncpy(products[size]->create_time, (const char*)sqlite3_column_text(stmt, 5), sizeof(products[size]->create_time)-1);
        size++;
    }
    
    sqlite3_finalize(stmt);
    *count = size;
    return products;
}

void product_list_free(product_t** products, int count) {
    for (int i = 0; i < count; i++) {
        free(products[i]);
    }
    free(products);
}