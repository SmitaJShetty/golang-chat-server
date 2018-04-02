package main

//first draft from: gopl book - chapter 8
import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const hostName string = "127.0.0.1"

type client chan<- string

type Message struct {
	From        string
	To          string
	MessageText string
}

type clientChannel struct {
	Name string
	Ch   client
}

var (
	entering = make(chan clientChannel)
	leaving  = make(chan clientChannel)
	messages = make(chan Message)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go relayer()

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

func relayer() {
	clients := make(map[clientChannel]bool)
	for {
		select {
		case mesg := <-messages:
			toWhoCh := getClientChannelFromName(clients, mesg.To)
			if toWhoCh != nil {
				toWhoCh <- mesg.From + " says:" + mesg.MessageText
			}

		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
		}
	}
}

func createChannelForNewClientAndRegister(clientName string) client {
	ch := make(chan string)
	ch <- "You are " + clientName
	entering <- getClientChannel(ch, clientName)
	return ch
}

func getClientChannelFromName(clients map[clientChannel]bool, clientName string) client {
	for cli := range clients {
		if cli.Name == clientName {
			return cli.Ch
		}
	}
	return nil
}

func handleConn(conn net.Conn) {
	client := make(chan string)
	go clientWriter(conn, client)

	who := conn.RemoteAddr().String()
	//toWho := conn.LocalAddr().String()
	client <- "you are " + who

	message := "entered :" + who
	messages <- getMessage("", who, message)

	cliCh := getClientChannel(client, who)
	entering <- cliCh

	input := bufio.NewScanner(conn)
	for input.Scan() {
		toWho, message := parseMessage(input.Text())
		messages <- getMessage(toWho, who, message)
	}

	leaving <- cliCh
	messages <- getMessage("", who, who+" has left")
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for mesg := range ch {
		fmt.Fprintln(conn, mesg)
	}
}

func getClientChannel(cli client, name string) clientChannel {
	return clientChannel{
		Ch:   cli,
		Name: name,
	}
}

func getMessage(toWho string, fromWho string, messageText string) Message {
	newMessage := Message{
		To:          toWho,
		From:        fromWho,
		MessageText: messageText,
	}

	return newMessage
}

func parseMessage(messageText string) (toWho string, message string) {
	//message format: address:text
	compositeMsg := strings.Split(messageText, ":")
	if len(compositeMsg) >= 2 {
		address, text := hostName+":"+compositeMsg[0], compositeMsg[1]
		return address, text
	}
	return "", ""
}

//listens on the global entering and leaving channels for announcements of arriving and departing clients
// func broadCaster() {
// 	clients := make(map[client]bool)
// 	for {
// 		select {
// 		case mesg := <-messages:
// 			for cli := range clients {
// 				cli <- mesg
// 			}
// 		case cli := <-entering:
// 			clients[cli] = true
// 		case cli := <-leaving:
// 			delete(clients, cli)
// 			close(cli)
// 		}
// 	}
// }
