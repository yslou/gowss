package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"crypto/tls"
	"net/url"
	"sync"
	"time"
	"math/rand"
	"net/http"
)

func rt(sn int, wg *sync.WaitGroup) {
	fmt.Printf("[%v] >>> start \n", sn)
	loc, _ := url.Parse("wss://localhost:8080/wss")
	org, _ := url.Parse("https://localhost")
	hdr := http.Header{}
	hdr.Add("x-auth-token", "deedbeef")
	//conn, err := websocket.Dial("wss://localhost:8080/wss", "", "https://localhost")
	conn, err := websocket.DialConfig(
		&websocket.Config{
			Location: loc,
			Origin: org,
			Version: websocket.ProtocolVersionHybi,
			TlsConfig: &tls.Config{
				InsecureSkipVerify :true,
			},
			Header: hdr,
		},
	)
	if err != nil {
		panic(err)
	}

	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				// graceful shutdown by server
				fmt.Printf("[%v]shutdown by server\n", sn);
				break
			}
			fmt.Printf("[%v]Couldn't receive msg %v\n", sn, err.Error())
			break
		}
		fmt.Printf("[%v]Received from server: %s\n", sn, msg)
		// return the msg
		err = websocket.Message.Send(conn, msg)
		if err != nil {
			fmt.Printf("[%v]Coduln't return msg", sn)
			break
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
	fmt.Printf("[%v]<<< end \n", sn)
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(3)

	go rt(0, &wg)
	go rt(1, &wg)
	go rt(2, &wg)

	fmt.Println("WAIT")
	wg.Wait()
	fmt.Println("DONE")
}
