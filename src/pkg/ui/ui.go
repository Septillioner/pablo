package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	WarnColor    = color.New(color.FgYellow).SprintFunc()
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	ActionColor  = color.New(color.FgWhite).SprintFunc()
	BoldColor    = color.New(color.Bold).SprintFunc()
	GrayColor    = color.New(color.FgHiBlack).SprintFunc()
	ThemeColor   = color.New(color.FgMagenta).SprintFunc()
)

// Header prints the ASCII art title for Pablo
func Header() {
	pabloArt := `    ┓ ┓  
┏┓┏┓┣┓┃┏┓
┣┛┗┻┗┛┗┗┛
┛        
`
	fmt.Println(ThemeColor(pabloArt))
	fmt.Println(GrayColor(strings.Repeat("=", 40)))
	fmt.Println(ThemeColor(BoldColor("  PUBLISH HELPER - v1.1.0")))
	fmt.Println(GrayColor("  Author: egeismailkosedag@gmail.com"))
	fmt.Println(GrayColor("  Github: github.com/septillioner"))
	fmt.Println(GrayColor(strings.Repeat("=", 40)))
	fmt.Println()
}

// Log prints a structured log message with a status mark and timestamp
func Log(mark string, message string) {
	timestamp := time.Now().Format("15:04:05")
	var formattedMark string

	switch mark {
	case "+":
		formattedMark = SuccessColor("[+]")
	case "-":
		formattedMark = ErrorColor("[-]")
	case "!":
		formattedMark = WarnColor("[!]")
	case "*":
		formattedMark = InfoColor("[*]")
	case ">":
		formattedMark = ActionColor("[>]")
	default:
		formattedMark = fmt.Sprintf("[%s]", mark)
	}

	fmt.Printf("%s %s %s\n", GrayColor(timestamp), formattedMark, message)
}

// Section prints a titled section divider
func Section(title string) {
	fmt.Println()
	fmt.Println(ThemeColor(BoldColor(strings.ToUpper(title))))
	fmt.Println(GrayColor(strings.Repeat("-", 40)))
}

// Result prints the final outcome of an operation
func Result(success bool, duration time.Duration) {
	fmt.Println(GrayColor(strings.Repeat("=", 40)))
	if success {
		fmt.Printf("%s %s (Duration: %v)\n", SuccessColor("RESULT:"), BoldColor("SUCCESS"), duration)
	} else {
		fmt.Printf("%s %s (Duration: %v)\n", ErrorColor("RESULT:"), BoldColor("FAILED"), duration)
	}
	fmt.Println(GrayColor(strings.Repeat("=", 40)))
}

// ProgressBar prints an ASCII progress bar
func ProgressBar(percent int, label string) {
	width := 20
	filled := int(float64(percent) / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	fmt.Printf("\r%s [%s] %d%%", label, bar, percent)
	if percent >= 100 {
		fmt.Println()
	}
}
