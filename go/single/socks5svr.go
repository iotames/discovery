package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// SOCKS5 协议常量
const (
	socks5Version = 0x05

	// 命令
	cmdConnect = 0x01 // CONNECT 命令

	// 地址类型
	atypIPv4   = 0x01 // IPv4 地址
	atypDomain = 0x03 // 域名

	// 响应码
	repSuccess                 = 0x00 // 成功
	repConnectionRefused       = 0x05 // 连接被拒绝
	repCommandNotSupported     = 0x07 // 不支持的命令
	repAddressTypeNotSupported = 0x08 // 不支持的地址类型

	// 认证方法
	authNoAuth = 0x00 // 无认证
)

func main() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		fmt.Printf("监听失败: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("SOCKS5 代理服务器启动，监听 :1080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}

// handleClient 处理单个客户端连接
func handleClient(client net.Conn) {
	defer client.Close()

	// 设置读取超时，防止恶意连接长时间占用
	client.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 1. 认证协商（仅支持无认证）
	if !handleAuth(client) {
		return
	}

	// 清除读取超时，后续转发不设限
	client.SetReadDeadline(time.Time{})

	// 2. 处理请求（仅支持 CONNECT）
	handleRequest(client)
}

// handleAuth 处理客户端认证协商（仅支持无认证）
// 返回 true 表示认证成功，false 表示失败
func handleAuth(client net.Conn) bool {
	// 读取版本号和客户端支持的认证方法数量
	header := make([]byte, 2)
	if _, err := io.ReadFull(client, header); err != nil {
		fmt.Printf("读取认证头失败: %v\n", err)
		return false
	}

	ver, nmethods := header[0], header[1]
	if ver != socks5Version {
		fmt.Printf("不支持的 SOCKS 版本: %d\n", ver)
		return false
	}

	// 读取客户端支持的认证方法列表
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(client, methods); err != nil {
		fmt.Printf("读取认证方法列表失败: %v\n", err)
		return false
	}

	// 检查客户端是否支持无认证 (0x00)
	for _, m := range methods {
		if m == authNoAuth {
			// 回复服务器选择无认证
			response := []byte{socks5Version, authNoAuth}
			if _, err := client.Write(response); err != nil {
				fmt.Printf("发送认证响应失败: %v\n", err)
				return false
			}
			return true
		}
	}

	// 没有共同的认证方法，回复拒绝
	response := []byte{socks5Version, 0xFF}
	client.Write(response)
	fmt.Println("客户端没有提供支持的认证方法")
	return false
}

// handleRequest 解析客户端请求并执行 CONNECT 转发
func handleRequest(client net.Conn) {
	// 读取请求头（固定 4 字节）
	header := make([]byte, 4)
	if _, err := io.ReadFull(client, header); err != nil {
		fmt.Printf("读取请求头失败: %v\n", err)
		return
	}

	ver, cmd, _, atyp := header[0], header[1], header[2], header[3]
	if ver != socks5Version {
		fmt.Printf("请求中的 SOCKS 版本错误: %d\n", ver)
		return
	}

	// 仅支持 CONNECT 命令
	if cmd != cmdConnect {
		sendReply(client, repCommandNotSupported)
		fmt.Printf("不支持的命令: 0x%02x\n", cmd)
		return
	}

	// 解析目标地址（仅支持 IPv4 和域名）
	targetAddr, err := parseAddress(client, atyp)
	if err != nil {
		fmt.Printf("解析地址失败: %v\n", err)
		sendReply(client, repAddressTypeNotSupported)
		return
	}

	// 连接目标服务器
	target, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		fmt.Printf("连接目标 %s 失败: %v\n", targetAddr, err)
		sendReply(client, repConnectionRefused)
		return
	}
	defer target.Close()

	// 构建成功响应，BND.ADDR 和 BND.PORT 使用代理连接目标时使用的本地地址
	localAddr := target.LocalAddr().(*net.TCPAddr)
	// 响应格式: VER, REP, RSV, ATYP, BND.ADDR, BND.PORT
	// 固定使用 IPv4 类型 (0x01)，长度 10 字节
	reply := make([]byte, 10)
	reply[0] = socks5Version
	reply[1] = repSuccess
	reply[2] = 0x00 // RSV
	reply[3] = atypIPv4
	copy(reply[4:8], localAddr.IP.To4())                            // 4 字节 IPv4 地址
	binary.BigEndian.PutUint16(reply[8:10], uint16(localAddr.Port)) // 2 字节端口

	if _, err := client.Write(reply); err != nil {
		fmt.Printf("发送成功响应失败: %v\n", err)
		return
	}

	// 双向数据转发，任意一方关闭连接即结束
	done := make(chan struct{}, 2)
	go func() {
		_, err := io.Copy(target, client)
		if err != nil && !isNormalConnectionClose(err) {
			fmt.Printf("客户端->目标转发异常错误: %v\n", err)
		}
		done <- struct{}{}
	}()
	go func() {
		_, err := io.Copy(client, target)
		if err != nil && !isNormalConnectionClose(err) {
			fmt.Printf("目标->客户端转发异常错误: %v\n", err)
		}
		done <- struct{}{}
	}()

	// 等待任意一个转发方向结束（通常是客户端或目标关闭连接）
	<-done
	// 另一个方向会因连接关闭而自动退出
}

