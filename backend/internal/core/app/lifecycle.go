package app

import (
	"context"
	"time"
)

func (a *App) Run() error { return a.server.Run() }

func (a *App) Close() error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if a.server != nil {
		if err := a.server.Shutdown(shutdownContext); err != nil {
			return err
		}
	}
	if a.Log != nil {
		if err := a.Log.Close(); err != nil {
			return err
		}
	}
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
