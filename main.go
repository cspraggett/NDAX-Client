package main

import (
	"encoding/json"
	"fmt"
	"log"
	"ndax/utils"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
)

var done chan interface{}
var interrupt chan os.Signal

type frame struct {
	M int    `json:"m"`
	I int    `json:"i"`
	N string `json:"n"`
	O string `json:"o"`
}

type login struct {
	APIKey    string `json:"apiKey"`
	Signature string `json:"signature"`
	UserId    string `json:"userId"`
	Nonce     string `json:"nonce"`
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	x := &frame{}
	for {
		err := connection.ReadJSON(x)
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %v\n", x)
	}
}

func sendHandler(connection *websocket.Conn) {
	defer close(done)
	apiKey := os.Getenv("API_KEY")
	userId := os.Getenv("USER_ID")
	nonce := os.Getenv("NONCE")
	secretKey := os.Getenv("SECRET_KEY")
	signature := utils.GetSignature(nonce, apiKey, userId, secretKey)

	loginFrame := frame{0, 2, "AuthenticateUser", ""}
	login := login{apiKey, signature, userId, nonce}
	log, _ := json.Marshal(login)
	loginFrame.O = string(log)

	connection.WriteJSON(loginFrame)
}

func main() {
	done = make(chan interface{})
	interrupt = make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt)

	socketUrl := "wss://api.ndax.io/WSGateway/"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websoket Server:", err)
	}
	defer conn.Close()
	fmt.Println("connection started")

	go sendHandler(conn)
	go receiveHandler(conn)

	for {
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 1000):
			err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
			if err != nil {
				log.Println("Error during writing to websocket", err)
				return
			}
		case <-interrupt:
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting...")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiveing channel. Exiting...")
			}
			return
		}
	}
}
