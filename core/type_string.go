package core

import "strconv"

func DeduceTypeEncoding(v string) (uint8, uint8) {
	oType := OBJ_TYPE_STRING

	//if the string is able to get convert into integer
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, OBJ_ENCODING_INT // take input into string and give output back in encoding
	}

	if len(v) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR //if the length of string lesser or equal to 44 then embeded encoding
	}
	return oType, OBJ_ENCODING_RAW

}
