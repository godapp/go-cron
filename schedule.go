package cron

import "time"

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// Define of schedule
//

// The Schedule describes a job's duty cycle.
type Schedule interface {
	// Return the next activation time, later than the given time.
	// NextTime is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}
