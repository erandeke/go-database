package core

import (
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n") // this is how we define nil in RESP
var RESP_OK []byte = []byte("+OK\r\n")  // this is how we define nil in RESP

func evalPING(args []string, c io.ReadWriter) error {

	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command") //as per redis server
	}
	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err

}

func evalSET(args []string, c io.ReadWriter) error {
	//set command  is like SET k  v so more than 1 argument
	if len(args) <= 1 {
		return errors.New("Not Valid Set command")
	}

	//initialize key, value
	var key, value string
	//expirry default
	var expiryDurationMs int64 = -1 // if expiration is not provided then bydefault redis returns -1

	key, value = args[0], args[1]
	oType, oEnc := DeduceTypeEncoding(value)

	//loop to take care for Expiry
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) { // that means K V EX <missing the time>
				return errors.New("Invalid syntax for ex ") // correct syntax is SET K V Ex 2
			}

			//get the expiry in seconds
			exDurationInSeconds, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errors.New("Error since the expiry provided is not in intger or its an out of range")
			}

			//convert into ms
			expiryDurationMs = exDurationInSeconds * 1000

		default:
			errors.New("ERR syntax error")

		}
	}

	PUT(key, NewObject(value, expiryDurationMs, oType, oEnc))
	c.Write([]byte("+OK\r\n"))
	return nil

}

func evalGET(args []string, c io.ReadWriter) error {
	//now to get from the cache we need to have only argument to be passed as a key so that we can retrieve the object out off it
	if len(args) != 1 {
		return errors.New("Expected exact one argument to be passed for getting the object from the redis cache")
	}

	//get the object
	obj := GET(args[0])

	//check if object exists
	if obj == nil {
		c.Write(RESP_NIL)
		return nil
	}
	//now we got the object here but we need to check if that is expired , if expired then return nil
	if obj.expiryAt != -1 && obj.expiryAt <= time.Now().UnixMilli() { // this means object got expired
		c.Write(RESP_NIL)
		return nil
	}

	c.Write(Encode(obj.value, false))
	return nil
}

func evalTTL(args []string, c io.ReadWriter) error {

	// to get the TTL we need exact one argument TTL k   <key>
	if len(args) != 1 {
		return errors.New("ERR invalid arguments ")
	}

	key := args[0]
	//get the object
	obj := GET(key)

	// if the key does not exist and TTL is fired on it return -2
	if obj == nil {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	//if object exists but no expiry is set on it
	if obj.expiryAt == -1 {
		c.Write([]byte(":-1\r\n"))
		return nil
	}

	//compute the time remaining to expire the object

	durationInMs := obj.expiryAt - time.Now().UnixMilli()

	//final check if the duration is < 0 that means object has been expired so return -2 since object will be flushed
	if durationInMs < 0 {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	//if the object is not expired then
	c.Write(Encode(int64(durationInMs/1000), false))
	return nil

}

// delete the key from redis cache
func evalDEL(args []string, c io.ReadWriter) error {
	var countOfDeletedObjects = 0

	for _, key := range args {
		if ok := DEL(key); ok {
			countOfDeletedObjects++
		}
	}
	c.Write(Encode(countOfDeletedObjects, false))
	return nil
}

// Expire k <it will expire the object >
// this command is setting for expiry  EXPIRE k1 10
func evalExpire(args []string, c io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("ERR syntax invalid number of arguments for expire")
	}
	var key string = args[0]
	//get the object
	obj := GET(key)

	expiryDurationInSeconds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errors.New("ERR in converting the string to int")
	}

	expiryDurationInMs := expiryDurationInSeconds * 1000
	// if the time out was not set or any other thing happend dure to which expire operation wasn't success
	if obj == nil {
		c.Write([]byte(":0\r\n"))
	}

	obj.expiryAt = time.Now().UnixMilli() + expiryDurationInMs

	//return 1 if the timeout was set
	c.Write([]byte(":1\r\n"))
	return nil

}

// This  method will write the append only file making the redis persistable so after crash, redis can be restructured from AOF
func evalBGREWRITEAOF(args []string, c io.ReadWriter) error {
	DumpAllAOF()
	c.Write([]byte("+OK\r\n"))
	return nil
}

func evalINCR(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'incr' command")
	}

	var key string = args[0]
	obj := GET(key)
	if obj == nil {
		obj = NewObject("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		PUT(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return errors.New("Type encoding is not a string")
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return errors.New("Type encoding is not a integer")
	}

	i, _ := strconv.ParseInt(obj.value.(string), 10, 64)
	i++
	obj.value = strconv.FormatInt(i, 10)

	c.Write(Encode(i, false))
	return nil
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {

	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	case "SET":
		return evalSET(cmd.Args, c)
	case "GET":
		return evalGET(cmd.Args, c)
	case "TTL":
		return evalTTL(cmd.Args, c)
	case "DEL":
		return evalDEL(cmd.Args, c)
	case "EXPIRE":
		return evalExpire(cmd.Args, c)
	case "BGREWRITEAOF":
		return evalBGREWRITEAOF(cmd.Args, c)
	case "INCR":
		return evalINCR(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c)

	}

}
