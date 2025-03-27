package panics

func Try(f func()) *Recovered {
	var c Catcher
	c.Try(f)
	return c.Recovered()
}
