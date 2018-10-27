package test

import (
	"fmt"
	"math"
	"testing"

	"github.com/ddosakura/go-tuner/tuning"
)

type PowerData struct {
	name string
	a    float64
	n    float64
	ans  float64
}

type PowerMelody struct {
	name string
}

func (PowerMelody) Play(t *tuning.Track, pd interface{}) interface{} {
	data := pd.(PowerData)
	data.name = "PowerMelody"
	data.ans = math.Pow(data.a, data.n)
	t.Finish()
	return data
}

type PrinterMelody struct {
	name string
}

func (PrinterMelody) Play(t *tuning.Track, pd interface{}) interface{} {
	if pd == nil {
		fmt.Println("error!")
		return pd
	}

	data := pd.(PowerData)
	data.name = "PrinterMelody"
	fmt.Println(data.a, "^",
		data.n, " = ",
		data.ans)
	t.Finish()
	return data
}

func TestTuning(t *testing.T) {
	music := tuning.NewMusic()
	var m1, m2 tuning.Melody
	m1 = PowerMelody{
		name: "powerFunc",
	}
	id := music.AddMainMelody(&m1, 1)
	m2 = PrinterMelody{
		name: "printer",
	}
	music.AddSubMelody(id, &m2, 1)
	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()
	track.Inputer(PowerData{
		name: "data",
		a:    2,
		n:    20,
	})
	tuning.Wait()
}
