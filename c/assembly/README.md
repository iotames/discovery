## 介绍

汇编语言学习笔记


## Linux

```bash
# 安装nasm汇编器
# Ubuntu
sudo apt install nasm

# 编译

# 编译链接一个 32位 程序
nasm -f elf32 hello_linux.asm -o hello32.o 
ld -m elf_i386 hello32.o -o hello32
./hello32

# 编译链接一个 64位 程序
nasm -f elf64 -o hello64.o hello_linux.asm
ld -m elf_x86_64 -o hello64 hello64.o
./hello64
```

## Windows

```bash
# 在Linux平台下编译
nasm -f win64 hello_win64.asm -o hello_win64.o
# 在Windows平台下使用MinGW-W64方案的gcc命令链接程序
gcc hello_win64.o -o hello64.exe
```
