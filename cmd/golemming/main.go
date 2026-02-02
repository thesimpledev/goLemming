// GoLemming - Autonomous desktop automation agent
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thesimpledev/golemming/internal/action"
	"github.com/thesimpledev/golemming/internal/agent"
	"github.com/thesimpledev/golemming/internal/config"
	"github.com/thesimpledev/golemming/internal/ui"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// Version is set by ldflags during build
var Version = "dev"

func main() {
	checkPlatform()

	// Parse flags
	goal := flag.String("goal", "", "Goal to accomplish (runs in headless mode)")
	headless := flag.Bool("headless", false, "Run in headless mode without TUI")
	maxIter := flag.Int("max-iterations", 100, "Maximum number of iterations")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("golemming %s\n", Version)
		return
	}

	// If goal is provided, run in headless mode
	if *goal != "" || *headless {
		if *goal == "" {
			fmt.Fprintln(os.Stderr, "Error: -goal is required in headless mode")
			os.Exit(1)
		}
		runHeadless(*goal, *maxIter)
		return
	}

	// Run interactive TUI
	p := tea.NewProgram(
		ui.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runHeadless runs the agent in headless mode (for scripting/automation).
func runHeadless(goal string, maxIter int) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if maxIter > 0 {
		cfg.MaxIterations = maxIter
	}

	// Create agent
	ag := agent.New(cfg)

	// Set up action callback for logging
	ag.OnAction(func(act *protocol.Action, result *action.Result) {
		timestamp := time.Now().Format("15:04:05")
		status := "OK"
		if !result.Success {
			status = "ERROR: " + result.Error
		}
		fmt.Printf("[%s] %s: %s [%s]\n", timestamp, act.Type, formatAction(act), status)
	})

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived interrupt, stopping agent...")
		cancel()
	}()

	// Run agent
	fmt.Printf("Starting agent with goal: %s\n", goal)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("---")

	err = ag.Run(ctx, goal)

	fmt.Println("---")

	// Print result
	state := ag.State()
	switch state {
	case agent.StateCompleted:
		fmt.Printf("Completed: %s\n", ag.Result())
	case agent.StateFailed:
		fmt.Printf("Failed: %s\n", ag.Result())
		os.Exit(1)
	case agent.StateStopped:
		fmt.Println("Stopped by user")
	default:
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Print history summary
	history := ag.History()
	fmt.Printf("\nTotal actions: %d\n", history.Len())
}

func formatAction(act *protocol.Action) string {
	switch act.Type {
	case protocol.ActionClick:
		btn := act.Button
		if btn == "" {
			btn = "left"
		}
		dbl := ""
		if act.Double {
			dbl = " double"
		}
		return fmt.Sprintf("%s%s at (%d, %d)", btn, dbl, act.X, act.Y)
	case protocol.ActionType_:
		text := act.Text
		if len(text) > 30 {
			text = text[:30] + "..."
		}
		return fmt.Sprintf("%q", text)
	case protocol.ActionKey:
		return act.Key
	case protocol.ActionScroll:
		return fmt.Sprintf("%s %d", act.Direction, act.Amount)
	case protocol.ActionFileRead:
		return act.Path
	case protocol.ActionFileWrite:
		return act.Path
	case protocol.ActionWait:
		return fmt.Sprintf("%dms", act.Ms)
	default:
		return ""
	}
}
