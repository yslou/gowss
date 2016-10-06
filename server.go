package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/net/websocket"
	"sync"
	"time"
)

var (
	ErrNotSupportedExtension = fmt.Errorf("Not supported extenstion")
	conns	= map[int]*websocket.Conn{}
	id = 0
	mux sync.Mutex
)

type msg struct {
	Code	int	`json:"code,omitable"`
	Ping	string	`json:"ping"`
}

func httpSendJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func loadPage(path string) (body []byte, err error) {
	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}
	if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".js") {
		return body, ErrNotSupportedExtension
	}
	f := filepath.Join("www", path)
	return ioutil.ReadFile(f)
}

func genericHandler(w http.ResponseWriter, r *http.Request) {
	b, err := loadPage(r.URL.Path)
	if err == nil {
		w.Write(b)
	} else {
		http.Error(w, err.Error(), 404)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	httpSendJSON(w, msg{Code: 0, Ping:"pong"})
}

func getId() int {
	mux.Lock()
	id++
	mux.Unlock()
	return id
}

func echo(ws *websocket.Conn) {
	id := getId()
	conns[id] = ws
	fmt.Println("Echoing... ws id=", id)
	for n := 0; n < 5; n++ {
		msg := "Hello  " + string(n+48)
		fmt.Println("Sending to client: " + msg)
		err := websocket.Message.Send(ws, msg)
		if err != nil {
			fmt.Println("Can't send")
			break
		}

		var reply string
		err = websocket.Message.Receive(ws, &reply)
		if err != nil {
			fmt.Println("Can't receive")
			break
		}
		fmt.Println("Received back from client: " + reply)
	}
	time.Sleep(60 * time.Second)
	delete(conns, id)
}

func broadcaster() {
	x := 0
	for {
		time.Sleep(10 * time.Second)
		for id, ws := range conns {
			msg := fmt.Sprintf("broadcast %v", x)
			fmt.Printf("[%v] Broadcast to client: %v\n", id, msg)
			err := websocket.Message.Send(ws, msg)
			if err != nil {
				fmt.Println("Can't send")
				break
			}
		}
		x++
	}
}

func main() {
	var err error

	http.HandleFunc("/ping:", ping)
	//http.Handle("/wss", websocket.Handler(echo))
	http.HandleFunc("/wss", func (w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-auth-token")
		fmt.Printf("Incomming websocket token = %v\n", token)
		s := websocket.Server{Handler: websocket.Handler(echo)}
		s.ServeHTTP(w, r)
	})
	http.HandleFunc("/", genericHandler)

	go broadcaster()
	fmt.Println("SERVER listen on :8080")
	err = http.ListenAndServeTLS(":8080", "server.pem", "server.key", nil)
	//err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("what? ", err)
	}
}
