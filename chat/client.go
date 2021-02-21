package main

import (
	"github.com/gorilla/websocket" // 서버 소켓을 관리하기 위한 오픈소스 서드파티 패키지 "고릴라"
)

// 1명의 채팅 사용자 의미
type client struct {

	// 이 클라이언트의 웹 소켓
	socket *websocket.Conn

	// 메세지가 전송되는 채널
	send chan []byte

	// 클라이언트가 채팅하는 방
	room *room
}

// 클라이언트가 ReadMessage 메소드를 통해 소켓에서 읽고, 받은 메세지를 forward 채널로 계속 전송
func (c *client) read() {
	defer c.socket.Close() // 호출한 함수가 리턴되기 직전에 실행(return 이 여러 개인 경우, 추가로 소켓을 닫는 호출을 해줄 필요 X)
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

// WriteMessage 메소드를 통해 소켓에서 송신 채널의 메세지를 계속 수신
func (c *client) write() {
	defer c.socket.Close() // 호출한 함수가 리턴되기 직전에 실행(return 이 여러 개인 경우, 추가로 소켓을 닫는 호출을 해줄 필요 X)
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
