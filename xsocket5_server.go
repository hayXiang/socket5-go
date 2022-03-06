package main

import (
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
		client._conn = __conn
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

			dest := sock5_remote_address(buffer[:len])
			log.Printf("dest %s\n", dest)
			__conn, err := net.Dial("tcp", dest)
			server := MyConnect{}
			server._conn = __conn
			if err != nil {
				log.Println(err)
				return
			}
			client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
			log.Println("start to transfer data")
			go func() {
				client.Forward(server, nil)
				server.Close()
			}()
			server.Forward(client, nil)
		}()
	}
}
