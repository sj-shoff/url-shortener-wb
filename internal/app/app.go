package app

import (
	"url-shortener-wb/internal/config"

	"github.com/wb-go/wbf/zlog"
)

type App struct {
}

func NewApp(cfg *config.Config) (*App, error) {
	zlog.Init()
	log := zlog.Logger

	log.Info().Msg("New application")
	panic("...")
}

func (a *App) Run() error {
	panic("...")
}
