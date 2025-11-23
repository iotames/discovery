# 编译器设置
CC = gcc
CFLAGS = -I./include -Wall -Wextra -std=c99 -D_DEFAULT_SOURCE \
         -DMG_ENABLE_HTTP=1 -DMG_ENABLE_LOG=1 -DMG_ENABLE_FILES=1 \
         -DMG_ENABLE_HTTP_WEBSOCKET=0 -DMG_ENABLE_MQTT=0 -DMG_ENABLE_SSI=0
LDFLAGS = -lpthread -ldl

# 目标文件
TARGET = $(BUILD_DIR)/pcurdsvr
BUILD_DIR = build
SRC_DIR = src

# 源文件
SOURCES = $(SRC_DIR)/main.c $(SRC_DIR)/http_server.c $(SRC_DIR)/database.c $(SRC_DIR)/product_handler.c $(SRC_DIR)/cJSON.c $(SRC_DIR)/sqlite3.c $(SRC_DIR)/mongoose.c
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(BUILD_DIR)/%.o)

# 默认目标
all: $(BUILD_DIR) $(TARGET)

# 创建构建目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# 链接目标文件
$(TARGET): $(OBJECTS)
	$(CC) -o $@ $^ $(LDFLAGS)

# 编译源文件
$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

# 清理
clean:
	rm -rf $(BUILD_DIR) $(TARGET) products.db

# 运行编译好的程序。换行后，命令行要使用<tab>键输入\t字符，不可用用空格符开头。
run: $(TARGET)
	./$(TARGET)

# 安装依赖 (需要手动下载sqlite3)
deps:
	@if [ ! -f include/sqlite3.h ] || [ ! -f src/sqlite3.c ]; then \
		echo "Downloading SQLite3..."; \
		wget -c https://www.sqlite.org/2025/sqlite-amalgamation-3510000.zip -O sqlite-amalgamation.zip; \
		apt install unzip; \
		unzip -o sqlite-amalgamation.zip; \
		mkdir -p include src; \
		cp sqlite-amalgamation-3510000/sqlite3.h include/; \
		cp sqlite-amalgamation-3510000/sqlite3.c src/; \
		rm -rf sqlite-amalgamation-3510000; \
	# 	rm -rf sqlite-amalgamation.zip \
		echo "Dependencies installed"; \
	else \
		echo "SQLite3 files already exist, skipping download"; \
	fi

.PHONY: all clean run deps
