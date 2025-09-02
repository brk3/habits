package nudge

import "fmt"

func Nudge(email string, hours int) {
	fmt.Print("nudge called with email:", email, " hours:", hours, "\n")
}
