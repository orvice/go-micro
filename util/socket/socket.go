// Package socket provides a pseudo socket
package socket

import (
	"io"

	"github.com/micro/go-micro/transport"
)

// socket is our pseudo socket for transport.Socket
type socket struct {
	// closed
	closed chan bool
	// remote addr
	remote string
	// local addr
	local string
	// send chan
	send chan *transport.Message
	// recv chan
	recv chan *transport.Message
}

func (s *socket) SetLocal(l string) {
	s.local = l
}

func (s *socket) SetRemote(r string) {
	s.remote = r
}

// Accept passes a message to the socket which will be processed by the call to Recv
func (s *socket) Accept(m *transport.Message) error {
	select {
	case <-s.closed:
		return io.EOF
	case s.recv <- m:
		return nil
	}
	return nil
}

// Process takes the next message off the send queue created by a call to Send
func (s *socket) Process(m *transport.Message) error {
	select {
	case <-s.closed:
		return io.EOF
	case msg := <-s.send:
		*m = *msg
	}
	return nil
}

func (s *socket) Remote() string {
	return s.remote
}

func (s *socket) Local() string {
	return s.local
}

func (s *socket) Send(m *transport.Message) error {
	select {
	case <-s.closed:
		return io.EOF
	default:
		// no op
	}

	// make copy
	msg := &transport.Message{
		Header: make(map[string]string),
		Body:   m.Body,
	}

	for k, v := range m.Header {
		msg.Header[k] = v
	}

	// send a message
	select {
	case s.send <- msg:
	case <-s.closed:
		return io.EOF
	}

	return nil
}

func (s *socket) Recv(m *transport.Message) error {
	select {
	case <-s.closed:
		return io.EOF
	default:
		// no op
	}

	// receive a message
	select {
	case msg := <-s.recv:
		// set message
		*m = *msg
	case <-s.closed:
		return io.EOF
	}

	// return nil
	return nil
}

// Close closes the socket
func (s *socket) Close() error {
	select {
	case <-s.closed:
		// no op
	default:
		close(s.closed)
	}
	return nil
}

// New returns a new pseudo socket which can be used in the place of a transport socket.
// Messages are sent to the socket via Accept and receives from the socket via Process.
// SetLocal/SetRemote should be called before using the socket.
func New() *socket {
	return &socket{
		closed: make(chan bool),
		local:  "local",
		remote: "remote",
		send:   make(chan *transport.Message, 128),
		recv:   make(chan *transport.Message, 128),
	}
}
