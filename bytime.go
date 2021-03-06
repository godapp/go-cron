package cron

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
//

type byTime []*JobEntry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].NextTime.IsZero() {
		return false
	}
	if s[j].NextTime.IsZero() {
		return true
	}
	return s[i].NextTime.Before(s[j].NextTime)
}
