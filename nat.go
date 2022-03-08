package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
)

const (
	SO_ORIGINAL_DST      = 80
	IP6T_SO_ORIGINAL_DST = 80
)

func getOriginalDestination(c net.Conn) (destination string, newConn net.Conn, err error) {
	tcpClientConn, ok := c.(*net.TCPConn)
	if !ok {
		err = errors.New("clientConn is not a *net.TCPConn")
		return
	}
	defer tcpClientConn.Close()

	clientConnFile, err := tcpClientConn.File()
	if err != nil {
		log.Println(err)
		return
	}
	defer clientConnFile.Close()

	addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println(err)
		return
	}

	newConn, err = net.FileConn(clientConnFile)
	if err != nil {
		log.Println(err)
		return
	}

	destination = fmt.Sprintf("%d.%d.%d.%d:%d",
		addr.Multiaddr[4],
		addr.Multiaddr[5],
		addr.Multiaddr[6],
		addr.Multiaddr[7],
		uint16(addr.Multiaddr[2])<<8+uint16(addr.Multiaddr[3]))
	return
}

func nat_inbound(address *string) {
	host, port := parseHostAndPort(address)
	bind_address := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[nat] listen on:%s\n", bind_address)
	process(&bind_address, func(client *MyConnect) (string, error) {
		destination, new_connect, _ := getOriginalDestination(client._conn)
		client._conn = new_connect
		return destination, nil
	}, nil, nil)
}
