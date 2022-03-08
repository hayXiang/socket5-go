package main

import (
	"fmt"
	"log"
	"strings"
)

func port_forward_inbound(address *string) {
	port_forward_info_list := strings.Split(*address, ",")
	for i := 0; i < len(port_forward_info_list); i++ {
		port_forward_info := strings.Split(port_forward_info_list[i], "->")
		if len(port_forward_info) == 2 {
			local_host, local_port := parseHostAndPort(&port_forward_info[0])
			dest_host, dest_port := parseHostAndPort(&port_forward_info[1])
			bind_address := fmt.Sprintf("%s:%s", local_host, local_port)
			target_address := fmt.Sprintf("%s:%s", dest_host, dest_port)
			log.Printf("[port forward]:%s->%s\n", bind_address, target_address)
			go process(&bind_address, func(_ *MyConnect) (string, error) {
				return target_address, nil
			}, nil, nil)
		}
	}
}
