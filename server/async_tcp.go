package server

// DO NOT USE THIS CODE -------------- REFER kqueue.go

import (
	"go-database/config"
	"log"
	"net"
	"syscall"
)

func RunAsyncTcpServer() error {

	log.Println("starting an asynchronous TCP server on", config.Host, config.Port)

	max_clients := 10 //come to this conclusion

	var con_clients = 0

	// Create a socket
	/*
		Create socket file descriptor.

		- AF_INET = ARPA Internet protocols (IP)
		- SOCK_STREAM = sequenced, reliable, two-way connection based byte streams

		See https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/socket.2.html
	*/
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)

	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}, //127.0.0.1
	}); err != nil {
		return err
	}

	// Start listening
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// AsyncIO starts here!!

	kQueue, err := syscall.Kqueue()

	if err != nil {
		log.Println("failed to create kqueue file descriptor", err)
		return nil

	}
	log.Print("Created kqueue ", kQueue)

	//defer syscall.Close(kQueue) //not sure

	//register the events (incomming con)
	changeEvent := syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
		Fflags: 0,
		Data:   0,
		Udata:  nil,
	}

	/*
		The kevent() system call is used to register events with the queue, and return any pending events to the user.
		First, we register the change event with the queue, leaving the third argument empty.

		See https://www.freebsd.org/cgi/man.cgi?query=kqueue&sektion=2
	*/
	changeEventRegistered, err := syscall.Kevent(
		kQueue,
		[]syscall.Kevent_t{changeEvent},
		nil,
		nil,
	)

	if err != nil || changeEventRegistered == -1 {
		log.Println("failed to register change event", err)
		return nil

	}

	for {
		// see if any FD is ready for an IO
		/*
			   Event loop, checking the kernel queue for new events and executing handlers.
				Then, we query the queue for pending events, leaving the second argument empty.
		*/
		log.Println("Polling for new events...")
		newEvents := make([]syscall.Kevent_t, 10)
		numNewEvents, err := syscall.Kevent(kQueue, nil, newEvents, nil)
		if err != nil {
			/*
				We sometimes get syscall.Errno == 0x4 (EINTR) but that's ok it seems. Just keep polling.
				See https://reviews.llvm.org/D42206
			*/
			continue
		}

		for i := 0; i < numNewEvents; i++ {
			currentEvent := newEvents[i]
			eventFileDescriptor := int(currentEvent.Ident)
			if currentEvent.Flags&syscall.EV_EOF != 0 {
				/*
					Handle client closing the connection. Closing the event file descriptor removes it from the queue.
				*/
				log.Println("Client disconnected.")
				syscall.Close(eventFileDescriptor)
			} else if eventFileDescriptor == serverFD { //// if the socket server itself is ready for an IO
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("Failed to create Socket for connecting to client:", err)
					continue
				}
				log.Print("Accepted new connection ", fd, " from ", serverFD)

				// increase the number of concurrent clients count
				con_clients++
				//syscall.SetNonblock(serverFD, true)

				/*
					Watch for data coming in through the new connection.
				*/
				socketEvent := syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
					Fflags: 0,
					Data:   0,
					Udata:  nil,
				}
				socketEventRegistered, err := syscall.Kevent(kQueue, []syscall.Kevent_t{socketEvent}, nil, nil)

				if err != nil || socketEventRegistered == -1 {
					log.Print("Failed to register Socket event:", err)
					continue
				}

			} else {
				log.Println("here executed")
				/* comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					con_clients -= 1
					continue
				}
				respond(cmd, comm) */
			}
		}

	}
}
