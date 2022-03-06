package main

import (
	"fmt"
	"log"
	"net"
)

func port_forward(local_address string, remote_address string) {
	listen, error := net.Listen("tcp", local_address)
	if error != nil {
		log.Println(error)
		return
	}
	defer listen.Close()
	for {
		__client, con_error := listen.Accept()
		if con_error != nil {
			log.Println(error)
			continue
		}
		client := MyConnect{}
		client._conn = __client
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
			uuid := fmt.Sprintf("%d", client._count)
			protocal := XsocktCreateProxyProtocal{}
			protocal._body = XsocktCreateProxyProtocalBoby{}
			protocal._body._uuid = []byte(uuid)
			protocal._body._address = []byte(remote_address)
			process_connect(command_connect, client, protocal, nil, nil)
		}()
	}
}
