// To prepare to connect to the device, run:
// adb forward tcp:1234 tcp:9000

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"code.google.com/p/go.net/websocket"
)

type DeviceStatus string

const (
	DEVICE_DISCONNECTED = "disconnected"
	DEVICE_CONNECTED    = "connected"
)

type DeviceStatusMessage struct {
	DeviceStatus DeviceStatus
}

type Connection struct {
	messages chan string
	deviceStatus chan DeviceStatus
	lastKnownStatus DeviceStatus
}

func NewConnection() *Connection {
	return &Connection{
		messages: make(chan string),
		deviceStatus: make(chan DeviceStatus),
		lastKnownStatus: DEVICE_DISCONNECTED,
	}
}

func SendDeviceStatusMessage(ws *websocket.Conn, status DeviceStatus) error {
	msg, err := json.Marshal(DeviceStatusMessage{status})
	if err != nil {
		panic("json.Marshal");
	}
	_, err = ws.Write(msg)
	return err
}

func MakeEchoServer(c *Connection) websocket.Handler {
	return func(ws *websocket.Conn) {
		SendDeviceStatusMessage(ws, c.lastKnownStatus)
		for {
			select {
			case status := <-c.deviceStatus:
				c.lastKnownStatus = status
				err := SendDeviceStatusMessage(ws, status)
				if err != nil {
					return
				}
			case message := <-c.messages:
				_, err := io.WriteString(ws, message)
				if err != nil {
					return
				}
			}
		}
	}
}

func StartHTTP(c *Connection) {
	http.Handle("/echo", websocket.Handler(MakeEchoServer(c)))
	http.Handle("/", http.FileServer(http.Dir("static")))

	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error());
	}
}

func ConnectToAndroid(c *Connection) {
	var timeout time.Duration = 1
	for {
		conn, err := net.Dial("tcp", "localhost:1234")
		if err != nil {
			fmt.Printf("Dial: %s\n", err.Error());
			time.Sleep(timeout * time.Millisecond)
			timeout *= 2
			continue;
		}
		timeout = 1
		fmt.Printf("connected\n")
		c.deviceStatus <- DEVICE_CONNECTED
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				c.deviceStatus <- DEVICE_DISCONNECTED
				fmt.Printf("%s\n", err.Error())
				break
			}
			c.messages <- message
		}
	}
}

func main() {
	c := NewConnection()
	go StartHTTP(c)
	ConnectToAndroid(c)
}
