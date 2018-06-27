package cron

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// Func Job supports for cron
//

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddFunc(spec string, funcJob func(), names ...string) error {
	name := c.makeName(names)
	return c.AddJob(spec, JobWrapper(funcJob), name)
}

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddOnceFunc(spec string, funcJob func(), names ...string) error {
	name := c.makeName(names)
	// Support Once run function.
	return c.AddJob(spec, JobWrapper(func() {
		defer c.Remove(name)
		funcJob()
	}), name)
}

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddFunc2(spec string, funcJob func(ctx *JobContext), names ...string) error {
	name := c.makeName(names)
	return c.AddJob2(spec, Job2Wrapper(funcJob), name)
}

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddOnceFunc2(spec string, funcJob func(ctx *JobContext), names ...string) error {
	name := c.makeName(names)
	// Support Once run function.
	return c.AddJob2(spec, Job2Wrapper(func(ctx *JobContext) {
		defer c.Remove(name)
		funcJob(ctx)
	}), name)
}
