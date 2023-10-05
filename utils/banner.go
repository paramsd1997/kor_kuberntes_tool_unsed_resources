package kor

import (
	"github.com/fatih/color"
)

func PrintLogo() {
	boldMagenta := color.New(color.FgHiMagenta, color.Bold)
	asciiLogo := `
	  _  _____  ____
	 | |/ / _ \|  _ \
	 | ' / | | | |_) |
	 | . \ |_| |  _ <
	 |_|\_\___/|_| \_\
	`

	boldMagenta.Println(asciiLogo)
}
