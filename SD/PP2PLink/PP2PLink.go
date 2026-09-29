package PP2PLink

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type PP2PLink_Req_Message struct {
	To      string
	Message string
}

type PP2PLink_Ind_Message struct {
	From    string
	Message string
}

type PP2PLink struct {
	Req     chan PP2PLink_Req_Message
	Ind     chan PP2PLink_Ind_Message
	address string
	dbg     bool
}

func NewPP2PLink(_address string, _dbg bool) *PP2PLink {
	p2p := &PP2PLink{
		Req:     make(chan PP2PLink_Req_Message, 50),
		Ind:     make(chan PP2PLink_Ind_Message, 50),
		address: _address,
		dbg:     _dbg,
	}
	p2p.Start()
	return p2p
}

func (module *PP2PLink) Start() {
	// Servidor: Escuta mensagens a chegar
	go func() {
		ln, err := net.Listen("tcp", module.address)
		if err != nil {
			fmt.Println("Erro ao iniciar servidor PL:", err)
			return
		}
		for {
			conn, err := ln.Accept()
			if err == nil {
				go module.handleConnection(conn)
			}
		}
	}()

	// Cliente: Envia mensagens com Retry (evita perda se o outro ainda não ligou)
	go func() {
		for req := range module.Req {
			go func(r PP2PLink_Req_Message) {
				for {
					conn, err := net.Dial("tcp", r.To)
					if err == nil {
						fmt.Fprintf(conn, r.Message+"\n")
						conn.Close()
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			}(req)
		}
	}()
}

func (module *PP2PLink) handleConnection(conn net.Conn) {
	defer conn.Close()
	message, _ := bufio.NewReader(conn).ReadString('\n')
	message = strings.TrimSpace(message)
	if message != "" {
		module.Ind <- PP2PLink_Ind_Message{From: conn.RemoteAddr().String(), Message: message}
	}
}