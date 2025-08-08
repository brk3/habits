package habit

type Habit struct {
	Name      string `json:"name"`
	Note      string `json:"note"`
	TimeStamp int64  `json:"timestamp"`
}
