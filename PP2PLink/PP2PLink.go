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
	// Servidor TCP: Escuta as mensagens recebidas
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

	// Cliente TCP: Envia as mensagens com Retry (garante a entrega)
	go func() {
		// Mapa para guardar uma fila separada para cada processo destino
		filasDeEnvio := make(map[string]chan string)

		for req := range module.Req {
			ch, existe := filasDeEnvio[req.To]
			if !existe {
				// Se a fila deste destino não existe, criamos uma com buffer
				ch = make(chan string, 100)
				filasDeEnvio[req.To] = ch

				// Inicia um ÚNICO carteiro (goroutine) dedicado para este destino.
				// Como ele pega uma mensagem de cada vez da fila, a ordem nunca é quebrada.
				go func(destino string, fila chan string) {
					for msg := range fila {
						for {
							conn, err := net.Dial("tcp", destino)
							if err == nil {
								fmt.Fprintf(conn, msg+"\n")
								conn.Close()
								break // Sai do loop de retry, vai pegar a próxima mensagem
							}
							// Espera 50ms e tenta de novo se o outro processo ainda não subiu
							time.Sleep(50 * time.Millisecond)
						}
					}
				}(req.To, ch)
			}
			// Coloca a mensagem na fila do destinatário correto
			ch <- req.Message
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