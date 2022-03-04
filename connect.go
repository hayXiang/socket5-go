package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type Interface_MyConnect interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	ToString() string
	Forward(connect MyConnect, process func([]byte) []byte) bool
}

type MyConnect struct {
	conn      net.Conn
	count     int
	is_closed bool
	uuid      string
	address   string
}

func (_self_ *MyConnect) Forward(destination MyConnect, process func([]byte) []byte) {
	if _self_.conn == nil || destination.conn == nil {
		return
	}
	log.Printf("[start] transfer data,%s\n", _self_.ToString())
	defer log.Printf("[end] transfer data, %s, count = %d\n", _self_.ToString(), _self_.count)
	_self_.count = 0
	for {
		buf := make([]byte, 1024)
		size, err := _self_.conn.Read(buf)
		if err != nil {
			log.Printf("read over,%s,error=%s\n", _self_.ToString(), err.Error())
			_self_.Close()
			break
		}
		_self_.count += size
		if size == 4 && binary.BigEndian.Uint32(buf[0:size]) == binary.BigEndian.Uint32(FLAG_QUIT) {
			log.Printf("read force quit,%s", _self_.ToString())
			break
		}

		if process != nil {
			data := process(buf[0:size])
			if data != nil {
				destination.Write(data)
			} else {
				destination.Write(buf[0:size])
			}
		} else {
			destination.Write(buf[0:size])
		}
	}
}

func (_self_ *MyConnect) Read(buf []byte) (int, error) {
	if _self_.conn == nil {
		return -1, nil
	}
	return _self_.conn.Read(buf)
}

func (_self_ *MyConnect) Write(buf []byte) (int, error) {
	if _self_.conn == nil {
		return -1, nil
	}

	return _self_.conn.Write(buf)
}

func (_self_ *MyConnect) Close() error {
	if _self_.conn == nil {
		return nil
	}

	if _self_.is_closed {
		return nil
	}
	_self_.is_closed = true
	log.Printf("server close : %s\n", _self_.ToString())
	return _self_.conn.Close()
}

func (_self_ *MyConnect) ToString() string {
	if _self_.conn == nil {
		return ""
	}

	if _self_.address == "" || _self_.uuid == "" {
		return fmt.Sprintf("%s->%s", _self_.conn.LocalAddr().String(), _self_.conn.RemoteAddr().String())
	} else {
		return fmt.Sprintf("%s->%s, host=%s, uuid=%s", _self_.conn.LocalAddr().String(), _self_.conn.RemoteAddr().String(), _self_.address, _self_.uuid)
	}
}
