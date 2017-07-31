package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:2300")

	CheckError(err)

	fmt.Println("Service Starting...")

	defer listener.Close()

	timeout := 120

	for {
		conn, err := listener.Accept()
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		if err != nil {
			continue
		}
		Log(conn.RemoteAddr().String(), "new client connectioned")

		go handleConnection(conn)
	}
}

//长连接入口
func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024+8)
	//声明一个管道用于接收解包的数据
	readerChan := make(chan []byte, 16)
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	go reader(readerChan)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				Log(conn.RemoteAddr().String(), "read error : ", err)
				return
			}
		}
		if n == 0 {
			conn.Write([]byte("file recv finished!\r\n"))
			conn.Close()
			Log(conn.RemoteAddr().String(), "read end")
			return
		}
		tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChan)
	}
}

//解包
func Unpack(buffer []byte, readerChan chan []byte) []byte {
	length := len(buffer)
	if length <= 8 {
		return buffer
	}
	packlen := BytesToInt(buffer[4:9])

	if packlen == 0 {
		return buffer
	}

	if length < packlen {
		return buffer
	}

	if length == packlen {
		readerChan <- buffer
		return make([]byte, 0)
	}
	readerChan <- buffer[:packlen]
	return buffer[packlen:]
}

type FileInfo struct {
	FileName string
	FileSize int64
}

func reader(readerChan chan []byte) {
	data := make([]byte, 0)
	//receive fileinfo
	data = <-readerChan
	datatype := string(data[0:4])
	buffer := data[8:]
	if datatype != "info" {
		Log("read datatype:", datatype)
		Log("read error:is not received file info")
		return
	}

	infojson := string(buffer)
	Log("read fileinfo:", infojson)

	var fileinfo FileInfo
	err := json.Unmarshal(buffer, &fileinfo)
	if err != nil {
		Log("json unmarshal error:", err)
	}

	fileext := path.Ext(fileinfo.FileName)
	fo, err := os.Create("d:\\web\\receive\\" + GetGuid() + fileext)
	if err != nil {
		fmt.Println("os.Create" + err.Error())
	}
	defer fo.Close()

	fmt.Println("the file's name is:", fileinfo.FileName)
	fmt.Println("the file's size is:", strconv.FormatInt(fileinfo.FileSize, 10))

	for {
		select {
		case data = <-readerChan:
			datatype := string(data[0:4])
			buffer := data[8:]
			if datatype == "data" {
				//write to the file
				_, err = fo.Write(buffer)
				if err != nil {
					Log("write error:", err)
				}
			} else if datatype == "flag" {
				flag := string(buffer)
				if flag == "filerecvend" {
					fmt.Println("file receive success")
					return
				}
			}
		}
	}
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
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

//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
