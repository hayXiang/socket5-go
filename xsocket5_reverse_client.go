package main

import (
	"log"
	"net"
	"time"
)

func start_reverse_xsocket5_client(address string) {
	for {
		__conn, error := net.Dial("tcp", address)
		connect := MyConnect{}
		connect.conn = __conn
		if error != nil {
			log.Println(error)
			log.Println("sleep 1 sec to retry")
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		log.Printf("connected:%s\n", connect.ToString())
		connect.Write(FLAG_CONNECT_COMMAND)
		command_loop(connect, address)
	}
}

func create_proxy_channel(proxy_address string, uuid []byte, address string, flag []byte, process func(buffer []byte) []byte) {
	log.Printf("dest %s\n", proxy_address)
	connTimeout := 5 * time.Second
	__conn, err := net.DialTimeout("tcp", address, connTimeout)
	client := MyConnect{}
	client.conn = __conn
	client.uuid = string(uuid)
	client.address = address
	defer client.Close()
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("[cnn]connect: %s\n", client.ToString())
		client.Write(uuid)
		_, err := client.Write(flag)
		if err != nil {
			log.Println(err)
		}
	}
	server_chan := make(chan MyConnect)
	go func() {
		log.Printf("[begin]dial to %s", proxy_address)
		__conn, err := net.DialTimeout("tcp", proxy_address, connTimeout)
		connect := MyConnect{}
		connect.conn = __conn
		connect.uuid = string(uuid)
		connect.address = proxy_address
		if err != nil {
			log.Println(err)
		}
		log.Printf("[end]dial to %s", proxy_address)
		if err == nil {
			server_chan <- connect
		} else {
			server_chan <- MyConnect{}
		}
	}()
	server := <-server_chan
	go func() {
		client.Forward(server, process)
		server.Close()
	}()
	server.Forward(client, nil)

}

func command_loop(connect MyConnect, address string) {
	defer connect.Close()
	len := 0
	data := make([]byte, 10240)
	for {
		buffer := make([]byte, 1024)
		size, err := connect.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(buffer[0:size])
		for i := 0; i < size; i++ {
			data[len] = buffer[i]
			len++
		}

		if data[len-2] == FLAG_CONNECT_PROXY[0] && data[len-1] == FLAG_CONNECT_PROXY[1] {
			uuid_len := int(data[len-3])
			uuid := data[len-3-uuid_len : len-3]
			info_len := int(data[len-3-uuid_len-1])
			index := len - 3 - uuid_len - 1 - info_len
			info := data[index : index+info_len]

			go create_proxy_channel(proxy_url(info), uuid, address, FLAG_CONNECT_PROXY, func(buffer []byte) []byte {
				record := Tls_Shake_Record{}
				if record.Parse(buffer) {
					record.Restore()
					return record.toByte()
				}
				return nil
			})
			data = make([]byte, 10240)
			len = 0
			continue
		}

		if data[len-2] == FLAG_CONNECT_PORT_FORWRD[0] && data[len-1] == FLAG_CONNECT_PORT_FORWRD[1] {
			uuid_len := int(data[len-3])
			uuid := data[len-3-uuid_len : len-3]
			info_len := int(data[len-3-uuid_len-1])
			index := len - 3 - uuid_len - 1 - info_len
			info := data[index : index+info_len]
			go create_proxy_channel(string(info), uuid, string(address), FLAG_CONNECT_PORT_FORWRD, nil)
			data = make([]byte, 10240)
			len = 0
			continue
		}
	}
}
