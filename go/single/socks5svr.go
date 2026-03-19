package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	// 监听所有接口的1080端口（SOCKS5标准端口）
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		fmt.Printf("监听失败: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("SOCKS5代理服务器启动，监听 :1080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}

// handleClient 处理单个SOCKS5客户端连接
func handleClient(client net.Conn) {
	defer client.Close()

	// 1. 认证协商阶段（RFC 1928 Section 3）
	if !handleAuth(client) {
		return
	}

	// 2. 处理请求（RFC 1928 Section 4）
	handleRequest(client)
}

// handleAuth 处理认证协商
func handleAuth(client net.Conn) bool {
	// 读取客户端认证方法请求
	// 格式: [VER(1), NMETHODS(1), METHODS(1-255)]
	buf := make([]byte, 2)
	_, err := io.ReadFull(client, buf)
	if err != nil {
		return false
	}

	ver, nmethods := buf[0], buf[1]
	if ver != 0x05 {
		fmt.Printf("不支持的SOCKS版本: %d\n", ver)
		return false
	}

	// 读取客户端支持的所有认证方法
	methods := make([]byte, nmethods)
	_, err = io.ReadFull(client, methods)
	if err != nil {
		return false
	}

	// 检查是否支持无认证方法 (0x00)
	supportNoAuth := false
	for _, method := range methods {
		if method == 0x00 {
			supportNoAuth = true
			break
		}
	}

	if !supportNoAuth {
		// 回复客户端不支持任何认证方法 (0xFF)
		client.Write([]byte{0x05, 0xFF})
		return false
	}

	// 选择无认证方法 (0x00)
	client.Write([]byte{0x05, 0x00})
	return true
}

// handleRequest 处理客户端请求
func handleRequest(client net.Conn) {
	// 读取请求头
	// 格式: [VER(1), CMD(1), RSV(1), ATYP(1)]
	header := make([]byte, 4)
	_, err := io.ReadFull(client, header)
	if err != nil {
		return
	}

	ver, cmd, _, atyp := header[0], header[1], header[2], header[3]
	if ver != 0x05 {
		return
	}

	// 根据ATYP解析目标地址
	targetAddr, err := parseAddress(client, atyp)
	if err != nil {
		sendReply(client, 0x08) // ADDR_TYPE_NOT_SUPPORTED
		return
	}

	// 根据命令类型处理
	switch cmd {
	case 0x01: // CONNECT (TCP连接)
		log.Println("-----使用TCP连接---CONNECT")
		handleConnect(client, targetAddr)
	case 0x03: // UDP ASSOCIATE
		log.Println("-----使用UDP连接----ASSOCIATE")
		handleUDPAssociate(client, targetAddr)
	default:
		// 不支持的命令
		sendReply(client, 0x07) // CMD_NOT_SUPPORTED
	}
}

// parseAddress 根据ATYP解析目标地址 (RFC 1928 Section 5)
func parseAddress(client net.Conn, atyp byte) (string, error) {
	switch atyp {
	case 0x01: // IPv4地址
		addr := make([]byte, 4)
		_, err := io.ReadFull(client, addr)
		if err != nil {
			return "", err
		}
		// 读取端口号 (2字节，大端序)
		port := make([]byte, 2)
		_, err = io.ReadFull(client, port)
		if err != nil {
			return "", err
		}
		portNum := binary.BigEndian.Uint16(port)
		return fmt.Sprintf("%s:%d", net.IP(addr).String(), portNum), nil

	case 0x03: // 域名
		// 先读取域名长度
		lenByte := make([]byte, 1)
		_, err := io.ReadFull(client, lenByte)
		if err != nil {
			return "", err
		}
		domainLen := int(lenByte[0])

		// 读取域名
		domain := make([]byte, domainLen)
		_, err = io.ReadFull(client, domain)
		if err != nil {
			return "", err
		}

		// 读取端口号
		port := make([]byte, 2)
		_, err = io.ReadFull(client, port)
		if err != nil {
			return "", err
		}
		portNum := binary.BigEndian.Uint16(port)
		return fmt.Sprintf("%s:%d", string(domain), portNum), nil

	default:
		// 不支持的地址类型
		// 需要丢弃剩余的数据
		if atyp == 0x04 { // IPv6
			// 跳过IPv6地址 (16字节) + 端口 (2字节)
			discardBuf := make([]byte, 18)
			client.Read(discardBuf)
		}
		return "", fmt.Errorf("不支持的地址类型: 0x%02x", atyp)
	}
}

// sendReply 发送SOCKS5响应 (RFC 1928 Section 6)
func sendReply(client net.Conn, rep byte) {
	// 响应格式: [VER(1), REP(1), RSV(1), ATYP(1), BND.ADDR, BND.PORT]
	// 这里简化处理，使用0.0.0.0:0作为绑定地址
	reply := []byte{0x05, rep, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	client.Write(reply)
}

// handleConnect 处理CONNECT命令
func handleConnect(client net.Conn, targetAddr string) {
	// 尝试连接目标服务器
	target, err := net.Dial("tcp", targetAddr)
	if err != nil {
		fmt.Printf("连接目标服务器失败 %s: %v\n", targetAddr, err)
		sendReply(client, 0x05) // CONNECTION_REFUSED
		return
	}
	defer target.Close()

	// 发送成功响应
	// 获取本地连接地址作为BND.ADDR
	localAddr := target.LocalAddr().(*net.TCPAddr)
	reply := make([]byte, 10)
	reply[0] = 0x05 // VER
	reply[1] = 0x00 // REP (成功)
	reply[2] = 0x00 // RSV
	reply[3] = 0x01 // ATYP (IPv4)
	copy(reply[4:8], localAddr.IP.To4())
	binary.BigEndian.PutUint16(reply[8:10], uint16(localAddr.Port))
	client.Write(reply)

	// 双向数据转发
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(target, client)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(client, target)
		done <- struct{}{}
	}()

	// 等待任意一方关闭连接
	<-done
}

