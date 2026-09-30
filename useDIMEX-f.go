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
	fmt.Printf("Processo %d iniciado...\n", myId)

	// O processo 0 dispara as centenas de snapshots concorrentemente
	if myId == 0 {
		go func() {
			for snapId := 1; snapId <= 200; snapId++ {
				time.Sleep(1 * time.Second)
				dmx.TriggerSnapshot(snapId)
			}
		}()
	}

	time.Sleep(3 * time.Second) // Aguarda a subida de todos os processos

	// Acesso intensivo ao recurso
	for i := 0; i < 400; i++ {
		dmx.Lock()
		
		escreverNoArquivo("mxOUT.txt", "|")
		time.Sleep(5 * time.Millisecond) // O acesso deve ser intensivo
		escreverNoArquivo("mxOUT.txt", ".")
		
		dmx.Unlock()
	}
	fmt.Printf("Processo %d finalizou seus acessos!\n", myId)
	select {} 
}

func escreverNoArquivo(nome, texto string) {
	f, _ := os.OpenFile(nome, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(texto)
}