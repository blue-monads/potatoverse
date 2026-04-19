package xcorn

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

type CornJob struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	LastRun  time.Time     `json:"last_run"`
	RunCount int64         `json:"run_count"`

	jLock sync.Mutex
}

func (c *CornCapability) ListActions() ([]string, error) {
	return []string{
		"list_jobs",
		"trigger_job",
	}, nil
}

func (c *CornCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "list_jobs":
		return c.listJobs()
	case "trigger_job":
		return c.triggerJob(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *CornCapability) listJobs() (any, error) {
	type jobInfo struct {
		Name     string `json:"name"`
		Interval string `json:"interval"`
		LastRun  string `json:"last_run,omitempty"`
		RunCount int64  `json:"run_count"`
	}

	jobs := make([]jobInfo, 0, len(c.jobs))
	for _, job := range c.jobs {
		var lastRun string
		if !job.LastRun.IsZero() {
			lastRun = job.LastRun.Format(time.RFC3339)
		}
		jobs = append(jobs, jobInfo{
			Name:     job.Name,
			Interval: job.Interval.String(),
			LastRun:  lastRun,
			RunCount: job.RunCount,
		})
	}
	return jobs, nil
}

func (c *CornCapability) triggerJob(params lazydata.LazyData) (any, error) {
	name := params.GetFieldAsString("name")
	if name == "" {
		return nil, errors.New("name is required")
	}

	job, ok := c.jobs[name]

	if !ok {
		return nil, fmt.Errorf("job not found: %s", name)
	}

	c.executeJob(job)
	return map[string]any{"triggered": name}, nil
}
