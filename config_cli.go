package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/davijardimn/varnqor/internal/settings"
)

func runConfigCLI(args []string) (bool, int) {
	if len(args) == 0 || args[0] == "show" {
		value, err := settings.LoadDefault()
		if err != nil {
			fmt.Fprintln(os.Stderr, "load settings:", err)
			return true, 1
		}
		if wantsJSON(args) {
			printJSON(value)
		} else {
			printConfig(value)
		}
		return true, 0
	}

	value, err := settings.LoadDefault()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load settings:", err)
		return true, 1
	}

	switch args[0] {
	case "refresh":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: varnqor config refresh <2-60 seconds>")
			return true, 2
		}
		seconds, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "refresh interval must be an integer")
			return true, 2
		}
		value.RefreshSeconds = seconds
	case "budget":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: varnqor config budget <tokens|off> [--reset-hour 0-23]")
			return true, 2
		}
		if args[1] == "off" {
			value.Budget.Enabled = false
		} else {
			amount, err := settings.ParseTokenAmount(args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "budget:", err)
				return true, 2
			}
			value.Budget.Enabled = true
			value.Budget.LimitTokens = amount
			for index := 2; index < len(args); index++ {
				if args[index] != "--reset-hour" || index+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "usage: varnqor config budget <tokens|off> [--reset-hour 0-23]")
					return true, 2
				}
				hour, err := strconv.Atoi(args[index+1])
				if err != nil {
					fmt.Fprintln(os.Stderr, "reset hour must be an integer from 0 to 23")
					return true, 2
				}
				value.Budget.ResetHour = hour
				index++
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown config command %q\n", args[0])
		return true, 2
	}

	saved, err := settings.SaveDefault(value)
	if err != nil {
		fmt.Fprintln(os.Stderr, "save settings:", err)
		return true, 1
	}
	printConfig(saved)
	return true, 0
}

func printConfig(value settings.Settings) {
	fmt.Println("VARNQOR CONFIG")
	fmt.Printf("Live refresh   %ds\n", value.RefreshSeconds)
	if value.Budget.Enabled {
		fmt.Printf("Daily budget   %d tokens · reset %02d:00 local · CONFIGURED\n", value.Budget.LimitTokens, value.Budget.ResetHour)
	} else {
		fmt.Println("Daily budget   OFF")
	}
	fmt.Println("Configured budgets are local user budgets, not official provider quotas.")
}
