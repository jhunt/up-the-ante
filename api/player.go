package api

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type player struct {
	conn *websocket.Conn
	room *room
	send chan []byte
}

func (p *player) readPump() {
	defer func() {
		p.room.leave <- p
		p.conn.Close()
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %s\n", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
		p.room.chatter <- message
	}
}

func (p *player) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.send:
			fmt.Printf("sending [%s] to %p\n", string(message), p)
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(p.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-p.send)
			}

			if err := w.Close(); err != nil {
				return
			}
			fmt.Printf("done with %p writes for now; closing writer...\n", p)

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			fmt.Printf("pinging %p...\n", p)
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
