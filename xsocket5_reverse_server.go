package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var connect_count_mutex sync.Mutex
var connect_count uint32 = 0
var mutex sync.Mutex
var connect_map = make(map[string]MyConnect, 1024)
var connect_ready_map = make(map[string]chan bool, 1024)

func port_forward(local_address string, remote_address string) {
	listen, error := net.Listen("tcp", local_address)
	if error != nil {
		log.Println(error)
	}
	defer listen.Close()
	for {
		__client, con_error := listen.Accept()
		if con_error != nil {
			log.Println(error)
		}
		client := MyConnect{}
		client.conn = __client
		go func() {
			defer client.Close()
			log.Println("receive a client request")
			//send host and port to remote
			mutex.Lock()
			command_connect, ok := connect_map["cmd"]
			mutex.Unlock()
			if !ok {
				log.Println("no cmd connect")
				return
			}

			connect_count_mutex.Lock()
			connect_count++
			connect_count_mutex.Unlock()
			buf := []byte(remote_address)
			size := len(buf)
			uuid := fmt.Sprintf("%d", client.count)
			uuid_byte := []byte(uuid)
			uuid_len := len(uuid_byte)
			info := make([]byte, size+1+uuid_len+1+2)
			for i := 0; i < size; i++ {
				info[i] = buf[i]
			}
			info[size] = byte(size)
			for i := 0; i < len(uuid_byte); i++ {
				info[size+1+i] = uuid_byte[i]
			}
			info[size+1+uuid_len] = byte(uuid_len)
			info[len(info)-2] = FLAG_CONNECT_PORT_FORWRD[0]
			info[len(info)-1] = FLAG_CONNECT_PORT_FORWRD[1]

			size, err := command_connect.Write(info)
			if err != nil || size < 0 {
				log.Println(err)
				return
			}
			//create remote connect session.
			is_connect_ok := false
			mutex.Lock()
			connect_ready_map[uuid] = make(chan bool)
			ready := connect_ready_map[uuid]
			mutex.Unlock()
			log.Printf("wait for connect ready, uuid = %s\n", uuid)
			go func() {
				time.Sleep(time.Second * time.Duration(100))
				if !is_connect_ok {
					mutex.Lock()
					ready <- false
					mutex.Unlock()
					return
				}
			}()

			is_connect_ok = <-ready
			if !is_connect_ok {
				log.Printf("wait for connect timeout,%s", uuid)
				return
			}
			mutex.Lock()
			connect := connect_map[uuid]
			defer connect.Close()
			mutex.Unlock()

			go func() {
				client.Forward(connect, nil)
				connect.Close()
			}()
			connect.Forward(client, nil)
			mutex.Lock()
			delete(connect_map, uuid)
			delete(connect_ready_map, uuid)
			mutex.Unlock()
		}()
	}
}

