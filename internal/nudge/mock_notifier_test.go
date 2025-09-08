package nudge

type mockNotifier struct {
	called    bool
	habits    []string
	threshold int
	err       error
}

func (m *mockNotifier) SendNudge(habits []string, hoursTillExpiry int) error {
	m.called = true
	m.habits = habits
	m.threshold = hoursTillExpiry
	return m.err
}
