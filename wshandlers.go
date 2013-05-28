package main

import (
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"time"
)

type wordgame struct {
	ws     *websocket.Conn
	send   chan []byte
	finish bool //not protecting with a mutex, but you should
	gameId int64
}

func NewWordGame(w *websocket.Conn) (ret *wordgame) {
	wg := new(wordgame)
	wg.ws = w
	wg.send = make(chan []byte, 10)
	wg.finish = false
	wg.gameId = wg.makeNewGame()
	return wg
}

func (wg *wordgame) Cleanup() {
	log.Println("Cleaning up word game with id", wg.gameId)
	wg.removeGame()
	close(wg.send)
	wg.ws.Close()
}

func (wg *wordgame) ReadPump() {
	for {
		switch wg.finish {
		case true:
			return
		default:
			// Use deadline to detect dead or stuck clients.
			wg.ws.SetReadDeadline(time.Now().Add(300 * time.Second))
			op, reader, err := wg.ws.NextReader() //blocks
			if err != nil {
				log.Println("Error getting next reader: ", err)
				return
			}
			if op == websocket.OpClose ||
				op == websocket.OpBinary { //binary messages are not expected
				return // fall through to cleanup
			}
			if op == websocket.OpText {
				wg.processMessage(reader)
			}
			// ignore pongs, pings, and other types and cycle around
		}
	}
}

func (wg *wordgame) WritePump() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		if wg.finish {
			return
		}
		select {
		case message, ok := <-wg.send:
			if !ok {
				_ = wg.write(websocket.OpClose, []byte{}) // don't care about error, closing
				wg.finish = true
			} else if string(message) == "__MAGIC_CLOSE_VALUE__" {
				wg.finish = true
			} else if err := wg.write(websocket.OpText, message); err != nil {
				wg.finish = true
			}
		case <-ticker.C:
			if err := wg.write(websocket.OpPing, []byte{}); err != nil {
				wg.finish = true
			}
		}
	}
}

func (wg *wordgame) write(opCode int, payload []byte) error {
	wg.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
	w, err := wg.ws.NextWriter(opCode)
	if err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

