#include <stdio.h>
#include <time.h>

int main() {
    time_t rawtime;
    struct tm *timeinfo;
    char buffer[80];

    // 获取当前时间戳
    time(&rawtime);
    // 转换为本地时间结构体
    timeinfo = localtime(&rawtime);
    // 格式化时间为"年-月-日 时:分:秒"格式
    strftime(buffer, sizeof(buffer), "%Y-%m-%d %H:%M:%S", timeinfo);
    // 输出结果
    printf("hello: %s\n", buffer);

    return 0;
}