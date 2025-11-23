CC=gcc
# 中文乱码，在CFLAGS添加-fexec-charset=UTF-8选项
CFLAGS=-Iinclude -Wall -Wextra -g
SRCDIR=src
OBJDIR=obj
SOURCES=$(SRCDIR)/main.c $(SRCDIR)/http_server.c $(SRCDIR)/router.c
OBJECTS=$(SOURCES:$(SRCDIR)/%.c=$(OBJDIR)/%.o)

# 根据操作系统设置目标文件名和链接库
ifeq ($(OS),Windows_NT)
	TARGET=webserver.exe
	LIBS=-lws2_32
	MKDIR=if not exist $(OBJDIR) mkdir $(OBJDIR)
	RM=del /Q
	RMOBJ=$(OBJDIR)\*.o
	RMTARGET=$(TARGET)
	RMREDIR=2>nul
else
	TARGET=webserver
	LIBS=
	MKDIR=mkdir -p $(OBJDIR)
	RM=rm -f
	RMOBJ=$(OBJDIR)/*.o
	RMTARGET=$(TARGET)
	RMREDIR=
endif

# 执行make或make all时，编译生成最终的可执行文件
all: $(TARGET)

$(TARGET): $(OBJECTS)
	$(CC) $(CFLAGS) $^ -o $@ $(LIBS)

$(OBJDIR)/%.o: $(SRCDIR)/%.c
	$(MKDIR)
	$(CC) $(CFLAGS) -c $< -o $@

# 用于清理编译生成的中间文件和可执行文件。执行 make clean 会删除 obj 目录下的 .o 文件和最终的可执行文件
clean:
	-$(RM) $(RMOBJ) $(RMREDIR) || true
	-$(RM) $(RMTARGET) $(RMREDIR) || true

# 用于运行编译好的程序。
# Windows命令窗口启动后，先执行chcp 65001，再运行程序。避免中文乱码问题
run: $(TARGET)
ifeq ($(OS),Windows_NT)
	chcp 65001 >nul
	$(TARGET)
else
	./$(TARGET)
endif

# 声明后面的目标（如 all、clean、run）不是实际的文件名，而是“伪目标”。
# .PHONY保证all、clean、run 始终被执行
.PHONY: all clean run