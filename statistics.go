package workerman_statistics_go

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"
	"unicode/utf8"
)

const (
	// PACKAGE_FIXED_LENGTH 包头长度
	PACKAGE_FIXED_LENGTH = 17
	// MAX_UDP_PACKGE_SIZE udp包最大长度
	MAX_UDP_PACKGE_SIZE = 65507
	// MAX_CHAR_VALUE char类型能保存的最大数值
	MAX_CHAR_VALUE = 127
	// MAX_UNSIGNED_SHORT_VALUE usigned short 能保存的最大数值
	MAX_UNSIGNED_SHORT_VALUE = 65535
)

type WorkerManMsgInfo struct {
	Module    string  `json:"module"`
	InterFace string  `json:"inter_face"`
	CostTime  float32 `json:"cost_time"`
	Status    int     `json:"status"`
	Code      uint32  `json:"code"`
	TimeStamp uint32  `json:"time_stamp"`
	Msg       string  `json:"msg"`
	MsgLen    uint16  `json:"msg_len"`
}

type WorkerManClient struct {
	srcAddr *net.UDPAddr
	dstAddr *net.UDPAddr
}

func NewWorkerManClient(clientIP string, clientPort int, serverIP string, serverPort int) *WorkerManClient {
	client := new(WorkerManClient)
	srcIP := net.ParseIP(clientIP)
	destIP := net.ParseIP(serverIP)
	client.srcAddr = &net.UDPAddr{IP: srcIP, Port: clientPort}
	client.dstAddr = &net.UDPAddr{IP: destIP, Port: serverPort}
	return client
}

func (c *WorkerManClient) Decode(data []byte) (infos WorkerManMsgInfo) {
	if len(data) == 0 {
		return
	}
	packageHead := []byte(string(data[:PACKAGE_FIXED_LENGTH]))
	content := string(data[PACKAGE_FIXED_LENGTH:])
	moduleLen, _ := utf8.DecodeRuneInString(string(packageHead[:1]))
	interFaceLen, _ := utf8.DecodeRuneInString(string(packageHead[1:2]))
	binary.Read(bytes.NewBuffer(packageHead[2:6]), binary.LittleEndian, &infos.CostTime)
	status, _ := utf8.DecodeRuneInString(string(packageHead[6:7]))
	infos.Status = int(status)
	binary.Read(bytes.NewBuffer(packageHead[7:11]), binary.BigEndian, &infos.Code)
	binary.Read(bytes.NewBuffer(packageHead[11:13]), binary.BigEndian, &infos.MsgLen)
	binary.Read(bytes.NewBuffer(packageHead[13:17]), binary.BigEndian, &infos.TimeStamp)
	infos.Module = content[:moduleLen]
	infos.InterFace = content[moduleLen : moduleLen+interFaceLen]
	infos.Msg = content[moduleLen+interFaceLen : int(moduleLen)+int(interFaceLen)+int(infos.MsgLen)]
	return
}

// Encode
// a -- 将字符串空白以 NULL 字符填满
// A -- 将字符串空白以 SPACE 字符 (空格) 填满
// h -- 16进制字符串，低位在前以半字节为单位
// H -- 16进制字符串，高位在前以半字节为单位
// c -- 有符号字符
// C -- 无符号字符
// s -- 有符号短整数 (16位，主机字节序)
// S -- 无符号短整数 (16位，主机字节序)
// n -- 无符号短整数 (16位, 大端字节序)
// v -- 无符号短整数 (16位, 小端字节序)
// i -- 有符号整数 (依赖机器大小及字节序)
// I -- 无符号整数 (依赖机器大小及字节序)
// l -- 有符号长整数 (32位，主机字节序)
// L -- 无符号长整数 (32位，主机字节序)
// N -- 无符号长整数 (32位, 大端字节序)
// V -- 无符号长整数 (32位, 小端字节序)
// f -- 单精度浮点数 (依计算机的范围)
// d -- 双精度浮点数 (依计算机的范围)
// x -- 空字节
// X -- 倒回一位
// @ -- 填入 NULL 字符到绝对位置
// pack('CCfCNnN', $module_name_length, $interface_name_length, $cost_time, $success ? 1 : 0, $code, strlen($msg), time()).$module.$interface.$msg;
func (c *WorkerManClient) Encode(info WorkerManMsgInfo) ([]byte, error) {
	buf := new(bytes.Buffer)
	if info.Status < 0 || info.Status > 1 {
		return nil, errors.New("status should be 0 or 1")
	}
	if len(info.Module) > MAX_CHAR_VALUE {
		info.Module = info.Module[:MAX_CHAR_VALUE]
	}
	if len(info.InterFace) > MAX_CHAR_VALUE {
		info.InterFace = info.InterFace[:MAX_CHAR_VALUE]
	}
	moduleLen := len(info.Module)
	interFaceLen := len(info.InterFace)
	availableLen := MAX_UDP_PACKGE_SIZE - PACKAGE_FIXED_LENGTH - moduleLen - interFaceLen
	if len(info.Msg) > availableLen {
		info.Msg = info.Msg[:availableLen]
	}
	msgLen := len(info.Msg)
	binary.Write(buf, binary.LittleEndian, []byte(string(rune(moduleLen))))    // 1
	binary.Write(buf, binary.LittleEndian, []byte(string(rune(interFaceLen)))) // 1
	binary.Write(buf, binary.LittleEndian, info.CostTime)                      // 4
	binary.Write(buf, binary.LittleEndian, []byte(string(rune(info.Status))))  // 1
	binary.Write(buf, binary.BigEndian, info.Code)                             // 4
	binary.Write(buf, binary.BigEndian, uint16(msgLen))                        // 2
	binary.Write(buf, binary.BigEndian, uint32(time.Now().Unix()))             // 4
	buf.WriteString(info.Module)
	buf.WriteString(info.InterFace)
	buf.WriteString(info.Msg)
	return buf.Bytes(), nil
}

func (c *WorkerManClient) Send(info WorkerManMsgInfo) error {
	conn, err := net.DialUDP("udp", c.srcAddr, c.dstAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	data, err := c.Encode(info)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}
