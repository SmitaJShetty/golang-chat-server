package main

//first draft from: gopl book - chapter 8
import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadCaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			log.Print(err)
			return
		}

		go handleConn(conn)
	}
}

//listens on the global entering and leaving channels for announcements of arriving and departing clients
func broadCaster() {
	clients := make(map[client]bool)
	for {
		select {
		case mesg := <-messages:
			for cli := range clients {
				cli <- mesg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "you are " + who
	messages <- who + " has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ": " + input.Text()
	}

	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for mesg := range ch {
		fmt.Fprintln(conn, mesg)
	}
}
