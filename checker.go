package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Avaliando snapshots e testando invariantes...")
	falhas := 0

	for snId := 1; snId <= 200; snId++ {
		estados := lerSnapshots(snId)
		if len(estados) < 3 { continue }

		qtdInMX := 0
		todosNoMX := true
		alguemWaiting := false

		for _, est := range estados {
			if est.estado == 2 { qtdInMX++ }
			if est.estado != 0 { todosNoMX = false }
			if strings.Contains(est.waitings, "1") { alguemWaiting = true }
		}

		// INV 1: no máximo um processo na SC.
		if qtdInMX > 1 {
			fmt.Printf("[FALHA] Snapshot %d violou Inv 1: %d processos na SC!\n", snId, qtdInMX)
			falhas++
		}

		// INV 2: se todos processos estão em "não quero a SC", então todos waitings tem que ser falsos.
		if todosNoMX && alguemWaiting {
			fmt.Printf("[FALHA] Snapshot %d violou Inv 2: Todos não querem a SC, mas há processos em waiting!\n", snId)
			falhas++
		}

		// INV 3: se um processo q está marcado como waiting em p, então p está na SC ou quer a SC.
		for _, p := range estados {
			for q_id, isWaiting := range p.waitings {
				if isWaiting == '1' {
					if p.estado == 0 { // Se P está em noMX
						fmt.Printf("[FALHA] Snapshot %d violou Inv 3: P%d está em noMX mas P%d está marcado como waiting nele!\n", snId, p.id, q_id)
						falhas++
					}
				}
			}
		}

		// INV 4: se um processo q quer a SC, ele não pode ser ignorado por quem já está na SC.
		for _, q := range estados {
			if q.estado == 1 { // Q quer a SC
				for _, p := range estados {
					if p.id != q.id {
						qEstaNoWaitingDeP := p.waitings[q.id] == '1'
						if p.estado == 2 && !qEstaNoWaitingDeP {
							fmt.Printf("[FALHA] Snapshot %d violou Inv 4: P%d está na SC, mas ignorou o pedido de P%d (q) que quer a SC!\n", snId, p.id, q.id)
							falhas++
						}
					}
				}
			}
		}
	}

	if falhas == 0 {
		fmt.Println("SUCESSO: Todas as invariantes passaram. O sistema está consistente.")
	} else {
		fmt.Printf("AVISO: O sistema registrou %d violações de invariantes.\n", falhas)
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