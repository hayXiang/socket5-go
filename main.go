package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	sni_name := ""
	sock5_address := ":8888"
	nat_address := ""
	port_forward_address := "" //127.0.0.1:9999->127.0.0.1:8080,127.0.0.1:9999->127.0.0.1:8080
	address := ":5201"
	http_server := ""
	is_server_mode := false
	is_reverse_mode := true

	if len(os.Args) < 2 {
		log.Println("xsocks5 -L ${address} -S ${socket5_address}")
		return
	}

	for idx, args := range os.Args {
		if args == "--socks5" || args == "-S" {
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") == -1 {
				sock5_address = os.Args[idx+1]
			}
		}

		if args == "--listen" || args == "-L" {
			is_server_mode = true
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") == -1 {
				address = os.Args[idx+1]
			}
		}

		if args == "--nat" || args == "-N" {
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") == -1 {
				nat_address = os.Args[idx+1]
			}
		}

		if args == "--sni" || args == "-I" {
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") == -1 {
				sni_name = os.Args[idx+1]
			}
		}

		if args == "--port-forwad" || args == "-P" {
			is_server_mode = true
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") != -1 && strings.Index(os.Args[idx+1], "->") != -1 {
				port_forward_address = os.Args[idx+1]
			}
		}

		if args == "--http" || args == "-H" {
			if (idx < len(os.Args)-1) && strings.Index(os.Args[idx+1], "-") == -1 {
				http_server = os.Args[idx+1]
			}
		}
	}

	if !is_server_mode && is_reverse_mode {
		address = os.Args[1]
	}

	if is_server_mode {
		if is_reverse_mode {
			if sock5_address != "" {
				go socks5_inbound(&sock5_address, &sni_name)
			}

			if nat_address != "" {
				go nat_inbound(&nat_address, &sni_name)
			}

			if port_forward_address != "" {
				go port_forward_inbound(&port_forward_address)
			}

			if http_server != "" {
				go start_http_server(&http_server)
			}
			start_reverse_xsocket5_server(&address)
		} else {
			start_xsocket5_server(&address)
		}
	} else {
		if is_reverse_mode {
			start_reverse_xsocket5_client(&address)
		} else {
			log.Panic("not support yet")
		}
	}

}
