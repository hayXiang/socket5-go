package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func startt_xsocket5_server(address string) {
	log.Println("listen on " + address)
	listen, error := net.Listen("tcp", address)
	if error != nil {
		log.Println(error)
		return
	}

	for {
		__conn, con_error := listen.Accept()
		client := MyConnect{}
		client.conn = __conn
		if con_error != nil {
			log.Println("error")
			return
		}
		go func() {
			defer client.Close()
			log.Println("##############################")
			log.Println("receive a client request")
			buffer := make([]byte, 128, 1024)
			_, err := client.Read(buffer)
			if err != nil {
				log.Println("read data failed")
				return
			}
			if buffer[0] != 0x05 {
				log.Println("only support socket5")
				return
			}
			log.Println("socket5 proxy protocal")
			client.Write([]byte{0x05, 0x00})
			len, err := client.Read(buffer)
			if err != nil {
				log.Println(err)
				return
			}
			b := buffer
			var host string
			switch b[3] {
			case 0x01: //IP V4
				host = net.IP(b[4:8]).String()
			case 0x03: //域名
				host = string(b[5 : len-2]) //b[4]表示域名的长度
			case 0x04: //IP V6
				host = net.IP(b[4:20]).String()
			}
			port := fmt.Sprintf("%d", binary.BigEndian.Uint16(b[len-2:len]))
			log.Printf("dest %s:%s\n", host, port)
			__conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
			server := MyConnect{}
			server.conn = __conn
			if err != nil {
				log.Println(err)
				return
			}
			client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
			log.Println("start to transfer data")
			go func() {
				client.Forward(server, func(buffer []byte) []byte {
					record := Tls_Shake_Record{}
					if record.Parse(buffer) {
						record.Restore()
						return record.toByte()
					}
					return nil
				})
				server.Close()
			}()
			server.Forward(client, nil)
		}()
	}
}
