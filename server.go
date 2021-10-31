package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DataJsonReq struct {
	Command string `json:"command"`
	Arg     string `json:"arg"`
}

type DataJsonAns struct {
	Status  string `json:"status"`
	Request string `json:"request"`
	Arg     string `json:"arg"`
	Data    string `json:"data"`
}

func createErrorMessage(errorCode string, errorInfo string) []byte {
	data := DataJsonAns{
		Status:  errorCode,
		Request: errorInfo,
		Arg:     "",
		Data:    "",
	}
	ans, _ := json.Marshal(data)
	return ans
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Print(err)
			fmt.Println("can't connect")
			break
		}

		var request DataJsonReq

		err = json.Unmarshal(p, &request)

		if err != nil {
			conn.WriteMessage(messageType, createErrorMessage("cant read json", ""))
			break
		}

		switch strings.ToLower(request.Command) {
		case "open":
			{
				filenameRoot := request.Arg
				data, err := ioutil.ReadFile(filenameRoot)

				if err != nil {
					conn.WriteMessage(messageType, createErrorMessage("file doesn't exists", request.Arg))
					break
				}

				ans := DataJsonAns{
					Status:  "ok",
					Request: "open",
					Arg:     request.Arg,
					Data:    string(data),
				}

				ansData, err := json.Marshal(ans)

				if err != nil {
					fmt.Print(err)
					break
				}

				err = conn.WriteMessage(messageType, ansData)

				if err != nil {
					fmt.Print(err)
					break
				}
				break
			}
		case "check":
			{
				folderRoot := request.Arg

				files, err := ioutil.ReadDir(folderRoot)

				if err != nil {
					conn.WriteMessage(messageType, createErrorMessage("folder doesn't exists", request.Arg))
					break
				}

				var fileData string

				for _, file := range files {
					fileData += file.Name() + "\n"
				}

				ans := DataJsonAns{
					Status:  "ok",
					Request: "check",
					Arg:     request.Arg,
					Data:    fileData,
				}

				ansData, err := json.Marshal(ans)

				if err != nil {
					fmt.Print(err)
					break
				}

				err = conn.WriteMessage(messageType, ansData)

				if err != nil {
					fmt.Print(err)
					break
				}
				break
			}
		case "delete":
			{
				fileRoot := request.Arg

				_, err := os.Stat(fileRoot)

				if err != nil {
					conn.WriteMessage(messageType, createErrorMessage("file doesn't exists, can't delete", request.Arg))
					break
				}

				err = os.Remove(fileRoot)

				if err != nil {
					conn.WriteMessage(messageType, createErrorMessage("can't delete file cause of internal error", request.Arg))
					break
				}

				ans := DataJsonAns{
					Status:  "ok",
					Request: "delete",
					Arg:     request.Arg,
					Data:    "",
				}

				ansData, err := json.Marshal(ans)

				if err != nil {
					fmt.Print(err)
					break
				}

				err = conn.WriteMessage(messageType, ansData)

				if err != nil {
					fmt.Print(err)
					break
				}
				break
			}
		case "upload":
			{
				fileData := request.Arg

				data := strings.Split(fileData, "\n\n\n")

				file, err := os.Create("files/" + data[0])

				if err != nil {
					conn.WriteMessage(messageType, createErrorMessage("can't read file", request.Arg))
					break
				}

				defer file.Close()
				file.WriteString(data[1])

				ans := DataJsonAns{
					Status:  "ok",
					Request: "upload",
					Arg:     data[0],
					Data:    "",
				}

				ansData, err := json.Marshal(ans)

				if err != nil {
					fmt.Print(err)
					break
				}

				err = conn.WriteMessage(messageType, ansData)

				if err != nil {
					fmt.Print(err)
					break
				}
				break
			}
		}
	}
}

func main() {
	port := os.Getenv("PORT")
	fmt.Print(port)
	http.HandleFunc("/echo", echoHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		panic("Error: " + err.Error())
	}
}
