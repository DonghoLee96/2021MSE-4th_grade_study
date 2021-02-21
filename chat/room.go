package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/matryer/goblueprints/chapter1/trace"
)

// 채팅방을 의미
type room struct {

	// 수신 메세지를 보관하는 채널
	// 수신한 메세지는 다른 클라이언트로 전달돼야 함
	forward chan []byte

	// 방에 들어오려는 클라이언트를 위한 채널
	join chan *client

	// 방을 나가길 원하는 클라이언트를 위한 채널
	leave chan *client

	// 현재 채팅방에 있는 모든 클라이언트를 보유
	clients map[*client]bool

	// 방 안에서의 활동 추적정보를 수신
	tracer trace.Tracer
}

// 새 채팅방을 만들어주는 함수, 채널을 사용하려면 make를 써줘야 함
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

// 채팅방 내 활동들
func (r *room) run() {
	for {

		// join, leave, forward 채널을 살펴보며 가장 먼저 값이 들어온 채널의 case 하나만 실행,
		// 실행이 끝나면 또 값이 들어온 채널을 실행
		// select 은 하나의 case만 실행시키는 것을 보장 -> go의 동시성
		select {
		case client := <-r.join: // 들어오려는 클라이언트가 있는 경우 = 들어오는 클라이언트 채널에 값이 들어온 경우
			// joining
			r.clients[client] = true // 모든 클라이언트가 있는 해시 테이블에 그 클라이언트 추가
			r.tracer.Trace("New client joined")
		case client := <-r.leave: // 떠나려는 클라이언트가 있는 경우 = 떠나는 클라이언트 채널에 값이 들어온 경우
			// leaving
			delete(r.clients, client) // 모든 클라이언트가 있는 해시 테이블에서 그 클라이언트 제거
			close(client.send)        // 그 클라이언트가 보내는 메세지 채널 종료
			r.tracer.Trace("Client left")
		case msg := <-r.forward: // 클라이언트가 메세지를 전송한 경우 = 채팅방의 수신 채널에 값이 들어온 경우
			r.tracer.Trace("Message received: ", string(msg))
			for client := range r.clients { // 채팅방 안에 모든 클라이언트에게 그 메세지를 전달
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// 웹소켓을 사용하기 위한 HTTP 연결 업그레이드
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// 서버 - 클라이언트 사이 웹소켓 연결 과정(헨드쉐이킹)
// 요청이 들어오면 socket을 가져오고 client를 생성
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
