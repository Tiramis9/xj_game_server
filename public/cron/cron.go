package cron

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"sync"
)

// Crontab crontab manager
type Job struct {
	inner *cron.Cron
	ids   map[string]cron.EntryID
	mutex sync.Mutex
}

// NewCrontab new crontab
func NewJob() *Job {
	return &Job{
		inner: cron.New(),
		ids:   make(map[string]cron.EntryID),
	}
}

// IDs ...
func (c *Job) IDs() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	validIDs := make([]string, 0, len(c.ids))
	invalidIDs := make([]string, 0)
	for sid, eid := range c.ids {
		if e := c.inner.Entry(eid); e.ID != eid {
			invalidIDs = append(invalidIDs, sid)
			continue
		}
		validIDs = append(validIDs, sid)
	}
	for _, id := range invalidIDs {
		delete(c.ids, id)
	}
	return validIDs
}

// Start start the crontab engine
func (c *Job) Start() {
	c.inner.Start()
}

// Stop stop the crontab engine
func (c *Job) Stop() {
	c.inner.Stop()
}

// DelByID remove one crontab task
func (c *Job) DelByID(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	eid, ok := c.ids[id]
	if !ok {
		return
	}
	c.inner.Remove(eid)
	delete(c.ids, id)
}

// AddByID add one crontab task
// id is unique
// spec is the crontab expression
func (c *Job) AddByID(id string, spec string, cmd cron.Job) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.ids[id]; ok {
		return errors.Errorf("crontab id exists")
	}
	eid, err := c.inner.AddJob(spec, cmd)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

// AddByFunc add function as crontab task
func (c *Job) AddByFunc(id string, spec string, f func()) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.ids[id]; ok {
		return errors.Errorf("crontab id exists")
	}
	eid, err := c.inner.AddFunc(spec, f)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

// IsExists check the crontab task whether existed with job id
func (c *Job) IsExists(jid string) bool {
	_, exist := c.ids[jid]
	return exist
}
