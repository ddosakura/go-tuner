package test

import (
	"fmt"
	"math"
	"testing"
	"time"

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
	time.Sleep(time.Millisecond * 100)
	fmt.Println(data.a, "^",
		data.n, " = ",
		data.ans)
	t.Finish()
	return data
}

func testTuning1(t *testing.T) {
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
}

func TestTuning_1(t *testing.T) {
	testTuning1(t)
	tuning.Wait()
}

type AdderMelody struct{}

func (AdderMelody) Play(t *tuning.Track, in interface{}) interface{} {
	data := in.([]int)
	if len(data) == 1 {
		fmt.Println("ans =", data[0])
		t.Finish()
		return nil
	}
	data = append(data[2:], data[0]+data[1])
	fmt.Println("data =", data)
	return data
}

func testTuning2(t *testing.T) {
	music := tuning.NewMusic()
	var m tuning.Melody
	m = AdderMelody{}
	id := music.AddMainMelody(&m, 1)
	music.ConnectMelody(id, id)
	track, _ := music.Build(id)
	tuning.RunTrack(track).Super().Immediately()
	track.Inputer([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
}

func TestTuning_2(t *testing.T) {
	tuning.GetCantor().MaxTrack = 1
	testTuning2(t)
	tuning.Wait()
}

func TestTuning_re_1(t *testing.T) {
	testTuning1(t)
	tuning.Wait()
}

func TestTuning_1and2(t *testing.T) {
	tuning.GetCantor().MaxTrack = 2
	fmt.Println("CALL 1")
	testTuning1(t)
	time.Sleep(time.Millisecond * 10)
	fmt.Println("CALL 2")
	testTuning2(t)
	tuning.Wait()
}
