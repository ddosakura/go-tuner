package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ddosakura/go-tuner/tuner"
	"github.com/ddosakura/go-tuner/tuning"
)

type CustomMelody struct {
	inStr  string
	outStr string
}

type CustomMelodyData struct {
	hasNext bool
	data    string
	t       *testing.T
}

func (cm CustomMelody) Play(t *tuning.Track, in interface{}) interface{} {
	if cm.inStr == "" {
		cm.inStr = "Hello World!"
	}
	if cm.outStr == "" {
		cm.outStr = "HELLO WORLD!"
	}

	data := in.(CustomMelodyData)
	if data.data == cm.inStr {
		fmt.Println("input ok!")
	} else if data.data == cm.outStr {
		fmt.Println("output ok!")
	} else {
		data.t.Fatal("unknow data")
		return nil
	}
	if !data.hasNext {
		fmt.Println("EXIT")
		t.Finish()
	}
	return data
}

func testRectifier(t *testing.T, callback func(*testing.T, *tuning.Music, string, *tuning.Melody)) {
	music := tuning.NewMusic()

	var m1 tuning.Melody
	m1 = CustomMelody{}
	id := music.AddMainMelody(&m1, 1)

	var m2 tuning.Melody
	m2 = tuner.InputRectifier{
		Transfer: func(in interface{}) (ds tuner.DataSheet) {
			data := in.(CustomMelodyData)
			// fmt.Println("m2 transfer...")
			ds.Buffer = []byte(data.data)
			ds.HasNext = data.hasNext
			return
		},
	}
	id2, _ := music.AddSubMelody(id, &m2, 1)

	var m3 tuning.Melody
	m3 = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			// fmt.Println("m3 transfer...")
			return CustomMelodyData{
				data:    strings.ToUpper(string(in.Buffer)),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}

	//id3, _ := music.AddSubMelody(id2, &m3, 1)
	callback(t, music, id2, &m3)

	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()
	track.Inputer(CustomMelodyData{
		data:    "Hello World!",
		t:       t,
		hasNext: true,
	})
	track.Inputer(CustomMelodyData{
		data:    "Hello World!",
		t:       t,
		hasNext: true,
	})
	track.Inputer(CustomMelodyData{
		data:    "Hello World!",
		t:       t,
		hasNext: false,
	})
	tuning.Wait()
}

func TestRectifier(t *testing.T) {
	testRectifier(t, func(t *testing.T, music *tuning.Music, id2 string, m3 *tuning.Melody) {
		id3, _ := music.AddSubMelody(id2, m3, 1)

		var m4 tuning.Melody
		m4 = CustomMelody{}
		music.AddSubMelody(id3, &m4, 1)
	})
}

func TestStreamFilter_IR_SF_OR(t *testing.T) {
	testRectifier(t, func(t *testing.T, music *tuning.Music, id2 string, m3 *tuning.Melody) {
		var m tuning.Melody
		m = tuner.StreamFilter{
			Filter: func(buf []byte) []byte {
				return buf[:5]
			},
		}
		id, _ := music.AddSubMelody(id2, &m, 1)

		id3, _ := music.AddSubMelody(id, m3, 1)

		var m4 tuning.Melody
		m4 = CustomMelody{
			outStr: "HELLO",
		}
		music.AddSubMelody(id3, &m4, 1)
	})
}

func TestStreamFilter_IR_SF_WriteCloser(t *testing.T) {
	music := tuning.NewMusic()
	pr, pw, _ := os.Pipe()
	var rc io.ReadCloser
	var wc io.WriteCloser
	rc = pr
	wc = pw
	go func() {
		defer func() {
			_ = recover()
		}()
		for {
			var buffer = make([]byte, 1024)
			n, err := rc.Read(buffer)
			if err != nil {
				return
			}
			buffer = buffer[:n]
			fmt.Println(string(buffer))
		}
	}()

	var inMelody tuning.Melody
	inMelody = tuner.InputRectifier{
		Transfer: func(in interface{}) (ds tuner.DataSheet) {
			data := in.([]byte)
			if len(data) == 0 {
				return tuner.DataSheet{
					Buffer:   []byte("byebye!"),
					HasNext:  false,
					Outputer: &wc,
				}
			}
			return tuner.DataSheet{
				Buffer:   data,
				HasNext:  true,
				Outputer: &wc,
			}
		},
	}
	id := music.AddMainMelody(&inMelody, 1)

	var m tuning.Melody
	m = tuner.StreamFilter{
		Filter: func(buf []byte) []byte {
			return buf[:5]
		},
	}
	music.AddSubMelody(id, &m, 1)

	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()

	track.Inputer([]byte("Hello World!"))
	track.Inputer([]byte("A, hello!"))
	track.Inputer([]byte(""))

	tuning.Wait()
}

