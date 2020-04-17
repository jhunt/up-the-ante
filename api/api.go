package api

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type API struct {
	rd    *redis.Client
	rooms map[string]*room
}

func New(addr string) API {
	return API{
		rooms: make(map[string]*room),
		rd: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (a API) RandomRoomPin() string {
	return "R-A-N-D-O-M"
}

func (a API) CreateRoom(pin string) error {
	if _, exists := a.rooms[pin]; exists {
		return nil
	}
	a.rooms[pin] = newRoom()
	go a.rooms[pin].run()
	return a.rd.Set(fmt.Sprintf("room:%s", pin), "1", 0).Err()
}

func (a API) RoomExists(pin string) bool {
	return a.rd.Exists(fmt.Sprintf("room:%s", pin)).Val() == 1
}

func (a API) Listen(bind string) error {
	router := mux.NewRouter()
	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir("htdocs"))))

	router.HandleFunc("/v1/create", func(w http.ResponseWriter, r *http.Request) {
		pin := a.RandomRoomPin()
		err := a.CreateRoom(pin)
		if err != nil {
			errJSON(w, err)
			return
		}

		okJSON(w, struct {
			Pin string `json:"pin"`
		}{
			Pin: pin,
		})
	}).Methods("POST")

	router.HandleFunc("/v1/join/{pin}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		room, ok := a.rooms[vars["pin"]]
		if !ok {
			w.WriteHeader(404)
			return
		}

		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errJSON(w, err)
			return
		}

		p := &player{conn: socket, room: room, send: make(chan []byte, 256)}
		go p.readPump()
		go p.writePump()
		room.join <- p
	}).Methods("GET")

	http.Handle("/", router)
	return http.ListenAndServe(bind, nil)
}
