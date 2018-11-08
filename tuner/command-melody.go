package tuner

import (
	"context"
	"os/exec"

	"github.com/ddosakura/go-tuner/tuning"
)

// SwapSign - swap sign
type SwapSign int

const (
	_ SwapSign = iota
	// CmRun - CM_RUN
	CmRun
	// CmRunWaitCancel - CM_RUN_WAIT_CANCEL
	CmRunWaitCancel
	// CmExit - CM_EXIT
	CmExit
)

// SwapData - swap
type SwapData struct {
	Sign SwapSign
}

// CommandMelody - the melody of command
type CommandMelody struct {
	singleFlag bool
	payload    *exec.Cmd
	cancel     context.CancelFunc
}

// NewCommandMelody - new CommandMelody
func NewCommandMelody(payload *exec.Cmd) CommandMelody {
	return CommandMelody{
		payload:    payload,
		singleFlag: false,
	}
}

// NewCommandMelodyWithCancel - new CommandMelody
func NewCommandMelodyWithCancel(payload *exec.Cmd, cancel context.CancelFunc) CommandMelody {
	return CommandMelody{
		payload:    payload,
		singleFlag: false,
		cancel:     cancel,
	}
}

// Play - run command
func (cm CommandMelody) Play(t *tuning.Track, in interface{}) interface{} {
	swap := in.(SwapData)
	switch swap.Sign {
	case CmRun:
		cm.payload.Run()
		t.Finish()
	case CmRunWaitCancel:
		cm.payload.Run()
	case CmExit:
		cm.cancel()
		t.Finish()
	}
	return swap
}
