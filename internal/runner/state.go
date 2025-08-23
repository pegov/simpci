package runner

type State struct {
	m map[string]string
}

func NewState() State {
	return State{m: make(map[string]string)}
}

func (s *State) AddTarget(name, scriptPath string) {
	s.m[name] = scriptPath
}

func (s *State) Script(name string) (string, bool) {
	v, ok := s.m[name]
	return v, ok
}
