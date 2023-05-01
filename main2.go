package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func showpinguin() string {
	f, err := os.Open("pinguin.txt")
	if err != nil {
		log.Fatal(err)
	}
	pingv := ""
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pingv += scanner.Text() + "\n"
		fmt.Printf("%s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return pingv
}

var (
	clients = make(map[string]net.Conn)
	mu      sync.Mutex
	leaving = make(chan message)
	join    = make(chan message)
)

func main() {
	showpinguin()
	port := ""
	arg := os.Args[1:]
	if len(arg) > 1 {
		fmt.Printf("[USAGE]: ./TCPChat $port")
		return
	}
	if len(arg) == 0 {
		port = "8989"
	} else if len(arg) == 1 {
		_, err := strconv.Atoi(arg[0])
		if err != nil {
			fmt.Println("ENTER NUMBER!")
			return
		}
		port = arg[0]
	}
	fmt.Printf("Listening on the port" + " : " + port)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}
	go broadcaster()
	defer l.Close()
	for {

		conn, err := l.Accept()
		if err != nil {
			return
		}
		mu.Lock()

		if len(clients) > 9 {
			conn.Write([]byte("Chat is full"))
			conn.Close()
		}
		mu.Unlock()
		go userConnection(conn)

	}
}

type message struct {
	text    string
	address string
}

func newMessages(msg string, username string) message {
	return message{
		text:    msg,
		address: username,
	}
}

var (
	messages = make(chan message)
	history  string
)

func userConnection(conn net.Conn) {
	defer conn.Close()
	username, errBool := GetName(conn)

	if len(username) < 3 {
		fmt.Fprint(conn, "ENTER A LONGER NAME,PLEASE!")
		return
	}
	if errBool {
		return
	}
	mu.Lock()
	clients[username] = conn
	mu.Unlock()
	conn.Write([]byte(history))
	join <- newMessages("  joined.", username)
	history += fmt.Sprintf("%s joined\n", username)
	time := time.Now().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("[%s][%s]:", time, username)
	input := bufio.NewScanner(conn)
	fmt.Fprint(conn, output)
	for input.Scan() {
		text := Checktext(input.Text())
		if lastCheckText(text) {
			output := fmt.Sprintf("[%s][%s]:%s", time, username, input.Text())

			history += output + "\n"
			fmt.Println(output)
			messages <- newMessages(output, username)
		} else {
			fmt.Fprintln(conn, "Enter normal text")
			fmt.Fprint(conn, "["+time+"]"+"["+username+"]"+":")
		}
	}
	mu.Lock()
	fmt.Println(username)
	delete(clients, username)
	mu.Unlock()
	leaving <- newMessages(" has left.", username)
	history += fmt.Sprintf("%s has left\n", username)
	return
}

func GetName(conn net.Conn) (string, bool) {
	pingv := showpinguin()
	conn.Write([]byte(pingv))
	conn.Write([]byte("[ENTER YOUR NAME:] "))
	input := bufio.NewScanner(conn)
	var name string
	if input.Scan() {
		text := strings.TrimSpace(input.Text())
		name = text
	}
	return name, CheckName(name, conn)
}

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			mu.Lock()
			for username, conn := range clients {
				time := time.Now().Format("2006-01-02 15:04:05")
				if msg.address != username {
					fmt.Fprintln(conn, ClearLine(msg.text)+msg.text)
				}

				fmt.Fprint(conn, "["+time+"]"+"["+username+"]"+":")

			}
			mu.Unlock()
		case msg := <-leaving:
			mu.Lock()
			for username, conn := range clients {
				time := time.Now().Format("2006-01-02 15:04:05")
				fmt.Fprintln(conn, ClearLine(msg.address+msg.text)+msg.address+msg.text)
				fmt.Fprint(conn, "["+time+"]"+"["+username+"]"+":")
			}
			mu.Unlock()
		case msg := <-join:
			mu.Lock()
			for username, conn := range clients {
				if msg.address != username {

					fmt.Fprintln(conn)
					time := time.Now().Format("2006-01-02 15:04:05")

					fmt.Fprintln(conn, ClearLine(msg.address+msg.text)+msg.address+msg.text)
					fmt.Fprint(conn, "["+time+"]"+"["+username+"]"+":")

				} else {
					continue
				}
			}
			mu.Unlock()

		}
	}
}

func CheckName(username string, conn net.Conn) bool {
	if strings.TrimSpace(username) == "" {
		conn.Write([]byte("You cannot enter empty text"))
		return true
	}
	mu.Lock()
	for name := range clients {
		if name == username {
			conn.Write([]byte("You cannot enter empty text'"))
			mu.Unlock()
			return true
		}
	}
	mu.Unlock()
	return false
}

func ClearLine(s string) string {
	return "\r" + strings.Repeat(" ", len(s)+len(s)+10) + "\r"
}

func Checktext(s string) string {
	return strings.TrimSpace(s)
}

func lastCheckText(s string) bool {
	if s == "" {
		return false
	}
	for _, w := range s {
		if (w >= 65 && w <= 90) || (w >= 97 && w <= 122) {
			return true
		} else {
			return false
		}
	}

	return true
}