// handleUDPAssociate 处理UDP ASSOCIATE命令 (RFC 1928 Section 7)
func handleUDPAssociate(client net.Conn, targetAddr string) {
	// UDP ASSOCIATE命令中，客户端发送的目标地址通常为0.0.0.0:0
	// 但我们需要返回代理服务器监听的UDP地址和端口

	// 创建UDP Socket
	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		sendReply(client, 0x01) // GENERAL_FAILURE
		return
	}
	defer udpConn.Close()

	// 获取UDP监听的地址
	udpAddr := udpConn.LocalAddr().(*net.UDPAddr)

	// 发送响应，包含UDP监听的地址和端口
	reply := make([]byte, 10)
	reply[0] = 0x05 // VER
	reply[1] = 0x00 // REP (成功)
	reply[2] = 0x00 // RSV
	reply[3] = 0x01 // ATYP (IPv4)
	copy(reply[4:8], udpAddr.IP.To4())
	binary.BigEndian.PutUint16(reply[8:10], uint16(udpAddr.Port))
	client.Write(reply)

	fmt.Printf("UDP关联已建立，监听UDP: %s\n", udpAddr.String())

	// 处理UDP数据转发
	handleUDPForwarding(udpConn, client, targetAddr)
}

// handleUDPForwarding 处理UDP数据转发
func handleUDPForwarding(udpConn *net.UDPConn, tcpClient net.Conn, targetAddr string) {
	// 存储客户端地址映射（简化实现，只处理单个客户端）
	var clientAddr *net.UDPAddr

	// 处理从客户端UDP连接接收的数据
	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			// 记录客户端地址
			clientAddr = addr

			// 解析UDP数据包 (RFC 1928 Section 7)
			// 格式: [RSV(2), FRAG(1), ATYP(1), DST.ADDR, DST.PORT, DATA...]
			if n < 4 { // 至少需要RSV(2)+FRAG(1)+ATYP(1)
				continue
			}

			rsv1, rsv2, frag, atyp := buf[0], buf[1], buf[2], buf[3]
			if rsv1 != 0 || rsv2 != 0 {
				continue
			}

			// 这里简化处理，只处理frag=0的数据包
			if frag != 0 {
				continue
			}

			// 解析目标地址和端口
			offset := 4
			var destAddr string

			switch atyp {
			case 0x01: // IPv4
				if n < offset+4+2 {
					continue
				}
				ip := net.IP(buf[offset : offset+4])
				offset += 4
				port := binary.BigEndian.Uint16(buf[offset : offset+2])
				offset += 2
				destAddr = fmt.Sprintf("%s:%d", ip.String(), port)

			case 0x03: // 域名
				if n < offset+1 {
					continue
				}
				domainLen := int(buf[offset])
				offset++
				if n < offset+domainLen+2 {
					continue
				}
				domain := string(buf[offset : offset+domainLen])
				offset += domainLen
				port := binary.BigEndian.Uint16(buf[offset : offset+2])
				offset += 2
				destAddr = fmt.Sprintf("%s:%d", domain, port)

			default:
				// 不支持的地址类型
				continue
			}

			// 剩余的数据
			data := buf[offset:n]

			// 转发到目标服务器
			targetUDPAddr, err := net.ResolveUDPAddr("udp", destAddr)
			if err != nil {
				continue
			}

			udpConn.WriteToUDP(data, targetUDPAddr)
		}
	}()

	// 处理从目标服务器返回的数据
	go func() {
		buf := make([]byte, 65535)
		for {
			if clientAddr == nil {
				continue
			}

			// 这里简化实现：假设我们知道目标地址
			// 实际应该为每个目标地址维护一个映射
			targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddr)
			if err != nil {
				continue
			}

			// 创建一个UDP客户端连接到目标服务器
			targetConn, err := net.DialUDP("udp", nil, targetUDPAddr)
			if err != nil {
				continue
			}
			defer targetConn.Close()

			// 监听返回的数据
			n, err := targetConn.Read(buf)
			if err != nil {
				continue
			}

			// 将数据封装为SOCKS5 UDP数据包格式
			packet := make([]byte, n+10) // 10字节头部
			// RSV = 0x0000
			packet[0] = 0x00
			packet[1] = 0x00
			// FRAG = 0x00
			packet[2] = 0x00
			// ATYP = 0x01 (IPv4)
			packet[3] = 0x01
			// 目标地址 (这里简化使用原始目标地址)
			copy(packet[4:8], targetUDPAddr.IP.To4())
			binary.BigEndian.PutUint16(packet[8:10], uint16(targetUDPAddr.Port))
			// 数据
			copy(packet[10:], buf[:n])

			// 发送回客户端
			if clientAddr != nil {
				udpConn.WriteToUDP(packet, clientAddr)
			}
		}
	}()

	// 保持TCP连接活跃（UDP ASSOCIATE需要保持TCP控制连接）
	// 这里简化处理，等待TCP连接关闭
	buf := make([]byte, 1)
	for {
		_, err := tcpClient.Read(buf)
		if err != nil {
			break
		}
	}
}
