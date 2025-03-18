package smtp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"mailfinger/utils"
	"net"
	"os"
	"strings"
	"time"
)

// DoQuery
// @Description: 单target查询
// @param target
func DoQuery(target string) {
	var err error

	// 格式化目标
	target = fmt.Sprintf("%s:25", target)

	// 创建目标文件夹
	folderName := fmt.Sprintf("./data/%s/smtp", strings.ReplaceAll(strings.Split(target, ":")[0], ".", "_"))
	err = os.MkdirAll(folderName, os.ModePerm)

	// 建立套接字
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		fmt.Printf("%s 连接失败: %v\n", target, err)
		return
	}

	defer conn.Close()

	// 读取欢迎消息
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("%s 读取欢迎消息失败: %v\n", target, err)
		return
	}

	// welcome数据包
	helloMsg := string(buffer[:n])
	utils.SaveToFile(fmt.Sprintf("%s/welcome_message.txt", folderName), helloMsg)

	// 发送 EHLO 命令
	_, err = conn.Write([]byte("EHLO mx.baidu.com\r\n"))
	if err != nil {
		fmt.Printf("发送 EHLO 失败: %v\n", err)
		return
	}

	// 读取 EHLO 响应
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Printf("读取 EHLO 响应失败: %v\n", err)
		return
	}

	// 保存 EHLO 响应到文件
	ehloResponse := string(buffer[:n])
	utils.SaveToFile(fmt.Sprintf("%s/ehlo_response.txt", folderName), ehloResponse)

	// 发送 STARTTLS 命令
	_, err = conn.Write([]byte("STARTTLS\r\n"))
	if err != nil {
		fmt.Printf("发送 STARTTLS 失败: %v\n", err)
		return
	}

	// 读取响应以确认服务器是否支持 STARTTLS
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Printf("读取 STARTTLS 响应失败: %v\n", err)
		return
	}

	// 保存 STARTTLS 响应到文件
	tlsResponse := string(buffer[:n])
	utils.SaveToFile(fmt.Sprintf("%s/starttls_response.txt", folderName), tlsResponse)

	// 查看是否支持 STARTTLS
	if !strings.HasPrefix(tlsResponse, "220") {
		fmt.Printf("%s 服务器未响应支持 STARTTLS\n", target)
		return
	}

	// 创建自定义 TLS 配置（跳过证书验证）
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,        // 仅用于测试环境，请勿在生产环境中使用
		ServerName:         "baidu.com", // 应与目标邮件服务器的SSL证书中的CN匹配
	}

	// 升级到 TLS 连接
	tlsConn := tls.Client(conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		fmt.Printf("TLS 握手失败: %v\n", err)
		return
	}
	defer tlsConn.Close()

	// 获取证书链并打印证书信息
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		fmt.Println("未获取到证书")
		return
	}

	fmt.Printf("%s 获取到证书数量:%d\n", target, len(certs))

	// 假设我们已经获取了证书链 certs
	for i, cert := range certs {
		// 创建或打开文件
		filename := fmt.Sprintf("%s/cert_%d.txt", folderName, i+1)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("无法创建或打开文件: %v\n", err)
			continue
		}
		defer file.Close()

		// 创建 writer 并写入证书详情
		writer := bufio.NewWriter(file)
		writer.WriteString(fmt.Sprintf("========= 证书 %d =========\n", i+1))
		writer.WriteString(fmt.Sprintf("主题: %v\n", cert.Subject))
		writer.WriteString(fmt.Sprintf("颁发者: %v\n", cert.Issuer))
		writer.WriteString(fmt.Sprintf("有效期:\n 开始: %v\n 结束: %v\n", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02")))
		writer.WriteString(fmt.Sprintf("序列号: %v\n", cert.SerialNumber))
		writer.WriteString(fmt.Sprintf("签名算法: %v\n", cert.SignatureAlgorithm))
		writer.WriteString(fmt.Sprintf("版本: %v\n", cert.Version))
		writer.Flush() // 别忘了刷新缓冲区
	}

}
