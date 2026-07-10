package ui

type PanelOptions struct {
	Title        string
	ConfigPath   string
	CountdownSec float64
	Store        *Store
	Render       func(store *Store)
	OnRun        func(store *Store)
	OnClose      func(store *Store)
}

type ShellOptions struct {
	Title            string
	ConfigPath       string
	CountdownSec     float64
	Store            *Store
	Render           func(store *Store)
	Controller       Controller
	OpenPanelOnStart bool
}
