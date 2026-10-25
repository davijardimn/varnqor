package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/doctor"
	"github.com/davijardimn/varnqor/internal/settings"
	"github.com/davijardimn/varnqor/internal/storage"
)

func runCLI(args []string) (bool, int) {
	if len(args) == 0 {
		return false, 0
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println("varnqor", Version)
		return true, 0
	case "doctor":
		store, storeErr := storage.OpenDefault()
		if store != nil {
			defer store.Close()
		}
		report := doctor.Run(store, storeErr, Version)
		if len(args) > 1 && args[1] == "--json" {
			data, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("VARNQOR DOCTOR %s\n\n", Version)
			for _, item := range report.Checks {
				fmt.Printf("%-5s %-26s %s\n", strings.ToUpper(item.Status), item.Label, item.Detail)
			}
			fmt.Printf("\n%s · %d pass · %d warn · %d fail\n", strings.ToUpper(report.Status), report.Pass, report.Warn, report.Fail)
		}
		if report.Fail > 0 {
			return true, 1
		}
		return true, 0
	case "status":
		store, storeErr := storage.OpenDefault()
		if store != nil {
			defer store.Close()
		}
		cfg, cfgErr := settings.LoadDefault()
		status := collectSystemStatus(store, storeErr, cfg, cfgErr, time.Now())
		if len(args) > 1 && args[1] == "--json" {
			data, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("VARNQOR %s\n", Version)
			fmt.Printf("Claude Code: %v · sessions: %d · projects: %d\n", status.Claude.Detected, status.Claude.Sessions, status.Claude.Projects)
			fmt.Printf("Today: %d tokens · %d calls · burn: %s\n", status.Usage.Tokens.Total, status.Usage.Calls, status.Burn.Level)
			if status.Quota.Available {
				fmt.Printf("Budget: %.1f%% used · %d remaining · source: %s\n", status.Quota.UtilizationPercent, status.Quota.Remaining, status.Quota.Source)
			}
			if status.UsageError != "" {
				fmt.Fprintln(os.Stderr, "error:", status.UsageError)
				return true, 1
			}
		}
		return true, 0
	default:
		return false, 0
	}
}
