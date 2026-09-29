package main

import (
	"SD/DIMEX"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run useDIMEX-f.go <id_do_processo>")
		return
	}
	myId, _ := strconv.Atoi(os.Args[1])

	enderecos := []string{"127.0.0.1:5000", "127.0.0.1:5001", "127.0.0.1:5002"}
	dmx := DIMEX.NewDIMEX(enderecos, myId, false)
	fmt.Printf("Processo %d iniciado na porta %s...\n", myId, enderecos[myId])

	if myId == 0 {
		go func() {
			for snapId := 1; snapId <= 50; snapId++ {
				time.Sleep(2 * time.Second)
				fmt.Printf("\n[Snapshot] A disparar snapshot %d\n", snapId)
				dmx.TriggerSnapshot(snapId)
			}
		}()
	}

	time.Sleep(3 * time.Second) // Aguarda todos os processos iniciarem

	for i := 0; i < 200; i++ {
		fmt.Printf("[P%d] A pedir Lock...\n", myId)
		dmx.Lock()
		
		fmt.Printf("[P%d] Na SC! A escrever no ficheiro...\n", myId)
		escreverNoArquivo("mxOUT.txt", "|")
		time.Sleep(10 * time.Millisecond)
		escreverNoArquivo("mxOUT.txt", ".")
		
		dmx.Unlock()
	}
	fmt.Printf("Processo %d terminou os acessos!\n", myId)
	select {} // Mantém o processo ativo
}

func escreverNoArquivo(nome, texto string) {
	f, _ := os.OpenFile(nome, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(texto)
}