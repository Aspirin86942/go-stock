package main

import "go-stock/backend/service/task"

var _ task.Scheduler = (*appTaskScheduler)(nil)

type appTaskScheduler struct {
	app *App
}

func (s *appTaskScheduler) Register(key, spec string, job func()) error {
	if entryID, exists := s.app.getCronEntry(key); exists {
		s.app.cron.Remove(entryID)
	}

	entryID, err := s.app.cron.AddFunc(spec, job)
	if err != nil {
		return err
	}

	s.app.setCronEntry(key, entryID)
	return nil
}

func (s *appTaskScheduler) Unregister(key string) {
	if entryID, exists := s.app.getCronEntry(key); exists {
		s.app.cron.Remove(entryID)
		s.app.removeCronEntry(key)
	}
}
