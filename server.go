package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		fmt.Println(r.Host)
		return true // puedes añadir lógica adicional para validar el origen si es necesario
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	handleOperations(conn)
}

func main() {
	fmt.Println("Will start ws server on localhost:8033")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:8033", nil))
}

// ======================
// actual business logic:
// ======================

type WSMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type ScanFoldersType struct {
	Folders []string
}

type OperationErrorType struct {
	Error string `json:"error"`
}

func handleOperations(conn *websocket.Conn) {
	// infinite loop that listen for messages
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)

		if err != nil {
			log.Println("error reading payload")
			log.Println(err)
			break
		}

		switch msg.Action {
		case "scan":
			var data ScanFoldersType

			if err = jsoniter.Unmarshal(msg.Data, &data); err != nil {
				log.Println("Error decoding DATA JSON:", err)
				conn.WriteJSON(OperationErrorType{err.Error()})
				break
			}

			result, err := handleScan(data)

			if err != nil {
				log.Println(err)
				conn.WriteJSON(OperationErrorType{err.Error()})
				break
			}

			err = conn.WriteJSON(result)

			if err != nil {
				log.Println("Error sending JSON response")
			}

			result = ScanFoldersResult{}
			runtime.GC()
		}
	}
}
