package DIMEX

import (
	PP2PLink "SD/PP2PLink"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type State int
const (
	noMX State = iota
	wantMX
	inMX
)

type dmxReq int
const (
	ENTER dmxReq = iota
	EXIT
)

type dmxResp struct{}

type DIMEX_Module struct {
	Req         chan dmxReq
	Ind         chan dmxResp
	addresses   []string
	id          int
	st          State
	waiting     []bool
	lcl         int
	reqTs       int
	nbrResps    int
	dbg         bool
	Pp2plink    *PP2PLink.PP2PLink
	snapsFeitos map[int]bool
}

func NewDIMEX(_addresses []string, _id int, _dbg bool) *DIMEX_Module {
	p2p := PP2PLink.NewPP2PLink(_addresses[_id], _dbg)
	dmx := &DIMEX_Module{
		Req:         make(chan dmxReq, 1),
		Ind:         make(chan dmxResp, 1),
		addresses:   _addresses,
		id:          _id,
		st:          noMX,
		waiting:     make([]bool, len(_addresses)),
		lcl:         0,
		reqTs:       0,
		dbg:         _dbg,
		Pp2plink:    p2p,
		snapsFeitos: make(map[int]bool),
	}
	dmx.Start()
	return dmx
}

func (module *DIMEX_Module) Lock() {
	module.Req <- ENTER
	<-module.Ind // Bloqueia até a permissão chegar
}

func (module *DIMEX_Module) Unlock() {
	module.Req <- EXIT
}

func (module *DIMEX_Module) Start() {
	go func() {
		for {
			select {
			case dmxR := <-module.Req:
				if dmxR == ENTER {
					module.handleUponReqEntry()
				} else if dmxR == EXIT {
					module.handleUponReqExit()
				}
			case msgOutro := <-module.Pp2plink.Ind:
				if strings.Contains(msgOutro.Message, "respOk") {
					module.handleUponDeliverRespOk(msgOutro)
				} else if strings.Contains(msgOutro.Message, "reqEntry") {
					module.handleUponDeliverReqEntry(msgOutro)
				} else if strings.Contains(msgOutro.Message, "marker") {
					module.handleSnapshotMarker(msgOutro)
				}
			}
		}
	}()
}

func (module *DIMEX_Module) handleUponReqEntry() {
	module.lcl++
	module.reqTs = module.lcl
	module.nbrResps = 0
	module.st = wantMX

	for i, addr := range module.addresses {
		if i != module.id {
			msg := fmt.Sprintf("reqEntry|%d|%d", module.id, module.reqTs)
			module.sendToLink(addr, msg, " ")
		}
	}
}

func (module *DIMEX_Module) handleUponReqExit() {
	module.st = noMX
	for id, isWaiting := range module.waiting {
		if isWaiting {
			msg := fmt.Sprintf("respOk|%d", module.id)
			module.sendToLink(module.addresses[id], msg, " ")
			module.waiting[id] = false
		}
	}
}

func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	module.nbrResps++
	if module.nbrResps == len(module.addresses)-1 {
		module.st = inMX
		module.Ind <- dmxResp{}
	}
}

func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	parts := strings.Split(msgOutro.Message, "|")
	if len(parts) < 3 { return }
	idOutro, _ := strconv.Atoi(parts[1])
	tsOutro, _ := strconv.Atoi(parts[2])

	if tsOutro > module.lcl { module.lcl = tsOutro }
	module.lcl++

	euTenhoPrioridade := before(module.id, module.reqTs, idOutro, tsOutro)

	// Injecao de falha -> Para testar as invariantes, remova a exclamação "!" da variável "euTenhoPrioridade" abaixo
	if module.st == noMX || (module.st == wantMX && !euTenhoPrioridade) {
		msg := fmt.Sprintf("respOk|%d", module.id)
		module.sendToLink(module.addresses[idOutro], msg, " ")
	} else {
		module.waiting[idOutro] = true
	}
}

func (module *DIMEX_Module) TriggerSnapshot(snId int) {
	module.salvarEstado(snId)
	module.snapsFeitos[snId] = true
	for i, addr := range module.addresses {
		if i != module.id {
			msg := fmt.Sprintf("marker|%d", snId)
			module.sendToLink(addr, msg, " ")
		}
	}
}

func (module *DIMEX_Module) handleSnapshotMarker(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	parts := strings.Split(msgOutro.Message, "|")
	snId, _ := strconv.Atoi(parts[1])

	if !module.snapsFeitos[snId] {
		module.salvarEstado(snId)
		module.snapsFeitos[snId] = true
		for i, addr := range module.addresses {
			if i != module.id {
				msg := fmt.Sprintf("marker|%d", snId)
				module.sendToLink(addr, msg, " ")
			}
		}
	}
}

func (module *DIMEX_Module) salvarEstado(snId int) {
	nomeArq := fmt.Sprintf("snap_%d_p%d.txt", snId, module.id)
	
	// Converte o array booleano de waiting em uma string "010" para facilitar a leitura no checker
	waitStr := ""
	for _, w := range module.waiting {
		if w { waitStr += "1" } else { waitStr += "0" }
	}

	conteudo := fmt.Sprintf("%d;%d;%d;%s\n", module.id, module.st, module.lcl, waitStr)
	f, _ := os.OpenFile(nomeArq, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer f.Close()
	f.WriteString(conteudo)
}

func (module *DIMEX_Module) sendToLink(address string, content string, space string) {
	module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{To: address, Message: content}
}

func before(oneId, oneTs, othId, othTs int) bool {
	if oneTs < othTs { return true }
	if oneTs > othTs { return false }
	return oneId < othId
}