func TestStreamFilter_ReadCloser_SF_WriteCloser(t *testing.T) {
	music := tuning.NewMusic()

	pr1, pw1, _ := os.Pipe()
	pr2, pw2, _ := os.Pipe()
	var rc1, rc2 io.ReadCloser
	var wc1, wc2 io.WriteCloser
	rc1 = pr1
	wc1 = pw1
	rc2 = pr2
	wc2 = pw2

	go func() {
		defer func() {
			_ = recover()
		}()
		for {
			var buffer = make([]byte, 1024)
			n, err := rc1.Read(buffer)
			if err != nil {
				return
			}
			buffer = buffer[:n]
			fmt.Println("ans {", string(buffer), "}")
		}
	}()

	var m tuning.Melody
	m = tuner.StreamFilter{
		Filter: func(buf []byte) []byte {
			// fmt.Println("input {", string(buf), "}")
			// fmt.Println("output {", string(buf[:3]), "}")
			return buf[:3]
		},
		BufLen: 5,
	}
	id := music.AddMainMelody(&m, 1)

	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()

	track.Inputer(tuner.DataSheet{
		Inputer:  &rc2,
		Outputer: &wc1,
	})

	go func() {
		for i := 0; i < 5; i++ {
			wc2.Write([]byte(strconv.Itoa(i) + ", Hello!"))
			// time.Sleep(time.Second * 1)
		}
		wc2.Close()
	}()

	tuning.Wait()
}

type OutputerMelody struct {
}

func (OutputerMelody) Play(t *tuning.Track, in interface{}) interface{} {
	data := in.(CustomMelodyData)
	if !data.hasNext {
		t.Finish()
	}
	if len(data.data) > 0 {
		fmt.Print(data.data)
	}
	return nil
}

type ErrputerMelody struct {
}

func (ErrputerMelody) Play(t *tuning.Track, in interface{}) interface{} {
	data := in.(CustomMelodyData)
	if !data.hasNext {
		t.Finish()
	}
	if len(data.data) > 0 {
		fmt.Printf("%c[%d;;%dm%s%c[0m", 0x1B, 1, 31, data.data, 0x1B)
	}
	return nil
}

func TestStreamFilter_ReadCloser_SF_OR(t *testing.T) {
	music := tuning.NewMusic()

	pr2, pw2, _ := os.Pipe()
	var rc2 io.ReadCloser
	var wc2 io.WriteCloser
	rc2 = pr2
	wc2 = pw2

	var m tuning.Melody
	m = tuner.StreamFilter{
		Filter: func(buf []byte) []byte {
			// fmt.Println("input {", string(buf), "}")
			// fmt.Println("output {", string(buf[:3]), "}")
			return buf[:3]
		},
		BufLen: 5,
	}
	id := music.AddMainMelody(&m, 1)

	var orm tuning.Melody
	orm = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			// fmt.Println("transfer", in)
			return CustomMelodyData{
				data:    strings.ToUpper(string(in.Buffer)),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}
	ormID, _ := music.AddSubMelody(id, &orm, 1)

	var om tuning.Melody
	om = OutputerMelody{}
	music.AddSubMelody(ormID, &om, 1)

	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()

	track.Inputer(tuner.DataSheet{
		Inputer: &rc2,
		// HasNext: true,
	})

	go func() {
		for i := 0; i < 5; i++ {
			wc2.Write([]byte(strconv.Itoa(i) + ", Hello!"))
			// time.Sleep(time.Second * 1)
		}
		wc2.Close()
	}()

	tuning.Wait()
}

