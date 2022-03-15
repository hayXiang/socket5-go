package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func get_command_connenct() (*MyConnect, bool) {
	mutex.Lock()
	var result *MyConnect = nil
	for _, v := range cmd_connect_map {
		if result == nil || v._create_time > result._create_time {
			result = v
		}
	}
	mutex.Unlock()
	if result != nil {
		return result, true
	}
	return nil, false
}

func process(local_address *string, getRemoteAddress func(client *MyConnect) (string, error), onConnectReady func(client *MyConnect), processClientData func([]byte) []byte) error {
	listen, _err := net.Listen("tcp", *local_address)
	if _err != nil {
		log.Println(_err)
		return _err
	}
	defer listen.Close()
	for {
		__conn, _err := listen.Accept()
		if _err != nil {
			log.Println(_err)
			return _err
		}
		client := MyConnect{_conn: __conn, _create_time: time.Now().Unix()}
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
	command_connect, ok := get_command_connenct()
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
	connect_ready_map[uuid] = make(chan bool, 1) //NOTE: muset > 1,
	ready := connect_ready_map[uuid]
	log.Printf("wait for connect ready, uuid = %s\n", uuid)
	mutex.Unlock()
	go func() {
		time.Sleep(time.Second * time.Duration(3))
		mutex.Lock()
		//delete(connect_map, uuid)
		delete(connect_ready_map, uuid)
		if !is_connect_ok {
			ready <- false
		}
		mutex.Unlock()
	}()
	is_connect_ok = <-ready
	defer close(ready)
	if !is_connect_ok {
		log.Printf("wait for connect timeout,%s", uuid)
		return errors.New("wait for connect timeout")
	}
	mutex.Lock()
	connect := data_connect_map[uuid]
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
	return nil
}

func parseUserAndPassword(address *string) (string, string) {
	index_split := strings.Index(*address, ":")
	if index_split != -1 {
		return (*address)[0:index_split], (*address)[index_split+1:]
	}
	return "", ""
}

func parseHostAndPort(address *string) (string, string) {
	index_split := strings.Index(*address, ":")
	if index_split != -1 {
		if index_split == 0 {
			return "0.0.0.0", (*address)[1:]
		} else {
			return (*address)[0:index_split], (*address)[index_split+1:]
		}
	}
	return "127.0.0.1", *address
}
