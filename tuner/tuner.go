package tuner

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/ddosakura/go-tuner/tuning"
)

// Tuner - the command framework
type Tuner struct {
	music  *tuning.Music
	bufLen int
}

// New - new the tuner
func New() *Tuner {
	return &Tuner{
		music:  tuning.NewMusic(),
		bufLen: 1,
	}
}

// BufLen - change buffer length
func (t *Tuner) BufLen(bufLen int) *Tuner {
	t.bufLen = bufLen
	return t
}

type LinkMainSF func(CmStreamFilter) (InInterface, OutInterface, OutInterface)
type LinkSF func(CmStreamFilter) (OutInterface, OutInterface)

// Load - load command
func (t *Tuner) Load(args ...string) LinkMainSF {
	return func(sfs CmStreamFilter) (InInterface, OutInterface, OutInterface) {
		return t.load(args2command(args), sfs)
	}
}

// RunAfter - load command after
func (t *Tuner) RunAfter(oi OutInterface, args ...string) LinkSF {
	return func(sfs CmStreamFilter) (OutInterface, OutInterface) {
		return t.runAfter(oi, args2command(args), sfs)
	}
}

type DataSheetCallback = func(*DataSheet)

// Run - run the commands
func (t *Tuner) Run(i0 InInterface, sf tuning.Melody, callback DataSheetCallback) {
	id := t.music.AddMainMelody(&sf, t.bufLen)
	track, _ := t.music.Build(id)
	tuning.RunTrack(track).Super().Immediately()

	cmdInPipe, _ := i0.cmd.StdinPipe()

	mainTrack, _ := t.music.Build(i0.mainID)
	tuning.RunTrack(mainTrack).Super().Immediately()

	ds := DataSheet{
		Outputer: &cmdInPipe,
	}
	callback(&ds)

	track.Inputer(ds)

	mainTrack.Inputer(SwapData{
		Sign: CmRun,
	})
}

// GetMusic - get the music
func (t *Tuner) GetMusic() *tuning.Music {
	return t.music
}

func args2command(args []string) string {
	// Buffer 是一个实现了读写方法的可变大小的字节缓冲
	var buffer bytes.Buffer
	for i := range args {
		buffer.WriteString(args[i])
	}
	return buffer.String()
}

type CmStreamFilter struct {
	OutSf tuning.Melody
	ErrSf tuning.Melody
}

type OutInterface struct {
	mainID     string
	OutID      string
	readCloser *io.ReadCloser
}

type InInterface struct {
	mainID string
	cmd    *exec.Cmd
}

func (t *Tuner) load(CMD string, sfs CmStreamFilter) (InInterface, OutInterface, OutInterface) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, CMD)

	cmdOutPipe, _ := cmd.StdoutPipe()
	cmdErrPipe, _ := cmd.StderrPipe()

	var m tuning.Melody
	m = NewCommandMelodyWithCancel(cmd, cancel)
	mainID := t.music.AddMainMelody(&m, t.bufLen)
	outID := t.music.AddMainMelody(&sfs.OutSf, t.bufLen)
	errID := t.music.AddMainMelody(&sfs.ErrSf, t.bufLen)

	return InInterface{
			mainID: mainID,
			cmd:    cmd,
		}, OutInterface{
			mainID:     mainID,
			OutID:      outID,
			readCloser: &cmdOutPipe,
		}, OutInterface{
			mainID:     mainID,
			OutID:      errID,
			readCloser: &cmdErrPipe,
		}
}

func (t *Tuner) runAfter(oi OutInterface, CMD string, sfs CmStreamFilter) (i1 OutInterface, i2 OutInterface) {
	i0, i1, i2 := t.load(CMD, sfs)
	track, _ := t.music.Build(oi.OutID)
	tuning.RunTrack(track).Super().Immediately()

	cmdInPipe, _ := i0.cmd.StdinPipe()

	mainTrack, _ := t.music.Build(i1.mainID)
	tuning.RunTrack(mainTrack).Super().Immediately()

	track.Inputer(DataSheet{
		Inputer:  oi.readCloser,
		Outputer: &cmdInPipe,
	})

	mainTrack.Inputer(SwapData{
		Sign: CmRun,
	})
	return
}

func (t *Tuner) Active(i OutInterface) {
	track, _ := t.music.Build(i.OutID)
	tuning.RunTrack(track).Super().Immediately()
	track.Inputer(DataSheet{
		Inputer: i.readCloser,
	})
}
