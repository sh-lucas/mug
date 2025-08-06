package main

import "github.com/sh-lucas/mug/global"

var Colors = struct {
	Reset  string
	Red    string
	Green  string
	Yellow string
	Blue   string
	Purple string
	Cyan   string
}{
	Reset:  global.Reset,
	Red:    global.Red,
	Green:  global.Green,
	Yellow: global.Yellow,
	Blue:   global.Blue,
	Purple: global.Purple,
	Cyan:   global.Cyan,
}
