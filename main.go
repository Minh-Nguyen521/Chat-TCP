package main

import (
	"fmt"
	"log"
	"net"
)

const (
	Port     = "8080"
	safeMode = false
)

func User_IP(message string) string {
	if safeMode {
		return "hihi"
	}
	return message
}

type MessageType int

const (
	Connected MessageType = iota + 1
	Disconnected
	NewMessage
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

type Client struct {
	Conn net.Conn
}

func server(message chan Message) {
	clients := map[string]*Client{}
	for {
		author := <-message
		switch author.Type {
		case Connected:
			addr := author.Conn.RemoteAddr().(*net.TCPAddr)
			log.Printf("User %s Connected to %s", User_IP(addr.IP.String()), Port)
			clients[addr.IP.String()] = &Client{
				Conn: author.Conn,
			}
		case Disconnected:
			addr := author.Conn.RemoteAddr().(*net.TCPAddr)
			log.Printf("User %s Disconnected to %s", User_IP(addr.IP.String()), Port)
			delete(clients, addr.IP.String())
		case NewMessage:
			addr := author.Conn.RemoteAddr().(*net.TCPAddr)
			log.Printf("User %s sent: %s", User_IP(addr.IP.String()), author.Text)
			formattedMessage := fmt.Sprintf("User %s sent: %s", User_IP(addr.IP.String()), author.Text)
			for _, client := range clients {
				if client.Conn.RemoteAddr().String() != addr.String() {
					client.Conn.Write([]byte(formattedMessage))
				}
			}
		}
	}
}

func client(conn net.Conn, message chan Message) {
	buffer := make([]byte, 512)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from connection: %s", err)
			message <- Message{
				Type: Disconnected,
				Conn: conn,
				Text: "",
			}
			conn.Close()
			return
		}
		message <- Message{
			Type: NewMessage,
			Conn: conn,
			Text: string(buffer[:n]),
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen to epic port %s: %s\n", Port, err.Error())
	}
	log.Printf("Listening to TCP connections on port %s ...\n", Port)

	// defer ln.Close()
	message := make(chan Message)

	go server(message)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		message <- Message{
			Type: Connected,
			Conn: conn,
		}

		go client(conn, message)
	}
}
