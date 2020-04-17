package api

import (
	"fmt"
)

type room struct {
	history [][]byte

	chatter chan []byte
	join    chan *player
	leave   chan *player
	players map[*player]bool
}

func newRoom() *room {
	return &room{
		history: make([][]byte, 0),
		chatter: make(chan []byte),
		join:    make(chan *player),
		leave:   make(chan *player),
		players: make(map[*player]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case player := <-r.join:
			go func() {
				for _, message := range r.history {
					select {
					case player.send <- message:
					default:
						close(player.send)
						return
					}
				}
				r.players[player] = true
			}()

		case player := <-r.leave:
			if _, ok := r.players[player]; ok {
				delete(r.players, player)
				close(player.send)
			}

		case message := <-r.chatter:
			r.history = append(r.history, message)
			fmt.Printf("chatter <%s>\n", string(message))
			for player := range r.players {
				fmt.Printf(" -->> %p\n", player)
				select {
				case player.send <- message:
				default:
					close(player.send)
					delete(r.players, player)
				}
			}
			fmt.Printf("sent all <%s>\n", string(message))
		}
	}
}
