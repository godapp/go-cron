package cron

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Cron struct {
	entries     []*JobEntry
	stop        chan struct{}
	add         chan *JobEntry
	snapshot    chan []*JobEntry
	remove      chan string
	running     bool
	ErrorLogger *log.Logger
	location    *time.Location
	mux         *sync.RWMutex
}

// JobEntry consists of a schedule and the func to execute on that schedule.
type JobEntry struct {
	// The schedule on which this job should be run.
	Schedule Schedule

	// The next time the job will run. This is the zero time if Cron has not been
	// started or this entry's schedule is unsatisfiable
	NextTime time.Time

	// The last time this job was run. This is the zero time if the job has never
	// been run.
	PrevTime time.Time

	// The Job to run.
	Job Job2

	// The Job's name
	Name string

	// Count of runs
	Count int
}

// New returns a new Cron job runner, in the Local time zone.
func New() *Cron {
	return NewDefault()
}

// NewDefault returns a new Cron job runner, in the Local time zone.
func NewDefault() *Cron {
	return NewWithLocation(time.Now().Location())
}

// NewWithLocation returns a new Cron job runner.
func NewWithLocation(location *time.Location) *Cron {
	return &Cron{
		entries:     nil,
		add:         make(chan *JobEntry, 1),
		stop:        make(chan struct{}),
		snapshot:    make(chan []*JobEntry),
		remove:      make(chan string, 1),
		running:     false,
		ErrorLogger: nil,
		location:    location,
		mux:         new(sync.RWMutex),
	}
}

// AddJob adds a Job to the Cron to be run on the given schedule and name.
// Return error if failed to parse spec, otherwise nil
func (c *Cron) AddJob(spec string, job Job, names ...string) error {
	return c.AddJob2(spec, &Job2To1{job: job}, names...)
}

// AddJob adds a Job2 to the Cron to be run on the given schedule and name.
// Return error if failed to parse spec, otherwise nil
func (c *Cron) AddJob2(spec string, job Job2, names ...string) error {
	schedule, err := Parse(spec)
	if err != nil {
		return err
	}
	c.Schedule2(schedule, job, c.makeName(names))
	return nil
}

// Remove an entry from being run in the future.
func (c *Cron) Remove(name string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.running {
		idx := c.indexByName(name)
		if idx == -1 {
			return
		}
		c.entries = c.entries[:idx+copy(c.entries[idx:], c.entries[idx+1:])]
	}
	c.remove <- name
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *Cron) Schedule(schedule Schedule, job Job, names ...string) {
	c.Schedule2(schedule, &Job2To1{job: job}, names...)
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *Cron) Schedule2(schedule Schedule, job Job2, names ...string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	entry := &JobEntry{
		Schedule: schedule,
		Job:      job,
		Name:     c.makeName(names),
	}

	if !c.running {
		p := c.indexByName(entry.Name)
		if p != -1 {
			c.logf("Duplicate names not allowed")
		}

		c.entries = append(c.entries, entry)
		return
	}

	c.add <- entry
}

// Entries returns a snapshot of the cron entries.
func (c *Cron) Entries() []*JobEntry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

// Location gets the time zone location
func (c *Cron) Location() *time.Location {
	return c.location
}

// Start the cron scheduler in its own go-routine, or no-op if already started.
func (c *Cron) Start() {
	if c.running {
		return
	}
	c.running = true
	go c.scheduleJobs()
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
func (c *Cron) Stop() {
	if !c.running {
		return
	}
	c.stop <- struct{}{}
	c.running = false
}

////

func (c *Cron) indexByName(name string) int {
	for p, e := range c.entries {
		if e.Name == name {
			return p
		}
	}
	return -1
}

func (c *Cron) invoke(entry *JobEntry) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			c.logf("cron: panic running job: %v\n%s", r, buf)
		}
	}()
	entry.Count += 1
	entry.Job.Run(&JobContext{
		Name:  entry.Name,
		Count: entry.Count,
	})
}

func (c *Cron) scheduleJobs() {
	// Figure out the next activation times for each entry.
	now := c.now()
	for _, entry := range c.entries {
		entry.NextTime = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].NextTime.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			timer = time.NewTimer(c.entries[0].NextTime.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				now = now.In(c.location)
				// Run every entry whose next time was less than now
				for _, e := range c.entries {
					if e.NextTime.After(now) || e.NextTime.IsZero() {
						break
					}
					go c.invoke(e)
					e.PrevTime = e.NextTime
					e.NextTime = e.Schedule.Next(now)
				}

			case newEntry := <-c.add:
				timer.Stop()
				now = c.now()
				newEntry.NextTime = newEntry.Schedule.Next(now)
				c.entries = append(c.entries, newEntry)

			case name := <-c.remove:
				p := c.indexByName(name)
				if p == -1 {
					break
				}

				c.entries = c.entries[:p+copy(c.entries[p:], c.entries[p+1:])]

			case <-c.snapshot:
				c.snapshot <- c.entrySnapshot()
				continue

			case <-c.stop:
				timer.Stop()
				return
			}

			break
		}
	}
}

func (c *Cron) logf(format string, args ...interface{}) {
	if c.ErrorLogger != nil {
		c.ErrorLogger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *Cron) entrySnapshot() []*JobEntry {
	entries := make([]*JobEntry, len(c.entries))
	for i, e := range c.entries {
		entries[i] = &JobEntry{
			Schedule: e.Schedule,
			NextTime: e.NextTime,
			PrevTime: e.PrevTime,
			Job:      e.Job,
		}
	}
	return entries
}

// now returns current time in c location
func (c *Cron) now() time.Time {
	return time.Now().In(c.location)
}

func (c *Cron) makeName(names []string) string {
	if len(names) <= 0 {
		return fmt.Sprintf("%d", time.Now().Unix())
	} else {
		return names[0]
	}
}
