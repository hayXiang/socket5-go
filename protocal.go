package main

var XSOCKS_PROTOCAL_START byte = 0x95
var XSOCKS_PROTOCAL_END byte = 0x27
var XSOCKS_PROTOCAL_VERSION byte = 0x01

//protocal type
var XSOCKS_PROTOCAL_TYPE_REGIST byte = 0x01
var XSOCKS_PROTOCAL_TYPE_CREATE_PROXY byte = 0x02
var XSOCKS_PROTOCAL_UUID_COMMAND = "0"

var XSOCKS_PROTOCAL_FORCE_QUIT = []byte{0x95, 0x27, 0x27, 0x95}

//protocal data
var XSOCKS_PROTOCAL_CLIENT_REGISTER = []byte{XSOCKS_PROTOCAL_START, 0x06, XSOCKS_PROTOCAL_VERSION, XSOCKS_PROTOCAL_TYPE_REGIST, 0x00, XSOCKS_PROTOCAL_END}

type Interface_Protocal_Method interface {
	Parse(data []byte) bool
	toByte() []byte
}

type XsocktCreateProxyProtocal struct {
	_flag_start  byte
	_total_len   byte
	_version     byte
	_type        byte
	_body_length byte
	_body        XsocktCreateProxyProtocalBoby
	_flag_end    byte
}

type XsocktCreateProxyProtocalBoby struct {
	_uuid_len    byte
	_uuid        []byte
	_address_len byte
	_address     []byte
}

func parse_protocal(data []byte) (begin int, protocal_len int, protocal_type int) {
	if data[0] != XSOCKS_PROTOCAL_START {
		return -1, -1, -1
	}
	protocal_len = int(data[1])
	if protocal_len > len(data) {
		return -1, -1, -1
	}

	if data[protocal_len-1] != XSOCKS_PROTOCAL_END {
		return -1, -1, -1
	}
	protocal_body_len := int(data[4])
	if protocal_body_len != protocal_len-6 {
		return -1, -1, -1
	}
	return 0, protocal_len, int(data[3])
}

func (protocal *XsocktCreateProxyProtocal) ToByte() []byte {
	body_data := protocal._body.ToByte()
	protocal._body_length = byte(len(body_data))
	protocal._total_len = 6 + protocal._body_length
	protocal_data := make([]byte, protocal._total_len)
	protocal_data[0] = XSOCKS_PROTOCAL_START
	protocal_data[1] = byte(protocal._total_len)
	protocal_data[2] = XSOCKS_PROTOCAL_VERSION
	protocal_data[3] = XSOCKS_PROTOCAL_TYPE_CREATE_PROXY
	protocal_data[4] = protocal._body_length
	copy(protocal_data[5:], body_data)
	protocal_data[protocal._total_len-1] = XSOCKS_PROTOCAL_END
	return protocal_data
}

func (protocal *XsocktCreateProxyProtocal) Parse(data []byte) bool {
	protocal._flag_start = data[0]
	protocal._total_len = data[1]
	protocal._version = data[2]
	protocal._type = data[3]
	protocal._body_length = data[4]
	protocal._body = XsocktCreateProxyProtocalBoby{}
	protocal._body.Parse(data[5 : 5+protocal._body_length])
	protocal._flag_end = data[protocal._total_len-1]
	return true
}

func (body *XsocktCreateProxyProtocalBoby) Parse(data []byte) bool {
	var index byte = 0
	body._uuid_len = data[index]
	index++
	body._uuid = data[index : index+body._uuid_len]
	index += body._uuid_len
	body._address_len = data[index]
	index++
	body._address = data[index : index+body._address_len]
	return true
}

func (body *XsocktCreateProxyProtocalBoby) ToByte() []byte {
	body._address_len = byte(len(body._address))
	body._uuid_len = byte(len(body._uuid))
	protocal_body_length := 1 + body._uuid_len + 1 + body._address_len
	protocal_data := make([]byte, 6+protocal_body_length)

	var index byte = 0
	protocal_data[index] = body._uuid_len
	index++
	copy(protocal_data[index:], body._uuid)
	index += body._uuid_len
	protocal_data[index] = body._address_len
	index++
	copy(protocal_data[index:], body._address)
	index += body._address_len
	protocal_data[index] = XSOCKS_PROTOCAL_END
	return protocal_data
}
