package db

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Server -
type Server struct {
	listener         net.Listener
	quit             chan struct{}
	exited           chan struct{}
	db               memoryDB
	connections      map[int]net.Conn
	connCloseTimeout time.Duration
}

// NewServer returns a new  server Instance
func NewServer() *Server {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to create listener", err.Error())
	}
	srv := &Server{
		listener:         listener,
		quit:             make(chan struct{}),
		exited:           make(chan struct{}),
		db:               newDB(),
		connections:      map[int]net.Conn{},
		connCloseTimeout: 10 * time.Second,
	}
	go srv.serve()
	return srv
}

func (srv *Server) serve() {
	var id int
	fmt.Println("listening for clients")
	for {
		select {
		case <-srv.quit:
			fmt.Println("shutting down the server")
			err := srv.listener.Close()
			if err != nil {
				fmt.Println("could not close the listener", err.Error())
			}
			if len(srv.connections) > 0 {
				srv.warnConnections(srv.connCloseTimeout)
				<-time.After(srv.connCloseTimeout)
				srv.closeConnections()
			}

			close(srv.exited)
			return
		default:
			tcpListener := srv.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				fmt.Println("failed to set listener deadline", err.Error())
			}

			conn, err := tcpListener.Accept()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			if err != nil {
				fmt.Println("failed to accept connection: ", err.Error())
			}

			write(conn, "Welcome to MemoryDB Server")

			srv.connections[id] = conn
			go func(connectionID int) {
				fmt.Println("Client with Id: ", connectionID, "joined")
				srv.handleConn(conn)

				delete(srv.connections, connectionID)
				fmt.Println("client with id:", connectionID, "left")
			}(id)
			id++
		}
	}
}

func write(c net.Conn, str string) {
	_, err := fmt.Fprintf(c, "%s\n->", str)
	if err != nil {
		log.Fatal(err)
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		str := strings.ToLower(strings.TrimSpace(scanner.Text()))
		values := strings.Split(str, " ")
		switch {
		case len(values) == 3 && values[0] == "set":
			srv.db.set(values[1], values[2])
			write(conn, "Ok")
		case len(values) == 2 && values[0] == "get":
			val, found := srv.db.get(values[1])
			if !found {
				write(conn, fmt.Sprintf("key %s not found", values[1]))
			} else {
				write(conn, val)
			}
		case len(values) == 2 && values[0] == "delete":
			srv.db.delete(values[1])
			write(conn, "Ok")
		case len(values) == 1 && values[0] == "exit":
			if err := conn.Close(); err != nil {
				fmt.Println("could not close the connection", err.Error())
			}
		default:
			write(conn, fmt.Sprintf("UNKNOWN: %s", str))
		}
	}
}

func (srv *Server) closeConnections() {
	fmt.Println("closing all connections")
	for id, conn := range srv.connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("could not close connection with id: ", id, err.Error())
		}
	}
}

func (srv *Server) warnConnections(timeout time.Duration) {
	for _, conn := range srv.connections {
		write(conn, fmt.Sprintf("host wants to shutdown the server in: %s", srv.connCloseTimeout.String()))
	}
}

// Stop will kill the server and save the records to  file
func (srv *Server) Stop() {
	fmt.Println("stopping the database server")
	close(srv.quit)
	<-srv.exited
	fmt.Println("saving in-memory records to file")
	srv.db.save()
	fmt.Println("database server successfully stopped")
}
