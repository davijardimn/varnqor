package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/davijardimn/varnqor/internal/settings"
)

func runConfigCLI(args []string) (bool, int) {
	if len(args) == 0 || args[0] == "show" {
		value, err := settings.LoadDefault()
		if err != nil { fmt.Fprintln(os.Stderr, "load settings:", err); return true, 1 }
		if wantsJSON(args) { printJSON(value) } else { printConfig(value) }
		return true, 0
	}

	value, err := settings.LoadDefault()
	if err != nil { fmt.Fprintln(os.Stderr, "load settings:", err); return true, 1 }

	switch args[0] {
	case "refresh":
		if len(args) != 2 { fmt.Fprintln(os.Stderr, "usage: varnqor config refresh <2-60 seconds>"); return true, 2 }
		seconds, err := strconv.Atoi(args[1])
		if err != nil { fmt.Fprintln(os.Stderr, "refresh interval must be an integer"); return true, 2 }
		value.RefreshSeconds = seconds
	case "language":
		if len(args) != 2 { fmt.Fprintln(os.Stderr, "usage: varnqor config language <system|en|pt-BR|es|de|fr>"); return true, 2 }
		value.Language = args[1]
	case "theme":
		if len(args) != 2 { fmt.Fprintln(os.Stderr, "usage: varnqor config theme <system|light|dark>"); return true, 2 }
		value.Theme = strings.ToLower(args[1])
	case "budget":
		if len(args) < 2 { fmt.Fprintln(os.Stderr, "usage: varnqor config budget <tokens|off> [--reset-hour 0-23]"); return true, 2 }
		if args[1] == "off" {
			value.Budget.Enabled = false
		} else {
			amount, err := settings.ParseTokenAmount(args[1])
			if err != nil { fmt.Fprintln(os.Stderr, "budget:", err); return true, 2 }
			value.Budget.Enabled = true
			value.Budget.LimitTokens = amount
			for index := 2; index < len(args); index++ {
				if args[index] != "--reset-hour" || index+1 >= len(args) { fmt.Fprintln(os.Stderr, "usage: varnqor config budget <tokens|off> [--reset-hour 0-23]"); return true, 2 }
				hour, err := strconv.Atoi(args[index+1])
				if err != nil { fmt.Fprintln(os.Stderr, "reset hour must be an integer from 0 to 23"); return true, 2 }
				value.Budget.ResetHour = hour
				index++
			}
		}
	case "optimize":
		if len(args) < 2 { fmt.Fprintln(os.Stderr, "usage: varnqor config optimize <balanced|economy|throughput|off> [--target 5-60] [--horizon 30-1440] [--mode preview|manual]"); return true, 2 }
		if strings.EqualFold(args[1], "off") {
			value.Optimization.Enabled = false
		} else {
			value.Optimization.Enabled = true
			value.Optimization.Goal = strings.ToLower(args[1])
			for index := 2; index < len(args); index++ {
				switch args[index] {
				case "--target":
					if index+1 >= len(args) { fmt.Fprintln(os.Stderr, "optimize: --target requires a value"); return true, 2 }
					parsed, err := strconv.ParseFloat(args[index+1], 64); if err != nil { fmt.Fprintln(os.Stderr, "optimize: target must be numeric"); return true, 2 }; value.Optimization.TargetReductionPercent = parsed; index++
				case "--horizon":
					if index+1 >= len(args) { fmt.Fprintln(os.Stderr, "optimize: --horizon requires minutes"); return true, 2 }
					parsed, err := strconv.Atoi(args[index+1]); if err != nil { fmt.Fprintln(os.Stderr, "optimize: horizon must be integer minutes"); return true, 2 }; value.Optimization.HorizonMinutes = parsed; index++
				case "--mode":
					if index+1 >= len(args) { fmt.Fprintln(os.Stderr, "optimize: --mode requires preview or manual"); return true, 2 }
					value.Optimization.ExecutionMode = strings.ToLower(args[index+1]); index++
				default:
					fmt.Fprintf(os.Stderr, "optimize: unknown option %q\n", args[index]); return true, 2
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown config command %q\n", args[0])
		return true, 2
	}

	saved, err := settings.SaveDefault(value)
	if err != nil { fmt.Fprintln(os.Stderr, "save settings:", err); return true, 1 }
	printConfig(saved)
	return true, 0
}

func printConfig(value settings.Settings) {
	fmt.Println("VARNQOR CONFIG")
	fmt.Printf("Live refresh   %ds\n", value.RefreshSeconds)
	fmt.Printf("Language       %s\n", value.Language)
	fmt.Printf("Theme          %s\n", strings.ToUpper(value.Theme))
	if value.Budget.Enabled { fmt.Printf("Daily budget   %d tokens · reset %02d:00 local · CONFIGURED\n", value.Budget.LimitTokens, value.Budget.ResetHour) } else { fmt.Println("Daily budget   OFF") }
	if value.Guardrails.Enabled { fmt.Printf("Pulse          ON · spike %.0f%% · project %.0f%% · session %.0f%% · min %d\n", value.Guardrails.SpikePercent, value.Guardrails.ProjectConcentrationPercent, value.Guardrails.SessionConcentrationPercent, value.Guardrails.MinimumTokens) } else { fmt.Println("Pulse          OFF") }
	if value.Optimization.Enabled { fmt.Printf("Optimize       %s · target %.0f%% · horizon %dm · %s\n", strings.ToUpper(value.Optimization.Goal), value.Optimization.TargetReductionPercent, value.Optimization.HorizonMinutes, strings.ToUpper(value.Optimization.ExecutionMode)) } else { fmt.Println("Optimize       OFF") }
	fmt.Println("Configured budgets are local user budgets, not official provider quotas.")
	fmt.Println("Optimize remains advisory toward external agents; no silent execution is performed.")
}
