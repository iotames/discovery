#include "product_handler.h"
#include "database.h"

void send_json_response(struct mg_connection* c, int status_code, const char* message, cJSON* data) {
    cJSON* root = cJSON_CreateObject();
    cJSON_AddNumberToObject(root, "code", status_code);
    cJSON_AddStringToObject(root, "message", message);
    
    if (data) {
        cJSON_AddItemToObject(root, "data", data);
    }
    
    char* json_str = cJSON_PrintUnformatted(root);
    mg_http_reply(c, status_code, "Content-Type: application/json\r\n", "%s", json_str);
    
    free(json_str);
    cJSON_Delete(root);
}

cJSON* product_to_json(product_t* product) {
    if (!product) return NULL;
    
    cJSON* json = cJSON_CreateObject();
    cJSON_AddNumberToObject(json, "id", product->id);
    cJSON_AddStringToObject(json, "name", product->name);
    cJSON_AddStringToObject(json, "description", product->description);
    cJSON_AddNumberToObject(json, "price", product->price);
    cJSON_AddNumberToObject(json, "stock", product->stock);
    cJSON_AddStringToObject(json, "create_time", product->create_time);
    return json;
}

void handle_product_get(struct mg_connection* c, struct mg_http_message* hm) {
    char id_str[32] = {0};
    
    // 从查询字符串获取id
    if (mg_http_get_var(&hm->query, "id", id_str, sizeof(id_str)) <= 0) {
        send_json_response(c, 400, "Missing product id", NULL);
        return;
    }
    
    int id = atoi(id_str);
    product_t* product = product_get(id);
    
    if (product) {
        cJSON* data = product_to_json(product);
        send_json_response(c, 200, "Success", data);
        free(product);
    } else {
        send_json_response(c, 404, "Product not found", NULL);
    }
}

void handle_product_list(struct mg_connection* c, struct mg_http_message* hm) {
    (void)hm;  // 标记为未使用，消除警告
    
    int count = 0;
    product_t** products = product_list(&count);
    
    cJSON* data = cJSON_CreateArray();
    for (int i = 0; i < count; i++) {
        cJSON* product_json = product_to_json(products[i]);
        cJSON_AddItemToArray(data, product_json);
    }
    
    send_json_response(c, 200, "Success", data);
    product_list_free(products, count);
}

void handle_product_create(struct mg_connection* c, struct mg_http_message* hm) {
    // 解析JSON请求体
    struct mg_str body = hm->body;
    if (body.len == 0) {
        send_json_response(c, 400, "Empty request body", NULL);
        return;
    }
    
    char* body_str = malloc(body.len + 1);
    memcpy(body_str, body.buf, body.len);  // 修复：使用 body.buf 而不是 body.ptr
    body_str[body.len] = '\0';
    
    cJSON* json = cJSON_Parse(body_str);
    free(body_str);
    
    if (!json) {
        send_json_response(c, 400, "Invalid JSON", NULL);
        return;
    }
    
    cJSON* name = cJSON_GetObjectItemCaseSensitive(json, "name");
    cJSON* description = cJSON_GetObjectItemCaseSensitive(json, "description");
    cJSON* price = cJSON_GetObjectItemCaseSensitive(json, "price");
    cJSON* stock = cJSON_GetObjectItemCaseSensitive(json, "stock");
    
    if (!cJSON_IsString(name) || !cJSON_IsNumber(price)) {
        cJSON_Delete(json);
        send_json_response(c, 400, "Missing required fields", NULL);
        return;
    }
    
    const char* desc = description ? description->valuestring : "";
    int stock_val = stock ? stock->valueint : 0;
    
    int product_id = product_create(name->valuestring, desc, price->valuedouble, stock_val);
    cJSON_Delete(json);
    
    if (product_id > 0) {
        cJSON* data = cJSON_CreateObject();
        cJSON_AddNumberToObject(data, "id", product_id);
        send_json_response(c, 201, "Product created", data);
    } else {
        send_json_response(c, 500, "Failed to create product", NULL);
    }
}