func sock5(address string, username string, password string) {
	listen, error := net.Listen("tcp", address)
	if error != nil {
		log.Println(error)
	}
	defer listen.Close()
	for {
		__client, con_error := listen.Accept()
		if con_error != nil {
			log.Println(error)
		}
		client := MyConnect{}
		client.conn = __client
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

			var host string
			switch buf[3] {
			case 0x01: //IP V4
				host = net.IP(buf[4:8]).String()
			case 0x03: //域名
				host = string(buf[5 : size-2]) //b[4]表示域名的长度
			case 0x04: //IP V6
				host = net.IP(buf[4:20]).String()
			}
			log.Println("host:" + host)
			connect_count_mutex.Lock()
			connect_count++
			connect_count_mutex.Unlock()
			uuid := fmt.Sprintf("%d", connect_count)
			uuid_byte := []byte(uuid)
			uuid_len := len(uuid_byte)
			info := make([]byte, size+1+uuid_len+1+2)
			for i := 0; i < size; i++ {
				info[i] = buf[i]
			}
			info[size] = byte(size)
			for i := 0; i < len(uuid_byte); i++ {
				info[size+1+i] = uuid_byte[i]
			}
			info[size+1+uuid_len] = byte(uuid_len)
			info[len(info)-2] = FLAG_CONNECT_PROXY[0]
			info[len(info)-1] = FLAG_CONNECT_PROXY[1]

			//send host and port to remote
			mutex.Lock()
			command_connect, ok := connect_map["cmd"]
			mutex.Unlock()
			if !ok {
				log.Println("no cmd connect")
				return
			}

			size, err = command_connect.Write(info)
			if err != nil || size < 0 {
				log.Println(err)
				return
			}
			//create remote connect session.
			state := 0
			mutex.Lock()
			connect_ready_map[uuid] = make(chan bool)
			ready := connect_ready_map[uuid]
			mutex.Unlock()
			log.Printf("wait for connect ready, uuid = %s\n", uuid)
			go func() {
				time.Sleep(time.Second * time.Duration(3))
				if state == 0 {
					mutex.Lock()
					ready <- false
					mutex.Unlock()
					return
				}
			}()

			ok = <-ready
			if !ok {
				log.Printf("wait for connect timeout,%s", uuid)
				return
			}
			state = 1
			mutex.Lock()
			connect := connect_map[uuid]
			connect.address = host
			defer connect.Close()
			mutex.Unlock()
			client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			go func() {
				client.Forward(connect, func(buffer []byte) []byte {
					record := Tls_Shake_Record{}
					if record.Parse(buffer) {
						record.Modify("emby.haycker.com")
						return record.toByte()
					}
					return nil
				})
				connect.Close()
			}()
			connect.Forward(client, nil)
			mutex.Lock()
			delete(connect_map, uuid)
			delete(connect_ready_map, uuid)
			mutex.Unlock()
			//go wait_for_connect(connect)
		}()
	}
}
func wait_for_connect(conn MyConnect) {
	log.Printf("[begin]wait for connect, %s\n", conn.ToString())
	defer func() {
		log.Printf("[end]wait for connect, %s\n", conn.ToString())
	}()
	len := 0
	data := make([]byte, 10240)
	for {
		buffer := make([]byte, 1024)
		buf_size, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		for i := 0; i < buf_size; i++ {
			data[len] = buffer[i]
			len++
		}
		//log.Println(buffer[0:buf_size])
		if data[len-2] == FLAG_CONNECT_COMMAND[0] && data[len-1] == FLAG_CONNECT_COMMAND[1] {
			mutex.Lock()
			connect_map["cmd"] = conn
			mutex.Unlock()
			go func() {
				for {
					buffer := make([]byte, 1024)
					_, err := conn.Read(buffer)
					if err != nil {
						log.Printf("cmd connect close,err:%s\n", err.Error())
						mutex.Lock()
						delete(connect_map, "cmd")
						mutex.Unlock()
						return
					}
				}
			}()
			return
		} else if (len > 1 && data[len-2] == FLAG_CONNECT_PORT_FORWRD[0] && data[len-1] == FLAG_CONNECT_PORT_FORWRD[1]) ||
			(len > 1 && data[len-2] == FLAG_CONNECT_PROXY[0] && data[len-1] == FLAG_CONNECT_PROXY[1]) {
			uuid := string(string(data[0 : len-2]))
			conn.uuid = uuid
			log.Printf("connect, %s ", conn.ToString())
			mutex.Lock()
			connect_map[uuid] = conn
			ready := connect_ready_map[uuid]
			if ready != nil {
				ready <- true
			}
			mutex.Unlock()
			return
		} else {
			log.Println(data[0:len])
		}

	}
}

func start_reverse_xsocket5_server(sock5_address string, username string, password string, address string) {
	local_address := ":9999"
	remote_address := ":22"
	go sock5(sock5_address, username, password)
	go port_forward(local_address, remote_address)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("listen on %s, socket5 on %s, port_forward on [%s->%s]", address, sock5_address, local_address, remote_address)
	defer listen.Close()
	for {
		__conn, err := listen.Accept()
		conn := MyConnect{}
		conn.conn = __conn
		if err != nil {
			log.Println(err)
			return
		}
		go wait_for_connect(conn)
	}
}