func newNamedStreamFilter(name string) (m tuning.Melody) {
	m = tuner.StreamFilter{
		Filter: func(buf []byte) []byte {
			// fmt.Println(name)
			return buf
		},
	}
	return
}

func testCommandMelody(t *testing.T, cmd *exec.Cmd, callback func(io.WriteCloser)) {
	music := tuning.NewMusic()

	pr, pw, _ := os.Pipe()
	var rc io.ReadCloser
	var wc io.WriteCloser
	rc = pr
	wc = pw

	cmdInPipe, _ := cmd.StdinPipe()
	cmdOutPipe, _ := cmd.StdoutPipe()
	cmdErrPipe, _ := cmd.StderrPipe()

	var inSf tuning.Melody
	inSf = newNamedStreamFilter("inputer")
	inSfID := music.AddMainMelody(&inSf, 1)

	var m tuning.Melody
	m = tuner.NewCommandMelody(cmd)
	mID := music.AddMainMelody(&m, 1)

	var outSf tuning.Melody
	outSf = newNamedStreamFilter("outputer")
	outSfID := music.AddMainMelody(&outSf, 1)
	var orm tuning.Melody
	orm = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			return CustomMelodyData{
				data:    string(in.Buffer),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}
	ormID, _ := music.AddSubMelody(outSfID, &orm, 1)
	var outPrinter tuning.Melody
	outPrinter = OutputerMelody{}
	music.AddSubMelody(ormID, &outPrinter, 1)

	var errSf tuning.Melody
	errSf = newNamedStreamFilter("errputer")
	errSfID := music.AddMainMelody(&errSf, 1)
	var erm tuning.Melody
	erm = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			return CustomMelodyData{
				data:    string(in.Buffer),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}
	ermID, _ := music.AddSubMelody(errSfID, &erm, 1)
	var errPrinter tuning.Melody
	// errPrinter = OutputerMelody{}
	errPrinter = ErrputerMelody{}
	music.AddSubMelody(ermID, &errPrinter, 1)

	inSfTrack, _ := music.Build(inSfID)
	tuning.RunTrack(inSfTrack).Super().Immediately()
	outSfTrack, _ := music.Build(outSfID)
	tuning.RunTrack(outSfTrack).Super().Immediately()
	errSfTrack, _ := music.Build(errSfID)
	tuning.RunTrack(errSfTrack).Super().Immediately()

	inSfTrack.Inputer(tuner.DataSheet{
		Inputer:  &rc,
		Outputer: &cmdInPipe,
	})
	outSfTrack.Inputer(tuner.DataSheet{
		Inputer: &cmdOutPipe,
	})
	errSfTrack.Inputer(tuner.DataSheet{
		Inputer: &cmdErrPipe,
	})

	mainTrack, _ := music.Build(mID)
	tuning.RunTrack(mainTrack).Super().Immediately()

	mainTrack.Inputer(tuner.SwapData{
		Sign: tuner.CmRun,
	})

	callback(wc)
}

func TestCommandMelody_ipython_withWait(t *testing.T) {
	cmd := exec.Command("ipython")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("print(0)\n"))
			time.Sleep(time.Second * 5)

			wc.Write([]byte("print 0\n"))
			time.Sleep(time.Second * 5)

			wc.Close()
		}()

		tuning.Wait()
	})
}

func TestCommandMelody_ipython_noWait(t *testing.T) {
	cmd := exec.Command("ipython")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("print(0)\n"))
			wc.Write([]byte("print 0\n"))
			wc.Write([]byte("exit\n"))
			wc.Close()
		}()

		tuning.Wait()
	})
}

func TestCommandMelody_tp_noExit(t *testing.T) {
	cmd := exec.Command("./test_prog/test_prog")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("out\n"))
			wc.Write([]byte("err\n"))
			wc.Write([]byte("out\n"))
			time.Sleep(time.Second * 5)
			wc.Close()
		}()

		tuning.Wait()
	})
}

