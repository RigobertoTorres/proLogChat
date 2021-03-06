package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	//"strings"
)

var TEST = 8888

var UDPBROAD = 8875

var HOST = 8876

var CLIENT2HOST = 8877

var HOST2CLIENT = 8878

var cns []net.Conn

var hostFlag bool = true

var hostUDP *net.UDPConn

var hostAddr *net.UDPAddr

var msgchan chan string
var sentchan chan string

type Client struct {
	conn     net.Conn
	nickname string
	ch       chan string
}

func main() {
	msgchan = make(chan string)
	sentchan = make(chan string)
	makeScreen()
	go updateScreen()
	msgchan <- "Starting quickchat..."
	netInit(10)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		curseExit()
		os.Exit(1)
	}()
	for {
		time.Sleep(8 * time.Millisecond)
	}
}

func netInit(connectionLimit int) {
	cns = make([]net.Conn, 0, connectionLimit)

	hostFlag = hostTest()

	if !hostFlag {
		//requestConnections()
		sendClientMessage()
	}

	if hostFlag {

		go acceptConnections()

	}
}

// broadcast message on port 8877
func sendClientMessage() {
	msgchan <- "What's your username?"
	username := <-sentchan
	msgchan <- "Set username to " + username
	for {
		text := <-sentchan
		//fmt.Println("Sending client message:" + text)
		conn, err := net.Dial("tcp", ":8877")
		if err != nil {
			fmt.Printf("Broadcast error: %v\n", err)
			break
		}
		var buffer bytes.Buffer
		buffer.WriteString(string(username))
		buffer.WriteString(": ")
		buffer.WriteString(text)
		conn.Write(buffer.Bytes())
		conn.Close()
	}
}

//broadcast availability on port 8876
func requestConnections() bool {

	msgchan <- "Client: Sending connection request"
	conn, err := net.Dial("tcp", ":8876")
	if err != nil {
		msgchan <- "Broadcast error"
		return false
	}
	io.WriteString(conn, "Requesting connection")
	go readIncomingMessages(conn)
	return true
}

func hostTest() (test bool) {
	laddr := net.TCPAddr{
		IP:   nil,
		Port: TEST,
	}
	_, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		test = !requestConnections()
		if !test {
			msgchan <- "Client!!"
		}
	} else {
		msgchan <- "Host!!"
		test = true
	}

	return
}

// listen for tcp responses on 8876
func acceptConnections() {
	msgchan <- "Listening for connections"
	ln, err := net.Listen("tcp", ":8876")
	if err != nil {
		msgchan <- "Can't accept connections! Err:"
		msgchan <- err.Error()
	} else {
		for {
			conn, err1 := ln.Accept()
			msgchan <- "Accepted tcp connection."
			if err1 != nil {
				msgchan <- "Error accepting connections:"
				msgchan <- err1.Error()
			} else {
				msgchan <- "Accepted tcp connection from:"
				msgchan <- conn.RemoteAddr().String()

				cns = append(cns, conn)
				go receiveClientMessage(msgchan, conn)

			}
		}
	}
}

func receiveClientMessage(msgchan chan<- string, clConn net.Conn) {
	msgchan <- "CL2HST: Listening for clients..."
	ln, err := net.Listen("tcp", ":8877")
	if err != nil {
		fmt.Printf("Error responding to UDP broadcast: %v\n", err)
	} else {
		for {
			buff := make([]byte, 2048)
			msgchan <- "CL2HST: Accepting client messages..."
			conn, err1 := ln.Accept()
			if err1 != nil {
				msgchan <- "CL2HOST: Error in accepting client message.."
			}
			if err1 == nil {
				//fmt.Println("CL2HST: Reading Client Message...")
				conn.Read(buff)
				//					var buffer bytes.Buffer
				//					buffer.WriteString("\n")
				//					buffer.WriteString(conn.LocalAddr().String())
				//					buffer.WriteString(": ")
				//					buffer.WriteString(string(buff))
				//					buffer.WriteString("\n")
				line := string(buff)
				msgchan <- line
				for _, c := range cns {
					io.WriteString(c, line + "\n")
				}
			}
		}
	}
}

func readIncomingMessages(c net.Conn) {
	//bufc := bufio.NewReader(c)
	for {
		buff := make([]byte, 2048)
		//line, _, err := bufc.ReadLine()
		c.Read(buff)
		//		if err != nil {
		//			fmt.Printf("TCP read error, closing connection: %v\n", err)
		//			c.Close()
		//			return
		//		}

		if buff != nil {
			msgchan <- string(buff)

		}
	}
}
