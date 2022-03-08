package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func process(local_address *string, getRemoteAddress func(client *MyConnect) (string, error), onConnectReady func(client *MyConnect), processClientData func([]byte) []byte) error {
	listen, _err := net.Listen("tcp", *local_address)
	if _err != nil {
		log.Println(_err)
		return _err
	}
	defer listen.Close()
	for {
		__client, _err := listen.Accept()
		if _err != nil {
			log.Println(_err)
			return _err
		}
		client := MyConnect{}
		client._conn = __client
		go func() {
			defer client.Close()
			log.Printf("receive a client request,%s\n", client.ToString())
			if getRemoteAddress == nil {
				log.Println("no onClientReady")
				return
			}
			address, _err := getRemoteAddress(&client)
			if _err != nil {
				log.Println(_err)
				return
			}
			process_connect(&address, &client, onConnectReady, processClientData)
		}()
	}
}

func process_connect(address *string, client *MyConnect, onConnectReady func(client *MyConnect), processClientData func([]byte) []byte) error {
	//send host and port to remote
	mutex.Lock()
	command_connect, ok := connect_map[XSOCKS_PROTOCAL_UUID_COMMAND]
	mutex.Unlock()
	if !ok {
		log.Println("no cmd connect")
		return errors.New("no cmd connect")
	}

	connect_count_mutex.Lock()
	connect_count++
	connect_count_mutex.Unlock()

	uuid := fmt.Sprintf("%d", connect_count)
	protocal := XsocktCreateProxyProtocal{}
	protocal._body = XsocktCreateProxyProtocalBoby{}
	protocal._body._uuid = []byte(uuid)
	protocal._body._address = []byte(*address)

	_, err := command_connect.Write(protocal.ToByte())
	if err != nil {
		log.Println(err)
		return err
	}
	//create remote connect session.
	is_connect_ok := false
	mutex.Lock()
	connect_ready_map[uuid] = make(chan bool)
	ready := connect_ready_map[uuid]
	mutex.Unlock()
	log.Printf("wait for connect ready, uuid = %s\n", uuid)
	go func() {
		time.Sleep(time.Second * time.Duration(3))
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
		return errors.New("wait for connect timeout")
	}
	mutex.Lock()
	connect := connect_map[uuid]
	connect._address = string(protocal._body._address)
	defer connect.Close()
	mutex.Unlock()
	if onConnectReady != nil {
		onConnectReady(client)
	}
	if connect._reserved_read_data != nil && len(connect._reserved_read_data) > 0 {
		client.Write(connect._reserved_read_data)
	}
	go func() {
		client.Forward(connect, processClientData)
		connect.Close()
	}()
	connect.Forward(client, nil)
	mutex.Lock()
	delete(connect_map, uuid)
	delete(connect_ready_map, uuid)
	mutex.Unlock()
	return nil
}