func TestCommandMelody_tp_noWait(t *testing.T) {
	cmd := exec.Command("./test_prog/test_prog")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("out\n"))
			wc.Write([]byte("err\n"))
			wc.Write([]byte("out\n"))
			time.Sleep(time.Microsecond * 1)
			wc.Write([]byte("exit(1)\n"))
			time.Sleep(time.Second * 5)
			wc.Close()
		}()

		tuning.Wait()
	})
}

func TestCommandMelody_tp_withWait(t *testing.T) {
	cmd := exec.Command("./test_prog/test_prog")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("out\n"))
			time.Sleep(time.Second * 1)
			wc.Write([]byte("err\n"))
			time.Sleep(time.Second * 1)
			wc.Write([]byte("out\n"))
			time.Sleep(time.Second * 1)
			wc.Write([]byte("exit(1)\n"))
			time.Sleep(time.Second * 1)
			wc.Close()
		}()

		tuning.Wait()
	})
}

func TestCommandMelody_exit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "./test_prog/test_prog")
	testCommandMelody(t, cmd, func(wc io.WriteCloser) {
		go func() {
			wc.Write([]byte("out\n"))
			wc.Write([]byte("err\n"))
			wc.Write([]byte("out\n"))
			time.Sleep(time.Second * 10)
			wc.Close()
		}()
		go func() {
			time.Sleep(time.Second * 5)
			fmt.Println("call cancel()")
			cancel()
		}()

		tuning.Wait()
	})
}

func TestTuner(t *testing.T) {
	T := tuner.New()
	i0, i1, _ := T.Load("./test_prog/test_prog")(tuner.CmStreamFilter{
		OutSf: tuner.NewDefaultStreamFilter(),
		ErrSf: tuner.NewDefaultStreamFilter(),
	})

	var orm tuning.Melody
	orm = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			return CustomMelodyData{
				data:    string(in.Buffer),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}
	ormID, _ := T.GetMusic().AddSubMelody(i1.OutID, &orm, 1)
	var outPrinter tuning.Melody
	outPrinter = OutputerMelody{}
	T.GetMusic().AddSubMelody(ormID, &outPrinter, 1)

	pr, pw, _ := os.Pipe()
	var rc io.ReadCloser
	var wc io.WriteCloser
	rc = pr
	wc = pw

	T.Active(i1)
	T.Run(i0, tuner.NewDefaultStreamFilter(), func(ds *tuner.DataSheet) {
		ds.Inputer = &rc
	})

	go func() {
		wc.Write([]byte("out\n"))
		wc.Write([]byte("err\n"))
		wc.Write([]byte("out\n"))
		time.Sleep(time.Second * 3)
		wc.Close()
	}()

	tuning.Wait()
}

func TestTuner2(t *testing.T) {
	T := tuner.New()
	i0, i1, _ := T.Load("ipython")(tuner.CmStreamFilter{
		OutSf: tuner.StreamFilter{
			Filter: func(d []byte) []byte {
				// strings.FieldsFunc(str, unicode.IsSpace)
				return d
			},
		},
		ErrSf: tuner.NewDefaultStreamFilter(),
	})

	var orm tuning.Melody
	orm = tuner.OutputRectifier{
		Transfer: func(in tuner.DataSheet) interface{} {
			return CustomMelodyData{
				data:    string(in.Buffer),
				t:       t,
				hasNext: in.HasNext,
			}
		},
	}
	ormID, _ := T.GetMusic().AddSubMelody(i1.OutID, &orm, 1)
	var outPrinter tuning.Melody
	outPrinter = OutputerMelody{}
	T.GetMusic().AddSubMelody(ormID, &outPrinter, 1)

	pr, pw, _ := os.Pipe()
	var rc io.ReadCloser
	var wc io.WriteCloser
	rc = pr
	wc = pw

	T.Active(i1)
	T.Run(i0, tuner.NewDefaultStreamFilter(), func(ds *tuner.DataSheet) {
		ds.Inputer = &rc
	})

	go func() {
		wc.Write([]byte("print('Hello World!')\n"))
		time.Sleep(time.Second * 3)
		wc.Close()
	}()

	tuning.Wait()
}

// ping, ls, echo, etc...
