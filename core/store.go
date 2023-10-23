package core

import "time"

//create struct that will hold the value and expiry SET K , V ex  so (v, ex will be storing as a part of custom object)

var dataStore map[string]*Obj

func init() {
	dataStore = make(map[string]*Obj)
}

func NewObject(value interface{}, durationInMs int64, oType uint8, oEnc uint8) *Obj {
	var expiryAt int64 = -1
	if durationInMs > 0 { // that means expiry is provided by the user now compute the exact expriy from current time
		expiryAt = time.Now().UnixMilli() + durationInMs
	}

	return &Obj{
		value:        value,
		expiryAt:     expiryAt,
		TypeEncoding: oType | oEnc,
	}

}

func PUT(k string, obj *Obj) {
	dataStore[k] = obj
}

func GET(k string) *Obj {
	v := dataStore[k]
	if v != nil {
		if v.expiryAt != -1 && v.expiryAt <= time.Now().UnixMilli() {
			DEL(k)
			return nil
		}
	}
	return v
}

func DEL(k string) bool {
	if _, ok := dataStore[k]; ok {
		delete(dataStore, k)
		return true
	}
	return false
}
