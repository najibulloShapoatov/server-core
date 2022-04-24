package scheduler

import (
	"errors"
	"github.com/najibulloShapoatov/server-core/cluster"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"github.com/robfig/cron/v3"
)

type Task struct {
	Name     string
	Spec     string
	MaxRetry int
	Job      ScheduleFunc
	Cluster  *cluster.Cluster

	entryID cron.EntryID
}

type ScheduleFunc func() error

var scheduler = newCron()

func newCron() *cron.Cron {
	c := cron.New(cron.WithSeconds())
	c.Start()

	return c
}

// RegisterJob registers a new job
func RegisterJob(task *Task) error {
	if task == nil {
		return errors.New("please define a task")
	}

	job := func() {
		err := runJob(task)
		if err != nil {
			log.Error(task.Name, err.Error())
		} else {
			log.Info("job success", task.Name)
		}
	}
	entryID, err := scheduler.AddFunc(task.Spec, job)
	task.entryID = entryID
	return err
}

// UnregisterJob unregisters a job
func UnregisterJob(task *Task) error {
	if task == nil {
		return errors.New("please define a task")
	}

	scheduler.Remove(task.entryID)
	return nil
}

func runJob(task *Task) error {
	if task.Cluster == nil {
		c, err := cluster.Join("scheduler")
		if err != nil {
			return runWithRetry(task, task.MaxRetry)
		} else {
			task.Cluster = c
			return runOnCluster(task)
		}
	} else {
		return runOnCluster(task)
	}
}

func runWithRetry(task *Task, attempts int) error {
	if err := task.Job(); err != nil {
		if attempts--; attempts > 0 {
			return runWithRetry(task, attempts)
		}
		return err
	}
	return nil
}

func runOnCluster(task *Task) error {
	err := task.Cluster.Lock(task.Name)

	if err == nil {
		err = runWithRetry(task, task.MaxRetry)
		_ = task.Cluster.Unlock(task.Name)
		return err
	}
	return nil
}
