package tuner

import (
	"bufio"
	"io"

	"github.com/ddosakura/go-tuner/tuning"
)

type DataSheet struct {
	Buffer   []byte
	HasNext  bool
	Inputer  *io.ReadCloser
	Outputer *io.WriteCloser
}

const defaultBufLen = 1024

type StreamFilterMode int

const (
	_ StreamFilterMode = iota
	Unknow
	Reader2Writer
)

type StreamFilter struct {
	Filter func([]byte) []byte
	BufLen int
}

func (sf StreamFilter) Play(t *tuning.Track, in interface{}) interface{} {
	if sf.BufLen < 1 {
		sf.BufLen = defaultBufLen
	}
	// fmt.Println("with bufLen =", sf.BufLen)

	ds := in.(DataSheet)
	var buffer []byte
	if ds.Inputer == nil {
		buffer = ds.Buffer
		var orNeedClose = false
		if !ds.HasNext {
			orNeedClose = true
			t.Finish()
		}

		return sf.callFilter(ds, buffer, orNeedClose)
	}

	buffer = make([]byte, sf.BufLen)

	// n, err := (*ds.Inputer).Read(buffer)
	// 创建带缓冲的Reader
	var bufReader = bufio.NewReader(*ds.Inputer)
	n, err := bufReader.Read(buffer)

	if err != nil {
		// 管道关闭后会出现 io.EOF 错误
		// if err == io.EOF {
		t.Finish()
		if ds.Outputer != nil {
			(*ds.Outputer).Close()
			return nil
		}
		return DataSheet{
			Buffer:  make([]byte, 0),
			HasNext: false,
		}
		// }
		// panic(err)
		// panic("StreamFilter Read Error!")
	}
	buffer = buffer[:n]

	ret := sf.callFilter(ds, buffer, false)

	t.Inputer(DataSheet{
		Inputer:  ds.Inputer,
		Outputer: ds.Outputer,
		HasNext:  true,
	})
	return ret
}

func (sf StreamFilter) callFilter(ds DataSheet, buffer []byte, orNeedClose bool) interface{} {
	buffer = sf.Filter(buffer)

	if ds.Outputer == nil {
		var hasNext = ds.HasNext
		if ds.Inputer != nil {
			hasNext = true
		}
		return DataSheet{
			Buffer:  buffer,
			HasNext: hasNext,
		}
	}

	// TODO: 尾保留机制
	(*ds.Outputer).Write(buffer)
	if orNeedClose {
		(*ds.Outputer).Close()
	}
	return nil
}

// NewDefaultStreamFilter - direct transfer
func NewDefaultStreamFilter() (m tuning.Melody) {
	m = StreamFilter{
		Filter: func(buf []byte) []byte {
			return buf
		},
	}
	return
}
