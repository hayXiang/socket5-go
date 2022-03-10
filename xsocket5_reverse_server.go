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
var connect_map = make(map[string]*MyConnect, 1024)
var connect_ready_map = make(map[string]chan bool, 1024)

func wait_for_connect(conn *MyConnect) {
	log.Printf("[begin]wait for connect, %s\n", conn.ToString())
	defer func() {
		log.Printf("[end]wait for connect, %s\n", conn.ToString())
	}()
	size := 0
	data := make([]byte, 10240)
	for {
		buffer := make([]byte, 1024)
		buf_size, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		for i := 0; i < buf_size; i++ {
			data[size] = buffer[i]
			size++
		}
		begin, protocal_length, protocal_type := parse_protocal(data)
		if begin == -1 || protocal_length == -1 {
			log.Println(data[0:size])
			continue
		}

		if protocal_type == int(XSOCKS_PROTOCAL_TYPE_REGIST) {
			mutex.Lock()
			connect_map[XSOCKS_PROTOCAL_UUID_COMMAND] = conn
			mutex.Unlock()
			go func() {
				for {
					buffer := make([]byte, 1024)
					_, err := conn.Read(buffer)
					if err != nil {
						log.Printf("cmd connect close,err:%s\n", err)
						mutex.Lock()
						delete(connect_map, XSOCKS_PROTOCAL_UUID_COMMAND)
						mutex.Unlock()
						return
					}
				}
			}()
		} else {
			protocal := XsocktCreateProxyProtocal{}
			protocal.Parse(data[begin : begin+protocal_length])
			conn._uuid = string(protocal._body._uuid)
			log.Printf("connect, %s ", conn.ToString())
			mutex.Lock()
			ready := connect_ready_map[conn._uuid]
			if ready != nil {
				connect_map[conn._uuid] = conn
				if protocal_length != size {
					conn._reserved_read_data = data[begin+protocal_length : size]
				}
				ready <- true
			} else {
				conn.Close()
			}
			mutex.Unlock()
		}
		return
	}
}

func start_reverse_xsocket5_server(address *string) {
	host, port := parseHostAndPort(address)
	bind_address := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[server] listen on:%s\n", bind_address)
	listen, err := net.Listen("tcp", bind_address)
	if err != nil {
		log.Println(err)
		return
	}
	defer listen.Close()
	for {
		__conn, err := listen.Accept()
		conn := MyConnect{_conn: __conn, _create_time: time.Now().Unix()}
		if err != nil {
			log.Println(err)
			return
		}
		go wait_for_connect(&conn)
	}
}