void handle_product_update(struct mg_connection* c, struct mg_http_message* hm) {
    // char id_str[32] = {0};
    
    // // 从查询字符串获取id
    // if (mg_http_get_var(&hm->query, "id", id_str, sizeof(id_str)) <= 0) {
    //     send_json_response(c, 400, "Missing product id", NULL);
    //     return;
    // }
    
    // int id = atoi(id_str);
    
    // 解析JSON请求体
    struct mg_str body = hm->body;
    if (body.len == 0) {
        send_json_response(c, 400, "Empty request body", NULL);
        return;
    }
    
    char* body_str = malloc(body.len + 1);
    memcpy(body_str, body.buf, body.len);  // 修复：使用 body.buf 而不是 body.ptr
    body_str[body.len] = '\0';
    
    cJSON* json = cJSON_Parse(body_str);
    free(body_str);
    
    if (!json) {
        send_json_response(c, 400, "Invalid JSON", NULL);
        return;
    }

    // 从JSON中获取id
    cJSON* id_obj = cJSON_GetObjectItemCaseSensitive(json, "id");
    if (!cJSON_IsNumber(id_obj)) {
        cJSON_Delete(json);
        send_json_response(c, 400, "Missing or invalid product id", NULL);
        return;
    }
    int id = id_obj->valueint;

    cJSON* name = cJSON_GetObjectItemCaseSensitive(json, "name");
    cJSON* description = cJSON_GetObjectItemCaseSensitive(json, "description");
    cJSON* price = cJSON_GetObjectItemCaseSensitive(json, "price");
    cJSON* stock = cJSON_GetObjectItemCaseSensitive(json, "stock");
    
    if (!cJSON_IsString(name) || !cJSON_IsNumber(price)) {
        cJSON_Delete(json);
        send_json_response(c, 400, "Missing required fields", NULL);
        return;
    }
    
    const char* desc = description ? description->valuestring : "";
    int stock_val = stock ? stock->valueint : 0;
    
    int result = product_update(id, name->valuestring, desc, price->valuedouble, stock_val);
    cJSON_Delete(json);
    
    if (result == 0) {
        send_json_response(c, 200, "Product updated", NULL);
    } else {
        send_json_response(c, 500, "Failed to update product", NULL);
    }
}

void handle_product_delete(struct mg_connection* c, struct mg_http_message* hm) {
    // char id_str[32] = {0};
    // // 从查询字符串获取id
    // if (mg_http_get_var(&hm->query, "id", id_str, sizeof(id_str)) <= 0) {
    //     send_json_response(c, 400, "Missing product id", NULL);
    //     return;
    // }
    // int id = atoi(id_str);

    // 解析JSON请求体
    struct mg_str body = hm->body;
    if (body.len == 0) {
        send_json_response(c, 400, "Empty request body", NULL);
        return;
    }
    char* body_str = malloc(body.len + 1);
    memcpy(body_str, body.buf, body.len);  // 修复：使用 body.buf 而不是 body.ptr
    body_str[body.len] = '\0';
    cJSON* json = cJSON_Parse(body_str);
    free(body_str);
    if (!json) {
        send_json_response(c, 400, "Invalid JSON", NULL);
        return;
    }

    // 从JSON中获取id
    cJSON* id_obj = cJSON_GetObjectItemCaseSensitive(json, "id");
    if (!cJSON_IsNumber(id_obj)) {
        cJSON_Delete(json);
        send_json_response(c, 400, "Missing or invalid product id", NULL);
        return;
    }
    int id = id_obj->valueint;

    int result = product_delete(id);
    
    if (result == 0) {
        send_json_response(c, 200, "Product deleted", NULL);
    } else {
        send_json_response(c, 500, "Failed to delete product", NULL);
    }
}