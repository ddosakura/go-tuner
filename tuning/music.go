package tuning

import (
	"github.com/satori/go.uuid"
)

// TrackChan - track chan
type TrackChan struct {
	melody *Melody
	bufLen int
	next   []*TrackChan
}

// Music - the factory of task
type Music struct {
	tc map[string]*TrackChan
}

// NewMusic - new factory
func NewMusic() *Music {
	return &Music{
		tc: map[string]*TrackChan{},
	}
}

// AddMainMelody - add melody
func (m *Music) AddMainMelody(melody *Melody, bufLen int) (id string) {
	id = uuid.Must(uuid.NewV4()).String()
	m.tc[id] = &TrackChan{
		melody: melody,
		next:   []*TrackChan{},
		bufLen: bufLen,
	}
	return
}

// AddSubMelody - add & connect melody
// Warning: not equal to `ConnectMelody(mainID, AddMainMelody(melody))`
func (m *Music) AddSubMelody(mainID string, melody *Melody, bufLen int) (id string, ok bool) {
	_, ok = m.tc[mainID]
	if ok {
		id = uuid.Must(uuid.NewV4()).String()
		m.tc[mainID].next = append(m.tc[mainID].next, &TrackChan{
			melody: melody,
			next:   []*TrackChan{},
			bufLen: bufLen,
		})
	}
	return
}

// ConnectMelody - connect melody
func (m *Music) ConnectMelody(m1 string, m2 string) (ok bool) {
	_, ok1 := m.tc[m1]
	_, ok2 := m.tc[m2]
	ok = ok1 && ok2
	if ok {
		m.tc[m1].next = append(m.tc[m1].next, m.tc[m2])
	}
	return
}

// Get - get main melody
func (m *Music) Get(id string) (tc *TrackChan, ok bool) {
	tc, ok = m.tc[id]
	return
}

// BuildChan - build task chan
func (m *Music) BuildChan(tc *TrackChan) (t *Track, total int) {
	t = NewTrack(tc.melody, tc.bufLen)
	total = 1
	for i := range tc.next {
		sub, sum := m.BuildChan(tc.next[i])
		t.Before(sub)
		total += sum
	}
	return
}

// Build - build task chan by id
func (m *Music) Build(id string) (t *Track, total int) {
	tc, ok := m.Get(id)
	if ok {
		t, total = m.BuildChan(tc)
	}
	return
}
