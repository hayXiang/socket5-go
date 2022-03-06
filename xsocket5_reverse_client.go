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
		connect._conn = __conn
		if error != nil {
			log.Println(error)
			log.Println("sleep 1 sec to retry")
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		log.Printf("connected:%s\n", connect.ToString())
		connect.Write(XSOCKS_PROTOCAL_CLIENT_REGISTER)
		command_loop(connect, address)
	}
}

func create_proxy_channel(address string, protocal XsocktCreateProxyProtocal, process func(buffer []byte) []byte) {
	log.Printf("dest %s\n", protocal._body._address)
	connTimeout := 5 * time.Second
	client_chan := make(chan MyConnect)
	go func() {
		__conn, err := net.DialTimeout("tcp", address, connTimeout)
		connect := MyConnect{}
		connect._conn = __conn
		connect._uuid = string(protocal._body._uuid)
		connect._address = address
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("[cnn]connect: %s\n", connect.ToString())
			_, err := connect.Write(protocal.ToByte())
			if err != nil {
				log.Println(err)
			}
		}
		if err == nil {
			client_chan <- connect
		} else {
			client_chan <- MyConnect{}
		}
	}()

	server_chan := make(chan MyConnect)
	go func() {
		log.Printf("[begin]dial to %s", protocal._body._address)
		__conn, err := net.DialTimeout("tcp", string(protocal._body._address), connTimeout)
		connect := MyConnect{}
		connect._conn = __conn
		connect._uuid = string(protocal._body._uuid)
		connect._address = string(protocal._body._address)
		if err != nil {
			log.Println(err)
		}
		log.Printf("[end]dial to %s", protocal._body._address)
		if err == nil {
			server_chan <- connect
		} else {
			server_chan <- MyConnect{}
		}
	}()
	server := <-server_chan
	client := <-client_chan
	defer client.Close()
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

		for i := 0; i < size; i++ {
			data[len] = buffer[i]
			len++
		}

		if start, protocal_length, _ := parse_protocal(data); start != -1 && protocal_length != -1 {
			protocal := XsocktCreateProxyProtocal{}
			protocal.Parse(data[start : start+protocal_length])
			go create_proxy_channel(address, protocal, func(buffer []byte) []byte {
				record := Tls_Shake_Record{}
				if record.Parse(buffer) {
					record.Restore()
					return record.ToByte()
				}
				return nil
			})
			data = make([]byte, 10240)
			len = 0
			continue
		}
	}
}
