package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var last_read_count_map = make(map[string]int)
var read_speed_map = make(map[string]int)
var last_write_count_map = make(map[string]int)
var write_speed_map = make(map[string]int)

func start_http_server(address *string) {
	force_quit := false
	host, port := parseHostAndPort(address)
	bind_address := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[http server]:%s\n", bind_address)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html_content := "<html>"
		html_content += "<head><meta http-equiv=\"refresh\" content=\"1\"> <style>table{width:100%} th{text-align:left}</style></head>"
		html_content += "<body>"
		html_content += "<table>"
		var keys []int
		for k := range connect_map {
			v, _ := strconv.Atoi(k)
			keys = append(keys, v)
		}
		html_content += "<tr><th>UUID</th><th>Address</th><th>Host</th><th>Read</th><th>Write</th><th>Creat Time</th><th>End Time</th><th>State</th></tr>"
		sort.Ints(keys)
		for i := len(keys) - 1; i >= 0; i-- {
			html_content += "<tr>"
			conn := connect_map[fmt.Sprintf("%d", keys[i])]
			html_content += fmt.Sprintf("<td>%s</td>", conn._uuid)
			html_content += fmt.Sprintf("<td>%s</td>", fmt.Sprintf("%s->%s", conn._conn.LocalAddr().String(), conn._conn.RemoteAddr().String()))
			html_content += fmt.Sprintf("<td>%s</td>", conn._address)
			html_content += fmt.Sprintf("<td style='width:200px'>%s</td>", fmt.Sprintf("%d[%d Kbps]", conn._read_count, read_speed_map[conn._uuid]*8/1024))
			html_content += fmt.Sprintf("<td style='width:200px'>%s</td>", fmt.Sprintf("%d[%d Kbps]", conn._write_count, write_speed_map[conn._uuid]*8/1024))
			html_content += fmt.Sprintf("<td>%s</td>", time.Unix(conn._create_time, 0).Format("2006-01-02 15:04:05"))
			close_time := ""
			if conn._close_time != 0 {
				close_time = time.Unix(conn._close_time, 0).Format("2006-01-02 15:04:05")
			}
			html_content += fmt.Sprintf("<td>%s</td>", close_time)
			state := "closed"
			if !conn._is_closed {
				state = "open"
			}
			html_content += fmt.Sprintf("<td>%s<td>", state)
			html_content += "<tr>"
		}
		html_content += "</table>"
		html_content += "</body>"
		html_content += "</html>"
		w.Write([]byte(html_content))
	})

	go func() {
		for {
			time.Sleep(time.Second * 1)
			for k, v := range connect_map {
				if force_quit {
					return
				}
				_, ok := last_read_count_map[k]
				if ok {
					read_speed_map[k] = (v._read_count - last_read_count_map[k])
				}
				last_read_count_map[k] = v._read_count

				_, ok = last_write_count_map[k]
				if ok {
					write_speed_map[k] = (v._write_count - last_write_count_map[k])
				}
				last_write_count_map[k] = v._write_count
			}
		}
	}()
	http.ListenAndServe(bind_address, nil)
	force_quit = true
}
