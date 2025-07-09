package main

import (
	"fmt"
	"os"

	"diagnoseQuery/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command line flags
	opts := ParseFlags()

	// Handle help and version flags
	if opts.ShouldShowHelp() {
		PrintHelp()
		return
	}

	if opts.ShouldShowVersion() {
		PrintVersion()
		return
	}

	// Handle generate sample flag
	if opts.GenerateSample {
		if err := GenerateSampleEnv(); err != nil {
			fmt.Printf("sample.env 파일 생성 실패: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("sample.env 파일이 성공적으로 생성되었습니다.")
		return
	}

	// Initialize the main TUI model with command line options
	mainModel, err := tui.NewMainModelWithOptions(opts.EnvPath, opts.EnvKeys)
	if err != nil {
		fmt.Printf("TUI 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(mainModel, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("프로그램 실행 실패: %v\n", err)
		os.Exit(1)
	}
}
