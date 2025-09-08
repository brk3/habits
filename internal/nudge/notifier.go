package nudge

type Notifier interface {
	SendNudge(habits []string, hoursTillExpiry int) error
}
