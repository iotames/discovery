## C语言商品管理系统

一个使用C语言开发的简单HTTP服务器程序，提供商品的增删改查(CRUD)服务。该项目使用JSON格式进行数据交互，通过SQLite实现数据持久化存储。


## 项目特点

- 基于C语言的轻量级HTTP服务器
- 提供RESTful API接口实现商品管理功能
- 使用JSON格式进行数据交互
- 采用SQLite数据库进行数据持久化
- 模块化设计，代码结构清晰，易于维护和扩展


## 项目结构

初版：

```bash
├── mongoose.c          # HTTP服务器库
├── mongoose.h
├── cJSON.c             # JSON解析库
├── cJSON.h
├── product_server.c    # 主程序
├── Makefile
```

第二版：

```bash
├── src/                    # 源代码目录
│   ├── main.c              # 程序入口
│   ├── http_server.c       # HTTP服务器实现
│   ├── http_server.h       # HTTP服务器头文件
│   ├── database.c          # 数据库操作实现
│   ├── database.h          # 数据库操作头文件
│   ├── product_handler.c   # 商品处理逻辑实现
│   ├── product_handler.h   # 商品处理逻辑头文件
│   ├── cJSON.c             # JSON解析库
│   └── sqlite3.c           # SQLite数据库实现
├── include/                # 头文件目录
│   ├── cJSON.h             # JSON解析库头文件
│   ├── mongoose.h          # HTTP服务器库头文件
│   └── sqlite3.h           # SQLite数据库头文件
├── build/                  # 构建输出目录
├── Makefile                # 构建脚本
└── README.md               # 项目说明文件
```


## 第三方依赖库

本项目使用了以下第三方库：

### 1. Mongoose (v7.19)

- 项目主页: https://github.com/cesanta/mongoose
- 用途: HTTP服务器库，用于处理HTTP请求和响应
- 特点: 轻量级、易于集成、跨平台支持

### 2. cJSON (v1.7.19)

- 项目主页: https://github.com/DaveGamble/cJSON
- 用途: JSON解析库，用于处理请求和响应的JSON数据
- 特点: 轻量级、ANSI C兼容、易于使用

### 3. SQLite (v3.51.0 amalgamation)

- 官方网站: https://www.sqlite.org/
- 用途: 嵌入式数据库，用于数据持久化存储
- 特点: 无需独立服务器、零配置、事务支持

- 下载主页：https://www.sqlite.org/download.html

1. https://www.sqlite.org/2025/sqlite-src-3510000.zip 13.54M 完整原始版本。这是发布时受版本控制的所有代码的快照。所有其他源码包都从此包派生出去。当前版本：`3.51.0` 
2. https://www.sqlite.org/2025/sqlite-amalgamation-3510000.zip 2.74M 合并压缩版本。超过 100 个单独的源文件被连接成一个名为 `sqlite3.c` 的 C 代码大文件。
3. https://www.sqlite.org/2025/sqlite-doc-3510000.zip 10.91 MB 文档作为静态 HTML 文件的捆绑包。


## API接口说明

### 1. 创建商品
```
POST /api/product/create
Content-Type: application/json

{
  "name": "商品名称",
  "price": 价格,
  "description": "商品描述"
}
```

### 2. 查询商品列表
```
GET /api/product/list
```

### 3. 查询单个商品
```
GET /api/product/get?id=商品ID
```

### 4. 更新商品
```
POST /api/product/update
Content-Type: application/json

{
  "id": 商品ID,
  "name": "新商品名称",
  "price": 新价格,
  "description": "新商品描述"
}
```

### 5. 删除商品
```
POST /api/product/delete
Content-Type: application/json

{
  "id": 商品ID
}
```


## 构建说明

### 环境要求

- GCC编译器 (推荐版本: 15.2.0)
- GNU Make (推荐版本: 4.3)
- 标准C库

### 依赖下载

```bash
# 下载并安装SQLite
# sudo apt install libsqlite3-dev
wget -c https://www.sqlite.org/2025/sqlite-amalgamation-3510000.zip
unzip sqlite-amalgamation-3510000.zip
cp sqlite-amalgamation-3510000/sqlite3.c src/
cp sqlite-amalgamation-3510000/sqlite3.h include/

# 下载并安装Mongoose
# git clone https://github.com/cesanta/mongoose.git
# cp mongoose/mongoose.h mongoose/mongoose.c pcurdsvr/
wget -c https://github.com/cesanta/mongoose/archive/refs/tags/7.13.tar.gz
tar -xzf 7.13.tar.gz
cp mongoose-7.13/mongoose.c src/
cp mongoose-7.13/mongoose.h include/

# 下载并安装cJSON
# git clone https://github.com/DaveGamble/cJSON.git
# cp cJSON/cJSON.h cJSON/cJSON.c pcurdsvr/
wget -c https://github.com/DaveGamble/cJSON/archive/refs/tags/v1.7.19.tar.gz
tar -xzf v1.7.19.tar.gz
cp cJSON-1.7.17/cJSON.c src/
cp cJSON-1.7.17/cJSON.h include/
```

