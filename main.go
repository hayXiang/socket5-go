package main

import (
	"log"
	"os"
)

func main() {
	username := ""
	password := ""
	sock5_address := ":8888"
	address := ":5201"
	is_server_mode := false
	is_reverse_mode := true

	if len(os.Args) < 2 {
		log.Println("xsocks5 -L ${address} -S ${socket5_address} -U ${username} -U ${password}")
		return
	}

	for idx, args := range os.Args {
		if args == "--username" || args == "-U" {
			username = os.Args[idx+1]
		}

		if args == "--password" || args == "-P" {
			username = os.Args[idx+1]
		}

		if args == "--socks5" || args == "-S" {
			sock5_address = os.Args[idx+1]
		}

		if args == "--address" || args == "-A" {
			address = os.Args[idx+1]
		}

		if args == "--listen" || args == "-L" {
			is_server_mode = true
		}

		if args == "--reverse" || args == "-R" {
			is_reverse_mode = true
		}
	}

	if !is_server_mode && is_reverse_mode {
		address = os.Args[1]
	}

	if is_server_mode && !is_reverse_mode {
		address = os.Args[1]
	}

	if is_server_mode {
		if is_reverse_mode {
			start_reverse_xsocket5_server(sock5_address, username, password, address)
		} else {
			startt_xsocket5_server(address)
		}
	} else {
		if is_reverse_mode {
			start_reverse_xsocket5_client(address)
		} else {
			log.Panic("not support yet")
		}
	}

}
