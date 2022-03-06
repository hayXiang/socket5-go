package main

import (
	"log"
	"time"
)

type MyError struct {
	msg string
}

func (err *MyError) Error() string {
	return err.msg
}

func process_connect(command_connect MyConnect, client MyConnect, protocal XsocktCreateProxyProtocal, onConnectReady func(), processClientData func([]byte) []byte) error {
	uuid := string(protocal._body._uuid)
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
		return &MyError{msg: "wait for connect timeout"}
	}
	mutex.Lock()
	connect := connect_map[uuid]
	connect._address = string(protocal._body._address)
	defer connect.Close()
	mutex.Unlock()
	if onConnectReady != nil {
		onConnectReady()
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
