package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func proxy_url(buf []byte) string {
	var host string
	switch buf[3] {
	case 0x01: //IP V4
		host = net.IP(buf[4:8]).String()
	case 0x03: //域名
		host = string(buf[5 : len(buf)-2]) //b[4]表示域名的长度
	case 0x04: //IP V6
		host = net.IP(buf[4:20]).String()
	}
	port := binary.BigEndian.Uint16(buf[len(buf)-2:])
	return fmt.Sprintf("%s:%d", host, port)
}
