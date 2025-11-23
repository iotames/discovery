## C语言学习实战项目

使用AI学习C语言，并搭建实战项目。

## 本项目使用的开发环境

为避免开发环境不一致，导致各种报错。约定环境如下：

1. Windows 11 专业版 x64 24H2

2. WinLibs UCRT 版本: [winlibs-x86_64-posix-seh-gcc-15.2.0-mingw-w64ucrt-13.0.0-r2.zip](https://github.com/brechtsanders/winlibs_mingw/releases/download/15.2.0posix-13.0.0-ucrt-r2/winlibs-x86_64-posix-seh-gcc-15.2.0-mingw-w64ucrt-13.0.0-r2.zip)
开发调试都在Windows11下运行，使用


## 环境准备

需要安装 `gcc` 编译工具。


### Windows

- Windows下的C语言工具链有非常多。如：`MinGW-w64`, `TDM-GCC`, `WinLibs`(MinGW-w64的发行版)，`MSYS2`(集成MinGW工具链，并提供完整的类 Unix 开发环境)

- `推荐` [WinLibs](https://winlibs.com/): 下载解压后，将 bin 目录添加到系统 PATH 环境变量中，即可使用。无需复杂的安装和配置过程。
- 提示找不到 `make` 命令：重命名 `mingw32-make.exe` 为 `make.exe` 即可: `copy mingw32-make.exe make.exe`


`WinLibs` 下载链接：

- `推荐` UCRT 版本 - 现代 Windows 版本（Win10+）: https://github.com/brechtsanders/winlibs_mingw/releases/download/15.2.0posix-13.0.0-ucrt-r2/winlibs-x86_64-posix-seh-gcc-15.2.0-mingw-w64ucrt-13.0.0-r2.zip

- MSVCRT 版本 - 更好的旧版 Windows 兼容性：https://github.com/brechtsanders/winlibs_mingw/releases/download/15.2.0posix-13.0.0-msvcrt-r1/winlibs-x86_64-posix-seh-gcc-15.2.0-mingw-w64msvcrt-13.0.0-r1.zip


尽管 GCC 和 MinGW-w64 可以在其他平台（例如 Linux）上使用以生成 Windows 可执行文件，但 `WinLibs` 项目仅专注于构建可在 Windows 上原生运行的版本。


### Linux

```bash
apt update
apt install gcc
```


## 编译说明

```bash
# 清除编译缓存
make clean

# 执行编译命令
make
```

1. Makefile所有命令行前必须是 `Tab`，不能是空格。

在 Makefile 中，所有以目标（如 all:、clean:、$(TARGET): 等）开头的规则下方，每一行命令前都必须是Tab，不能是空格，否则会报 `missing separator` 错误。