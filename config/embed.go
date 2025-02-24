package config

import (
	"embed"
)

//go:embed checkItems/baseline_config_linux.yaml
var BaseLineConfigLinux embed.FS

//go:embed checkItems/baseline_config_windows.yaml
var BaseLineConfigWindows embed.FS

//go:embed checkItems/*
var StaticFiles embed.FS

//go:embed html/index.html
var Index embed.FS

//go:embed html/results.html
var Results embed.FS

//go:embed html/*
var Content embed.FS
