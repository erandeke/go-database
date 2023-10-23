package core

import (
	"bytes"
	"errors"
	"fmt"
)

// takes the string command in resp encoded form and returns the
// string read , delta and error
// +ping\r\n
func readSimpleString(data []byte) (string, int, error) {

	pos := 1
	for ; data[pos] != '\r'; pos++ {

	}

	return string(data[1:pos]), pos + 2, nil

}

// reads a RESP encoded error from data and returns
// the error string, the delta, and the error
// "-Key not found \r\n"
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// reads a RESP encoded integer from data and returns
// the error string, the delta, and the error
// ":100\r\n"
func readInteger64(data []byte) (int64, int, error) {
	var value int64 = 0
	pos := 1
	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

// read RESP enncoded Bulkstring
// $4\r\nPING\r\n
func readBulkString(data []byte) (string, int, error) {

	pos := 1
	len, delta := readLength(data[pos:])
	pos += delta
	return string(data[pos:(pos + len)]), pos + len + 2, nil

}

//read the length of string in bulk string

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}

	return 0, 0
}

// Read array
func readArray(data []byte) (interface{}, int, error) {

	pos := 1

	count, delta := readLength(data[pos:])
	pos += delta

	var elements []interface{} = make([]interface{}, count)
	for i := range elements {
		elem, delta, err := decodeCommand(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elements[i] = elem
		pos = pos + delta

	}
	return elements, pos, nil

}

// devoder for RESP
func decodeCommand(data []byte) (interface{}, int, error) {

	if len(data) == 0 {
		return nil, 0, errors.New("Invalid data")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '.':
		return readInteger64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil

}

// Decode the commands
func DecodeArraysAsString(data []byte) ([]string, error) {

	value, err := decode(data)
	if err != nil {
		return nil, err
	}
	objects := value.([]interface{}) //ping , hello kindOf
	tokens := make([]string, len(objects))
	for i := range tokens {
		tokens[i] = objects[i].(string)
	}

	return tokens, nil
}

//decoder

func decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("No data")
	}

	value, _, err := decodeCommand(data)
	return value, err

}

//encode to send back the pong

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v)) //simple string
		}
		return encodeString(v) //bulk string
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]string) {
			buf.Write(encodeString(b))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	}
	return []byte{}
}

func encodeString(v string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
}
