package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

func sock5_destination_address(buf []byte) string {
	var host string
	switch buf[3] {
	case 0x01: //IP V4
		host = net.IP(buf[4:8]).String()
	case 0x03: //域名
		host = string(buf[5 : len(buf)-2]) //b[4]表示域名的长度
	case 0x04: //IP V6
		host = net.IP(buf[4:20]).String()
	}
	port := binary.BigEndian.Uint16(buf[len(buf)-2:])
	return fmt.Sprintf("%s:%d", host, port)
}

func sock5_auth(client *MyConnect, username *string, password *string) (string, error) {
	user_name_auth := false
	{
		buf := make([]byte, 1024)
		size, err := client.Read(buf)
		if err != nil {
			log.Println(err)
			return "", err
		}
		if buf[0] != 0x05 {
			log.Println("only support sock5")
			return "", errors.New("only support sock5")
		}
		auth_methods := buf[2:size]
		log.Println("socket protocal")

		for i := 0; i < len(auth_methods); i++ {
			if auth_methods[i] == 0x02 {
				user_name_auth = true
				break
			}
		}

		if user_name_auth {
			client.Write([]byte{0x05, 0x2})
		} else {
			client.Write([]byte{0x05, 0x0})
		}
	}

	if user_name_auth {
		buf := make([]byte, 1024)
		_, err := client.Read(buf)
		if err != nil {
			log.Println(err)
			return "", err
		}

		if buf[0] != 0x01 {
			log.Println("socket user name auth protocal is not correct")
			return "", errors.New("socket user name auth protocal is not correct")
		}

		username_len := buf[1]
		auth_username := string(buf[2 : 2+username_len])
		password_len := buf[2+username_len]
		auth_password := string(buf[3+username_len : 3+username_len+password_len])

		if (auth_username == *username && auth_password == *password) || (*username == "" && *password == "") {
			client.Write([]byte{0x05, 0x0})
		} else {
			client.Write([]byte{0x05, 0xF})
		}
	}

	buf := make([]byte, 1024)
	size, err := client.Read(buf)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if buf[0] != 0x05 {
		log.Println("only support socket protocal")
		return "", errors.New("only support socket protocal")
	}

	return sock5_destination_address(buf[0:size]), nil
}

type Socket5Uri struct {
	host     string
	port     string
	username string
	password string
}

func ParseSocket5Uri(address *string, uri *Socket5Uri) {
	index_split := strings.LastIndex(*address, "@")
	if index_split != -1 {
		username_password := (*address)[0:index_split]
		sockets_bind_address := (*address)[index_split+1:]
		uri.username, uri.password = parseHostAndPort(&username_password)
		uri.host, uri.port = parseHostAndPort(&sockets_bind_address)
	} else {
		uri.host, uri.port = parseHostAndPort(address)
	}

}

func socks5_inbound(address *string, sni_mask_name *string) {
	uri := Socket5Uri{}
	ParseSocket5Uri(address, &uri)
	bind_address := fmt.Sprintf("%s:%s", uri.host, uri.port)
	log.Printf("[socks5] listen on:%s\n", bind_address)
	process(&bind_address, func(client *MyConnect) (string, error) {
		return sock5_auth(client, &uri.username, &uri.password)
	}, func(client *MyConnect) {
		client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}, func(buffer []byte) []byte {
		record := Tls_Shake_Record{}
		if record.Parse(buffer) {
			record.Modify(sni_mask_name)
			return record.ToByte()
		}
		return nil
	})
}
