package main

import (
	"fmt"
	"strconv"

	"github.com/iotames/easyserver"
	"github.com/iotames/easyserver/httpsvr"
	"github.com/iotames/easyserver/response"
)

func responseData(ctx httpsvr.Context, data map[string]any) {
	easyserver.ResponseJson(ctx, data, "ok", 0)
}

func responseFail(ctx httpsvr.Context, err error, code int) {
	easyserver.ResponseJsonFail(ctx, err.Error(), code)
}

func checkRequest(imgURI, sizeStr string) (width int, err error) {
	if imgURI == "" || sizeStr == "" {
		return 0, fmt.Errorf("参数缺失，请提供 imguri 和 size 参数。\n 例：http://127.0.0.1:%d/resizeimg?imguri=https://example.com/image.jpg&size=100。\n 本地文件可以使用 ./image.jpg 或 /path/to/image.jpg", WebPort)
	}

	width, err = strconv.Atoi(sizeStr)
	if err != nil || width <= 0 {
		return 0, fmt.Errorf("size 参数必须是正整数")
	}
	return width, nil
}

func newHttpServer() *httpsvr.EasyServer {
	s := easyserver.NewServer(fmt.Sprintf(":%d", WebPort))
	response.SetOkCode(0)
	return s
}
