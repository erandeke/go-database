package main

import (
	"bufio"
	"flag"
	"go-database/config"
	"go-database/server"

	//"go-database/server"
	"go-database/socket"
	"log"
	"os"
	"strings"
)

func SetupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the go database server")
	flag.IntVar(&config.Port, "port", 7379, "port for the go database server")
	flag.Parse()
}

func main() {

	SetupFlags()
	log.Println("Rolling the server")

	var test = make([]int, 0) //size specifies the length of slice
	test = append(test, 1, 2, 3, 4, 5)
	log.Println(test[1:4])

	//server.RunSyncTcpSever()
	s, err := socket.Listen("0.0.0.0", 7379)
	if err != nil {
		log.Println("Failed to create Socket:", err)
		os.Exit(1)
	}
	eventLoop, err := server.NewEventLoop(s)
	if err != nil {
		log.Println("Failed to create kqueue:", err)
		os.Exit(1)
	}
	log.Println("Server started. Waiting for incoming connections. ^C to exit.")

	eventLoop.Handle(func(s *socket.Socket) {
		reader := bufio.NewReader(s)
		for {
			line, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(line) == "" {
				break
			}
			log.Print("Read on ", s, ": ", line)
			s.Write([]byte(line))
		}
		s.Close()
	})
	//server.RunAsyncTcpServer()

}
