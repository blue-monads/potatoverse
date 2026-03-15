package xcorn

import (
	"time"

	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func (c *CornCapability) loop() {
	for {
		dur, name := c.nextTick()

		if name == "" {
			select {
			case <-c.reload:
			case <-c.done:
				return
			}
			continue
		}

		select {
		case <-c.reload:
		case <-time.After(dur):
			job, ok := c.jobs[name]
			if ok {
				c.executeJob(job)
			}
		case <-c.done:
			return
		}
	}
}

func (c *CornCapability) nextTick() (time.Duration, string) {
	if len(c.jobs) == 0 {
		return 0, ""
	}

	var (
		shortest time.Duration = -1
		next     string
	)

	now := time.Now()
	for _, job := range c.jobs {
		var wait time.Duration
		if job.LastRun.IsZero() {
			wait = 0
		} else {
			wait = max(job.LastRun.Add(job.Interval).Sub(now), 0)
		}

		if shortest < 0 || wait < shortest {
			shortest = wait
			next = job.Name
		}
	}

	return shortest, next
}

func (c *CornCapability) executeJob(job *CornJob) {
	job.jLock.Lock()
	defer job.jLock.Unlock()

	job.LastRun = time.Now()
	job.RunCount++

	engine := c.builder.app.Engine().(xtypes.Engine)

	engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    1,
		EventType:  "corn",
		ActionName: job.Name,
		Params:     make(map[string]string),
		Request:    nil,
	})

}