// parseAddress 根据 ATYP 解析目标地址和端口
// 支持 IPv4 (0x01) 和 域名 (0x03)
// 返回 "host:port" 格式的字符串
func parseAddress(client net.Conn, atyp byte) (string, error) {
	switch atyp {
	case atypIPv4:
		// 读取 4 字节 IPv4 地址
		addr := make([]byte, 4)
		if _, err := io.ReadFull(client, addr); err != nil {
			return "", fmt.Errorf("读取 IPv4 地址失败: %w", err)
		}
		// 读取 2 字节端口
		port := make([]byte, 2)
		if _, err := io.ReadFull(client, port); err != nil {
			return "", fmt.Errorf("读取端口失败: %w", err)
		}
		ip := net.IP(addr).String()
		return fmt.Sprintf("%s:%d", ip, binary.BigEndian.Uint16(port)), nil

	case atypDomain:
		// 读取域名长度（1 字节）
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(client, lenByte); err != nil {
			return "", fmt.Errorf("读取域名长度失败: %w", err)
		}
		domainLen := int(lenByte[0])

		// 读取域名
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(client, domain); err != nil {
			return "", fmt.Errorf("读取域名失败: %w", err)
		}

		// 读取端口
		port := make([]byte, 2)
		if _, err := io.ReadFull(client, port); err != nil {
			return "", fmt.Errorf("读取端口失败: %w", err)
		}
		return fmt.Sprintf("%s:%d", string(domain), binary.BigEndian.Uint16(port)), nil

	default:
		// 不支持 IPv6 (0x04) 和其他类型
		return "", fmt.Errorf("不支持的地址类型: 0x%02x", atyp)
	}
}

// sendReply 发送 SOCKS5 响应（简化版，绑定地址固定为 0.0.0.0:0）
// 用于发送错误响应
func sendReply(client net.Conn, rep byte) {
	// 响应格式：VER, REP, RSV, ATYP, BND.ADDR, BND.PORT
	// 使用 IPv4 0.0.0.0:0 作为绑定地址（对于错误响应无意义）
	reply := []byte{socks5Version, rep, 0x00, atypIPv4, 0, 0, 0, 0, 0, 0}
	if _, err := client.Write(reply); err != nil {
		fmt.Printf("发送错误响应失败: %v\n", err)
	}
}

// isClosedConnError 判断错误是否为连接关闭的常见错误
// 用于在转发中忽略正常的关闭错误
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	// 网络错误，如 "use of closed network connection"
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}

// isNormalConnectionClose 判断错误是否属于连接正常关闭（可忽略）
// 包括：io.EOF, net.ErrClosed, 以及包含 "use of closed network connection" 的错误
func isNormalConnectionClose(err error) bool {
	if err == nil {
		return false
	}
	// 直接匹配常见错误
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return true
	}
	// 某些情况下错误可能被包装，需要递归检查
	// 同时检查错误字符串，因为旧版 Go 可能没有 net.ErrClosed
	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	// 尝试解包并递归检查
	if unwrappable, ok := err.(interface{ Unwrap() error }); ok {
		return isNormalConnectionClose(unwrappable.Unwrap())
	}
	return false
}
