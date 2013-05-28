package main

import (
	"github.com/garyburd/go-websocket/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	bufsize int = 16384
)

func serveWs(w http.ResponseWriter, r *http.Request) {
	// Disallow anything other than GETs
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var ws *websocket.Conn
	var err error
	
	// Performs an HTTP 1.1 upgrade to switch protocols to WebSockets
	if ws, err = websocket.Upgrade(w, r.Header, nil, bufsize, bufsize); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	wordgame := NewWordGame(ws)

	defer wordgame.Cleanup()
	go wordgame.WritePump()
	wordgame.ReadPump()
}

func sigintCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	os.Exit(0)
}

func main() {
	// Catch signals
	go sigintCatcher()

	// Start serving
	var err error
	http.HandleFunc("/", serveWs)

	err = http.ListenAndServe("127.0.0.1:8888", nil)
	if err != nil {
		panic(err)
	}
}
