package main

import (
	"encoding/binary"
)

var SSL_3_0 = 0x0300
var TLS_1_0 = 0x0301
var TLS_1_1 = 0x0302
var TLS_1_2 = 0x0303
var TLS_1_3 = 0x0304

type Interface_Tls_Method interface {
	Parse(data []byte) bool
	toByte() []byte
	Modify(new_sni_name string) bool
	Restore() bool
}

type Interface interface {
	Modify_sni(new_sni_name []byte) int
}

type Tls_Shake_Record struct {
	_type       uint8
	_version    uint16
	_length     uint16
	_shake_hand Tls_Shake_Hand_Layer
}

func (record *Tls_Shake_Record) Restore() bool {
	custom_extension_index := -1
	for i := len(record._shake_hand._extensions) - 1; i >= 0; i-- {
		extension := &record._shake_hand._extensions[i]
		if extension._type == 65282 {
			custom_extension_index = i
			break
		}
	}

	var sni_extension *Tls_Shake_Hand_Layer_extension = nil
	for i := 0; i < len(record._shake_hand._extensions); i++ {
		extension := &record._shake_hand._extensions[i]
		if extension._type == 0x00 {
			sni_extension = extension
			break
		}
	}

	if custom_extension_index != -1 {
		var changed_length int = 0
		real_sni_name := record._shake_hand._extensions[custom_extension_index]._content
		for i := 0; i < len(real_sni_name); i++ {
			real_sni_name[i] = real_sni_name[i] - 1
		}
		if sni_extension != nil {
			changed_length += sni_extension.Modify_sni(real_sni_name)
		}

		changed_length -= (len(real_sni_name) + 2 + 2)
		record._shake_hand._extensions = append(record._shake_hand._extensions[:custom_extension_index], record._shake_hand._extensions[custom_extension_index+1:]...)
		appendTlsLength(record, changed_length)
	}
	return true
}

func (record *Tls_Shake_Record) Modify(new_sni_name *string) bool {
	if new_sni_name == nil {
		return false
	}

	var sni_extension *Tls_Shake_Hand_Layer_extension = nil
	for i := 0; i < len(record._shake_hand._extensions); i++ {
		extension := &record._shake_hand._extensions[i]
		if extension._type == 0x00 {
			sni_extension = extension
			break
		}
	}

	if sni_extension != nil {
		var changed_length int = 0
		old_sni_name := sni_extension._content[5:]
		changed_length += sni_extension.Modify_sni([]byte(*new_sni_name))
		custom_extension := Tls_Shake_Hand_Layer_extension{}
		custom_extension._type = 65282
		custom_extension._length = uint16(len(old_sni_name))
		//防检测
		custom_extension._content = make([]byte, len(old_sni_name))
		for i := 0; i < len(old_sni_name); i++ {
			custom_extension._content[i] = old_sni_name[i] + 1
		}
		record._shake_hand._extensions = append(record._shake_hand._extensions, custom_extension)
		changed_length += int(custom_extension._length) + 2 + 2
		appendTlsLength(record, changed_length)
		return true
	}
	return false
}

func (record *Tls_Shake_Record) ToByte() []byte {
	data := make([]byte, record._length+1+2+2)
	index := 0
	//record
	data[index] = record._type //1
	index += 1
	binary.BigEndian.PutUint16(data[index:index+2], record._version) // 2
	index += 2
	binary.BigEndian.PutUint16(data[index:index+2], record._length) // 2
	index += 2
	copy(data[index:], record._shake_hand.toByte()[0:])
	return data
}

type Tls_Shake_Hand_Layer struct {
	_type                       uint8
	_length                     [3]byte
	_version                    uint16
	_random                     [32]byte
	_session_length             uint8
	_session                    []byte
	_cipher_suites_length       uint16
	_cipher_suites              []byte
	_compression_methods_length uint8
	_compression_methods        []byte
	_extensions_length          uint16
	_extensions                 []Tls_Shake_Hand_Layer_extension
}

func (shake_hand_layer *Tls_Shake_Hand_Layer) toByte() []byte {
	length_data := make([]byte, 4)
	for i := 0; i < 3; i++ {
		length_data[i+1] = shake_hand_layer._length[i]
	}
	data := make([]byte, binary.BigEndian.Uint32(length_data)+1+3)
	index := 0
	//record
	data[index] = shake_hand_layer._type //1
	index += 1
	copy(data[index:index+3], shake_hand_layer._length[0:]) // 3
	index += 3
	binary.BigEndian.PutUint16(data[index:index+2], shake_hand_layer._version) // 2
	index += 2
	copy(data[index:], shake_hand_layer._random[0:])
	index += 32
	data[index] = shake_hand_layer._session_length
	index += 1
	copy(data[index:], shake_hand_layer._session[0:])
	index += int(shake_hand_layer._session_length)
	binary.BigEndian.PutUint16(data[index:index+2], shake_hand_layer._cipher_suites_length)
	index += 2
	copy(data[index:], shake_hand_layer._cipher_suites[0:])
	index += int(shake_hand_layer._cipher_suites_length)
	data[index] = shake_hand_layer._compression_methods_length
	index += 1
	copy(data[index:], shake_hand_layer._compression_methods[0:])
	index += int(shake_hand_layer._compression_methods_length)
	binary.BigEndian.PutUint16(data[index:index+2], shake_hand_layer._extensions_length)
	index += 2
	for i := 0; i < len(shake_hand_layer._extensions); i++ {
		data_extension := shake_hand_layer._extensions[i].toByte()
		copy(data[index:], data_extension[0:])
		index += len(data_extension)
	}
	return data
}