### 编译项目

```bash
# 编译项目
make

# 运行编译好的程序。相当于执行: ./build/your_file
make run

# 清理编译产物
make clean
```


## 使用示例

### 启动服务
```bash
# 编译并运行
make run

# 或者直接运行编译好的程序
./build/pcurdsvr
```

### API测试

```bash
# 创建商品
curl -X POST http://localhost:8000/api/product/create \
  -H "Content-Type: application/json" \
  -d '{"name":"iPhone 16","price":5999.99,"description":"The New iPhone"}'

# 查询商品列表
curl http://localhost:8000/api/product/list

# 查询单个商品
curl "http://localhost:8000/api/product/get?id=1"

# 更新商品
curl -X POST http://localhost:8000/api/product/update \
  -H "Content-Type: application/json" \
  -d '{"id":1,"name":"My HarmonyOS Phone","price":7999.99,"description":"The HUAWEI HarmonyOS phone"}'

# 删除商品
curl -X POST http://localhost:8000/api/product/delete \
  -H "Content-Type: application/json" \
  -d '{"id":1}'
```


## 代码结构说明

### 主要模块

1. **main.c**: 程序入口，初始化服务器并启动监听
2. **http_server.c/h**: HTTP服务器核心实现，处理连接和请求分发
3. **database.c/h**: 数据库操作封装，提供商品数据的增删改查接口
4. **product_handler.c/h**: 商品业务逻辑处理，解析请求并调用数据库接口
5. **cJSON.c/h**: JSON数据解析库
6. **mongoose.c/h**: HTTP服务器库
7. **sqlite3.c/h**: SQLite数据库库

### 设计特点

- 模块化设计：各功能模块职责明确，便于维护和扩展
- 松耦合：模块间通过接口交互，降低依赖性
- 易于测试：核心业务逻辑与HTTP处理分离
- 可扩展性：支持添加新的API接口和业务功能

## 注意事项

1. 项目默认监听8000端口，请确保该端口未被占用
2. 数据库文件(products.db)会在首次运行时自动创建
3. 为保证数据安全，生产环境应添加身份验证和输入验证机制
4. 当前实现为单线程模型，高并发场景下可能需要优化


## Prompt For AI

基本需求：
合理利用C语言生态的第三方组件库，做一个简单的HTTP服务器程序，以API接口的形式，提供商品的增删改查服务。

业务代码要求：
1. 完整实现每一个函数，项目要完整可用。绝对不能由于篇幅限制而只展示什么关键结构。
2. 要能解析POST提交的json商品数据。
3. 商品详情查询接口示例：GET /api/product/get?id=3。商品列表查询：GET /api/product/list。新增：POST /api/product/create。更新：POST /api/product/update。删除：POST /api/product/delete
4. 使用sqlite和SQL语句实现数据的持久化存储。

完整性要求【最重要】：
完整实现每一个函数，这是最重要的要求。如果由于篇幅限制无法实现，就直接说做不了，不用写了。不要长篇大论，然后缺胳膊断腿的。

数据格式：
业务数据的展示和交互使用JSON格式。

输出要求：
详细列出项目的每个依赖，要包含完整的构建教程，Makefile和源代码的注释要详尽。代码和构建都要符合工程化标准。

工程化要求：
1. 按项目工程化的实践标准，划分多个目录，拆分多个文件。各功能代码，按函数和文件拆分，明确职责边界，充分解耦。
2. 多处代码的共同部分，要抽离出来封装成函数。
3. 不论是C语言标准库还是第三方库，必须说明版本，还有是否使用某些扩展，否则经常符号找不到。
4. 代码中哪些符号是第三方库的引用，必须使用注释加以说明，必须带版本号。

第三方库要求：
1. 涉及的所有第三方库的选择，要有候选库对比，第三方库的资源链接或项目主页。
2. 最终选择的第三方库，要有获选理由和下载教程，使用教程。
3. 第三方库的获取方式：首选从官方URL或项目主页的源码文件获取，其次为官方编译好的链接库，最后是从apt install之类的命令行工具获取。
4. 第三方库必须明确注明版本，避免下载的版本和项目引用的版本不一致。


## 许可证

本项目仅供学习和研究使用，第三方库遵循其各自的许可证协议。