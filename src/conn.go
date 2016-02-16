package main

import (
	"github.com/gorilla/websocket"
	"net/http" 	
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub 
}

func (c *connection) reader() {	
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		username := c.h.connections[c]
		thisGame  := games[username]
		var rival string
		if thisGame.Players[0] == username {
     			rival = thisGame.Players[1]
		} else {
     			rival = thisGame.Players[0]
		}
		_, ok := connPlayer[rival]
		if ok {
			connPlayer[rival].send <- message			
		}		
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h *hub
}

func (wsh wsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	user, _ := aaa.CurrentUser(rw, req) 
	ws, err := upgrader.Upgrade(rw, req, nil)	
	if err != nil {
		return
	}	
	c := &connection{send: make(chan []byte, 256), ws: ws, h: wsh.h}		
	connPlayer[user.Username] = c	
	c.h.connections[c] = user.Username
	defer func() { c.h.unregister <- c }()
	go c.writer()
	c.reader()
}