type Tls_Shake_Hand_Layer_extension struct {
	_type    uint16
	_length  uint16
	_content []byte
}

func (sni_extension *Tls_Shake_Hand_Layer_extension) Modify_sni(new_sni_name []byte) int {
	old_sni := Tls_sni{}
	old_sni._list_length = binary.BigEndian.Uint16(sni_extension._content[0:2])
	old_sni._type = sni_extension._content[2]
	old_sni._length = binary.BigEndian.Uint16(sni_extension._content[3:5])
	old_sni._name = sni_extension._content[5:]

	sni_extension._content = make([]byte, len(new_sni_name)+5)
	changed_length := (len(new_sni_name) - len(old_sni._name))
	sni_extension._length = uint16(int(sni_extension._length) + changed_length)
	binary.BigEndian.PutUint16(sni_extension._content[0:2], uint16(changed_length+int(old_sni._list_length)))
	binary.BigEndian.PutUint16(sni_extension._content[3:5], uint16(changed_length+int(old_sni._length)))

	copy(sni_extension._content[5:], new_sni_name)
	return changed_length
}

func (extension *Tls_Shake_Hand_Layer_extension) toByte() []byte {
	data := make([]byte, extension._length+2+2)
	index := 0
	//record
	binary.BigEndian.PutUint16(data[index:index+2], extension._type)
	index += 2
	binary.BigEndian.PutUint16(data[index:index+2], extension._length)
	index += 2
	copy(data[index:], extension._content[0:])
	return data
}

type Tls_sni struct {
	_list_length uint16
	_type        uint8
	_length      uint16
	_name        []byte
}

func appendTlsLength(record *Tls_Shake_Record, changed_len int) {
	record._shake_hand._extensions_length = uint16(int(record._shake_hand._extensions_length) + changed_len)
	_shark_hand_len_byte := make([]byte, 4)
	for i := 0; i < 3; i++ {
		_shark_hand_len_byte[i+1] = record._shake_hand._length[i]
	}
	_shark_hand_len := binary.BigEndian.Uint32(_shark_hand_len_byte)
	_shark_hand_len += uint32(changed_len)
	binary.BigEndian.PutUint32(_shark_hand_len_byte, _shark_hand_len)
	for i := 0; i < 3; i++ {
		record._shake_hand._length[i] = _shark_hand_len_byte[i+1]
	}
	record._length = uint16(int(record._length) + changed_len)
}

func (record *Tls_Shake_Record) Parse(data []byte) bool {
	if !(data[0] == 0x16 && data[1] == 0x03 && data[2] <= 0x04 && data[5] == 0x01) {
		return false
	}
	record._type = data[0]
	record._version = binary.BigEndian.Uint16(data[1:3])
	record._length = binary.BigEndian.Uint16(data[3:5])
	record._shake_hand._type = data[5]
	record._shake_hand._length[0] = data[6]
	record._shake_hand._length[1] = data[7]
	record._shake_hand._length[2] = data[8]
	record._shake_hand._version = binary.BigEndian.Uint16(data[9:11])
	index := 11
	for i := 0; i < 32; i++ {
		record._shake_hand._random[i] = data[index]
		index++
	}
	record._shake_hand._session_length = data[index]
	index++
	record._shake_hand._session = data[index : index+int(record._shake_hand._session_length)]
	index += int(record._shake_hand._session_length)
	record._shake_hand._cipher_suites_length = binary.BigEndian.Uint16(data[index : index+2])
	index += 2
	record._shake_hand._cipher_suites = data[index : index+int(record._shake_hand._cipher_suites_length)]
	index += int(record._shake_hand._cipher_suites_length)
	record._shake_hand._compression_methods_length = data[index]
	index++
	record._shake_hand._compression_methods = data[index : index+int(record._shake_hand._compression_methods_length)]
	index += int(record._shake_hand._compression_methods_length)
	record._shake_hand._extensions_length = binary.BigEndian.Uint16(data[index : index+2])
	index += 2
	end := index + int(record._shake_hand._extensions_length)
	for index < end {
		extension := Tls_Shake_Hand_Layer_extension{}
		extension._type = binary.BigEndian.Uint16(data[index : index+2])
		index += 2
		extension._length = binary.BigEndian.Uint16(data[index : index+2])
		index += 2
		if extension._length > 0 {
			conent_length := int(extension._length)
			extension._content = data[index : index+conent_length]
			index += conent_length
		}
		record._shake_hand._extensions = append(record._shake_hand._extensions, extension)
	}
	return true
}
