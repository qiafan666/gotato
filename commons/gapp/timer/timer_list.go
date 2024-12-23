package timer

// timerList 为了Timer能使用sort.Sort排序
type timerList []*Timer

// Len len
func (s timerList) Len() int {
	return len(s)
}

// Swap swap
func (s timerList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less less
func (s timerList) Less(i, j int) bool {
	switch {
	case s[i].endTs < s[j].endTs:
		return true
	case s[i].endTs > s[j].endTs:
		return false
	default:
		return s[i].id < s[j].id
	}
}
