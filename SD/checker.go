package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("A iniciar validação dos snapshots...")
	falhas := 0

	for snId := 1; snId <= 50; snId++ {
		estados := lerSnapshots(snId)
		if len(estados) < 3 { continue }

		qtdInMX := 0
		todosNoMX := true
		alguemWaiting := false

		for _, est := range estados {
			if est.estado == 2 { qtdInMX++ }
			if est.estado != 0 { todosNoMX = false }
			if strings.Contains(est.waitings, "true") { alguemWaiting = true }
		}

		// Invariante 1
		if qtdInMX > 1 {
			fmt.Printf("[ERRO] Snapshot %d violou Inv 1: %d processos na SC!\n", snId, qtdInMX)
			falhas++
		}
		// Invariante 2
		if todosNoMX && alguemWaiting {
			fmt.Printf("[ERRO] Snapshot %d violou Inv 2: Ninguém quer SC mas há processos em espera!\n", snId)
			falhas++
		}
	}

	if falhas == 0 {
		fmt.Println("SUCESSO: Todas as invariantes passaram (Sistema Consistente).")
	} else {
		fmt.Printf("FALHA: %d violações detetadas.\n", falhas)
	}
}

type EstadoProc struct {
	id, estado, lcl int
	waitings        string
}

func lerSnapshots(snId int) []EstadoProc {
	var estados []EstadoProc
	for pid := 0; pid < 3; pid++ {
		nomeArq := fmt.Sprintf("snap_%d_p%d.txt", snId, pid)
		b, err := os.ReadFile(nomeArq)
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(b)), ";")
			id, _ := strconv.Atoi(parts[0])
			st, _ := strconv.Atoi(parts[1])
			lcl, _ := strconv.Atoi(parts[2])
			estados = append(estados, EstadoProc{id: id, estado: st, lcl: lcl, waitings: parts[3]})
		}
	}
	return estados
}