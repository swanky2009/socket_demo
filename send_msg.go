package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	//"sync"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:2300")

	CheckError(err)

	defer conn.Close()

	//go handleConnectionRead(conn)

	messnager := make(chan int)
	timeout :=18
	conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	//心跳计时
	go HeartBeating(conn, messnager, timeout)

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "" {
			continue
		}
		if command == "quit" {
			return
		}
		status := sender(conn, command, messnager)
		if status == false {
			conn, err = net.Dial("tcp", "localhost:2300")
			CheckError(err)
			Log(conn.RemoteAddr().String(), "reconnect server")
			
			//心跳计时
			go HeartBeating(conn, messnager, timeout)

			sender(conn, command, messnager)			
		}
	}
}

func sender(conn net.Conn, command string, mess chan int) (status bool) {
	n, err := conn.Write([]byte(command))
	if err != nil {
		Log(conn.RemoteAddr().String(), "sender error :", err)
		return false
	}
	Log(conn.RemoteAddr().String(), "sender data length :", n)
	mess <- n
	return true
}

//心跳计时
func HeartBeating(conn net.Conn, writerChannel chan int, timeout int) {
	for {
		select {
		case n := <-writerChannel:
			Log(conn.RemoteAddr().String(), "heart witer data:", strconv.Itoa(n))
			conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			break
		case <-time.After(time.Second * time.Duration(timeout)):
			Log(conn.RemoteAddr().String(), "heart time out")
			conn.Close()
			return
		}
	}
}

func connDial() (conn net.Conn) {
	conn, err := net.Dial("tcp", "localhost:2300")
	CheckError(err)
	return conn
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func Log(v ...interface{}) {
	log.Println(v...)
}

func handleConnectionRead(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		data, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), " connection error : ", err)
			return
		}
		Log(conn.RemoteAddr().String(), "receive service data :", string(data))
		time.Sleep(9 * time.Second)
	}
}
