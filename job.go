package cron

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// Job defines of cron
//

// Context of job
type JobContext struct {
	Name  string
	Count int
}

// Job is an interface for submitted cron jobs.
type Job interface {
	Run()
}

// Job is an interface for submitted cron jobs.
type Job2 interface {
	Run(ctx *JobContext)
}

// A wrapper that turns a func() into a cron.Job
type JobWrapper func()

func (fun JobWrapper) Run() { fun() }

// A wrapper that turns a func(ctx *JobContext) into a cron.Job
type Job2Wrapper func(ctx *JobContext)

func (fun Job2Wrapper) Run(ctx *JobContext) { fun(ctx) }

// A wrapper that turns a func(ctx *JobContext) into a cron.Job
type Job2To1 struct {
	job Job
}

func (fun Job2To1) Run(ctx *JobContext) { fun.job.Run() }
