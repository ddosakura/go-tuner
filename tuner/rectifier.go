package tuner

import (
	"github.com/ddosakura/go-tuner/tuning"
)

type InputRectifier struct {
	Transfer func(interface{}) DataSheet
}

func (ir InputRectifier) Play(t *tuning.Track, in interface{}) (ds interface{}) {
	ds = ir.Transfer(in)
	if !ds.(DataSheet).HasNext {
		// fmt.Println("inputer finish!")
		t.Finish()
	}
	return
}

type OutputRectifier struct {
	Transfer func(DataSheet) interface{}
}

func (or OutputRectifier) Play(t *tuning.Track, in interface{}) interface{} {
	if !in.(DataSheet).HasNext {
		// fmt.Println("outputer finish!")
		t.Finish()
	}
	return or.Transfer(in.(DataSheet))
}
