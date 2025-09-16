package xtypes

import (
	"io/fs"
	"log/slog"
)

type BuilderOption struct {
	App App

	Logger *slog.Logger
}

type Builder func(opt BuilderOption) (*Defination, error)

type Defination struct {
	Name            string
	Slug            string
	Info            string
	Icon            string
	Version         string
	AssetData       fs.FS
	AssetDataPrefix string

	LinkPattern string

	OnInit       func(sid int64) error
	IsInitilized func(sid int64) (bool, error)
	OnDeInit     func(sid int64) error
	OnClose      func() error
}
