package tuning

// Melody - the payload of task
type Melody interface {
	Play(*Track, interface{}) interface{}
}

// BaseMelody - simple Impl
type BaseMelody struct {
}

// Play - impl melody
func (BaseMelody) Play(t *Track, in interface{}) interface{} {
	return in
}
