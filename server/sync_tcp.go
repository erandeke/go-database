package server

import (
	"fmt"
	"go-database/config"
	"go-database/core"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

func RunSyncTcpSever() {

	log.Println("Starting synchornous TCP server on", config.Host, config.Port)

	//concurrent clients

	var con_clients = 0

	//listening to the host :Port

	Lsn, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Panic(err)
	}

	//open an infinite for lop for listening to clients that will be connecting

	for {

		c, err := Lsn.Accept() //blocking call waiting for client to connect
		if err != nil {
			panic(err)
		}

		//increment the no of concurrent clients
		con_clients += 1
		log.Println("client connected with address:", c.RemoteAddr(), "concurrent_clients :", con_clients)

		for {

			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				con_clients -= 1
				log.Println("Client disconnected", c.RemoteAddr(), "concurrent clients", con_clients)
				if err == io.EOF {
					log.Println("I am here ")
					break
				}
				log.Println(err)
			}

			log.Println("command", cmd)
			respondCommand(cmd, c)

		}

	}

}

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	//read the input from the socket
	//max read allowed in one shot is 512 byte ==> figure out why?
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:]) // reads the data from the connection. here "n" == number of bytes read or size of bytes ?
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArraysAsString(buf[:n]) //no of bytes that will have to process
	//for examle  ping -> +ping\r\n 7 bytes
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil

}

/* func respondCommand(t string, c net.Conn) error {

	// whatever command I have got I just need to echo back to client
	//writes the data to connection
	if _, err := c.Write([]byte(t)); err != nil {
		log.Println(err)
		return err

	}

	return nil
} */

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respondCommand(cmd *core.RedisCmd, c io.ReadWriter) error {
	err := core.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
	return nil
}
