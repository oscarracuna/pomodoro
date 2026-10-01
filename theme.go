package main

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme is a full palette for the timer. Bg fills the terminal, Surface is
// the card, and Focus, Short, and Long color the active phase.
type Theme struct {
	Name    string
	Bg      color.Color
	Surface color.Color
	Fg      color.Color
	Muted   color.Color
	Accent  color.Color
	Focus   color.Color
	Short   color.Color
	Long    color.Color
	Border  color.Color
	Track   color.Color
}

func (t Theme) phaseColor(p phase) color.Color {
	switch p {
	case phaseShort:
		return t.Short
	case phaseLong:
		return t.Long
	default:
		return t.Focus
	}
}

func hex(s string) color.Color {
	return lipgloss.Color(s)
}

var themes = []Theme{
	{
		Name:    "Mocha",
		Bg:      hex("#1E1E2E"),
		Surface: hex("#313244"),
		Fg:      hex("#CDD6F4"),
		Muted:   hex("#A6ADC8"),
		Accent:  hex("#CBA6F7"),
		Focus:   hex("#A6E3A1"),
		Short:   hex("#FAB387"),
		Long:    hex("#89B4FA"),
		Border:  hex("#585B70"),
		Track:   hex("#45475A"),
	},
	{
		Name:    "Latte",
		Bg:      hex("#EFF1F5"),
		Surface: hex("#FFFFFF"),
		Fg:      hex("#4C4F69"),
		Muted:   hex("#6C6F85"),
		Accent:  hex("#8839EF"),
		Focus:   hex("#40A02B"),
		Short:   hex("#FE640B"),
		Long:    hex("#1E66F5"),
		Border:  hex("#CCD0DA"),
		Track:   hex("#ACB0BE"),
	},
	{
		Name:    "Nord",
		Bg:      hex("#2E3440"),
		Surface: hex("#3B4252"),
		Fg:      hex("#ECEFF4"),
		Muted:   hex("#A7B1C5"),
		Accent:  hex("#88C0D0"),
		Focus:   hex("#A3BE8C"),
		Short:   hex("#EBCB8B"),
		Long:    hex("#B48EAD"),
		Border:  hex("#4C566A"),
		Track:   hex("#434C5E"),
	},
	{
		Name:    "Gruvbox",
		Bg:      hex("#282828"),
		Surface: hex("#3C3836"),
		Fg:      hex("#EBDBB2"),
		Muted:   hex("#A89984"),
		Accent:  hex("#FABD2F"),
		Focus:   hex("#B8BB26"),
		Short:   hex("#FE8019"),
		Long:    hex("#83A598"),
		Border:  hex("#665C54"),
		Track:   hex("#504945"),
	},
	{
		Name:    "Rosé Pine",
		Bg:      hex("#191724"),
		Surface: hex("#1F1D2E"),
		Fg:      hex("#E0DEF4"),
		Muted:   hex("#908CAA"),
		Accent:  hex("#C4A7E7"),
		Focus:   hex("#9CCFD8"),
		Short:   hex("#F6C177"),
		Long:    hex("#EBBCBA"),
		Border:  hex("#524F67"),
		Track:   hex("#403D52"),
	},
	{
		Name:    "Tokyo Night",
		Bg:      hex("#1A1B26"),
		Surface: hex("#24283B"),
		Fg:      hex("#C0CAF5"),
		Muted:   hex("#737AA2"),
		Accent:  hex("#BB9AF7"),
		Focus:   hex("#9ECE6A"),
		Short:   hex("#E0AF68"),
		Long:    hex("#7AA2F7"),
		Border:  hex("#3B4261"),
		Track:   hex("#414868"),
	},
	{
		Name:    "Dracula",
		Bg:      hex("#282A36"),
		Surface: hex("#343746"),
		Fg:      hex("#F8F8F2"),
		Muted:   hex("#9AA0C3"),
		Accent:  hex("#BD93F9"),
		Focus:   hex("#50FA7B"),
		Short:   hex("#FFB86C"),
		Long:    hex("#8BE9FD"),
		Border:  hex("#6272A4"),
		Track:   hex("#44475A"),
	},
	{
		Name:    "Charm",
		Bg:      hex("#120E16"),
		Surface: hex("#1E1824"),
		Fg:      hex("#F7F3F8"),
		Muted:   hex("#A898B0"),
		Accent:  hex("#FF4FD8"),
		Focus:   hex("#7D56F4"),
		Short:   hex("#5EEAD4"),
		Long:    hex("#ECFD66"),
		Border:  hex("#4A3F58"),
		Track:   hex("#3D3348"),
	},
}
