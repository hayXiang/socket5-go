package main

import (
	"log"
	"net"
	"time"
)

func start_xsocket5_server(address *string) {
	log.Println("listen on " + *address)
	listen, error := net.Listen("tcp", *address)
	if error != nil {
		log.Println(error)
		return
	}

	for {
		__conn, con_error := listen.Accept()
		client := MyConnect{_conn: __conn, _create_time: time.Now().Unix()}
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

			dest := sock5_destination_address(buffer[:len])
			log.Printf("dest %s\n", dest)
			__conn, err := net.Dial("tcp", dest)
			server := MyConnect{_conn: __conn, _create_time: time.Now().Unix()}
			if err != nil {
				log.Println(err)
				return
			}
			client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
			log.Println("start to transfer data")
			go func() {
				client.Forward(&server, nil)
				server.Close()
			}()
			server.Forward(&client, nil)
		}()
	}
}
