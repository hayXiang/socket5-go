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
	_conn               net.Conn
	_count              int
	_is_closed          bool
	_uuid               string
	_address            string
	_reserved_read_data []byte
}

func (_self_ *MyConnect) Forward(destination *MyConnect, process func([]byte) []byte) {
	if _self_._conn == nil || destination._conn == nil {
		return
	}
	log.Printf("[start] transfer data,%s\n", _self_.ToString())
	defer log.Printf("[end] transfer data, %s, count = %d\n", _self_.ToString(), _self_._count)
	_self_._count = 0
	for {
		buf := make([]byte, 1024)
		size, err := _self_._conn.Read(buf)
		if err != nil {
			log.Printf("read over,%s,error=%s\n", _self_.ToString(), err.Error())
			_self_.Close()
			break
		}
		print(string(buf[:size]))
		_self_._count += size
		if size == 4 && binary.BigEndian.Uint32(buf[0:size]) == binary.BigEndian.Uint32(XSOCKS_PROTOCAL_FORCE_QUIT) {
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
	if _self_._conn == nil {
		return -1, nil
	}
	return _self_._conn.Read(buf)
}

func (_self_ *MyConnect) Write(buf []byte) (int, error) {
	if _self_._conn == nil {
		return -1, nil
	}
	return _self_._conn.Write(buf)
}

func (_self_ *MyConnect) Close() error {
	if _self_._conn == nil {
		return nil
	}

	if _self_._is_closed {
		return nil
	}
	_self_._is_closed = true
	log.Printf("server close : %s\n", _self_.ToString())
	return _self_._conn.Close()
}

func (_self_ *MyConnect) ToString() string {
	if _self_._conn == nil {
		return ""
	}

	if _self_._address == "" || _self_._uuid == "" {
		return fmt.Sprintf("%s->%s", _self_._conn.LocalAddr().String(), _self_._conn.RemoteAddr().String())
	} else {
		return fmt.Sprintf("%s->%s, host=%s, uuid=%s", _self_._conn.LocalAddr().String(), _self_._conn.RemoteAddr().String(), _self_._address, _self_._uuid)
	}
}
