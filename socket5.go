package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func sock5_remote_address(buf []byte) string {
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

func sock5(address string, username string, password string) {
	listen, error := net.Listen("tcp", address)
	if error != nil {
		log.Println(error)
		return
	}
	defer listen.Close()
	for {
		__client, con_error := listen.Accept()
		if con_error != nil {
			log.Println(error)
			return
		}
		client := MyConnect{}
		client._conn = __client
		go func() {
			defer client.Close()
			log.Println("receive a client request")
			user_name_auth := false
			{
				buf := make([]byte, 1024)
				size, err := client.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}
				if buf[0] != 0x05 {
					log.Println("only support sock5")
					return
				}
				auth_methods := buf[2:size]
				log.Println("socket protocal")

				for i := 0; i < len(auth_methods); i++ {
					if auth_methods[i] == 0x02 {
						user_name_auth = true
						break
					}
				}

				if user_name_auth {
					client.Write([]byte{0x05, 0x2})
				} else {
					client.Write([]byte{0x05, 0x0})
				}
			}

			if user_name_auth {
				buf := make([]byte, 1024)
				_, err := client.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}

				if buf[0] != 0x01 {
					log.Println("socket user name auth protocal is not correct.")
					return
				}

				username_len := buf[1]
				auth_username := string(buf[2 : 2+username_len])
				password_len := buf[2+username_len]
				auth_password := string(buf[3+username_len : 3+username_len+password_len])

				if (auth_username == username && auth_password == password) || (username == "" && password == "") {
					client.Write([]byte{0x05, 0x0})
				} else {
					client.Write([]byte{0x05, 0xF})
				}
			}

			buf := make([]byte, 1024)
			size, err := client.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}

			if buf[0] != 0x05 {
				log.Println("only support socket protocal")
				return
			}
			target_address := sock5_remote_address(buf[0:size])
			log.Println("host:" + target_address)
			connect_count_mutex.Lock()
			connect_count++
			connect_count_mutex.Unlock()
			uuid := fmt.Sprintf("%d", connect_count)

			//send host and port to remote
			mutex.Lock()
			command_connect, ok := connect_map["cmd"]
			mutex.Unlock()
			if !ok {
				log.Println("no cmd connect")
				return
			}
			protocal := XsocktCreateProxyProtocal{}
			protocal._body = XsocktCreateProxyProtocalBoby{}
			protocal._body._uuid = []byte(uuid)
			protocal._body._address = []byte(target_address)
			process_connect(command_connect, client, protocal, func() {
				client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			}, func(buffer []byte) []byte {
				record := Tls_Shake_Record{}
				if record.Parse(buffer) {
					record.Modify("emby.haycker.com")
					return record.ToByte()
				}
				return nil
			})
		}()
	}
}
