package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/davijardimn/varnqor/internal/storage"
)

func runOptimizeCLI(args []string) (bool, int) {
	filter, err := parseExploreArgs(args)
	if err != nil { fmt.Fprintln(os.Stderr,"optimize:",err); return true,2 }
	status := loadCLIStatus()
	if status.UsageError != "" { fmt.Fprintln(os.Stderr,"optimize:",status.UsageError); return true,1 }
	store, err := storage.OpenDefault()
	if err != nil { fmt.Fprintln(os.Stderr,"optimize:",err); return true,1 }
	defer store.Close()
	app := &App{store:store}
	report, err := app.GetOptimization(filter)
	if err != nil { fmt.Fprintln(os.Stderr,"optimize:",err); return true,1 }
	if wantsJSON(args) { printJSON(report); return true,0 }

	fmt.Printf("VARNQOR OPTIMIZE · %s · %s\n\n",strings.ToUpper(report.Goal),exploreRangeLabel(filter.RangeMinutes))
	fmt.Printf("State          %s\n",strings.ToUpper(report.State))
	fmt.Printf("Current pace   %d tokens/h\n",report.Scenario.CurrentPerHour)
	fmt.Printf("Projection     %d tokens / %s\n",report.Scenario.ProjectedTokens,exploreRangeLabel(report.Scenario.HorizonMinutes))
	fmt.Printf("Target         -%.0f%% · %d tokens · %s\n",report.Scenario.TargetReductionPercent,report.Scenario.TargetTokens,strings.ToUpper(report.Scenario.Feasibility))
	fmt.Printf("Execution      %s · external auto-execution: OFF\n",strings.ToUpper(report.Mode))
	if len(report.Actions)==0 {
		fmt.Println("\nNo optimization playbook is strong enough to surface for this window.")
		return true,0
	}
	fmt.Printf("\nPLAYBOOKS · %d\n",len(report.Actions))
	for index, action := range report.Actions {
		fmt.Printf("%02d  %-6s %s\n",index+1,strings.ToUpper(action.Priority),action.Title)
		fmt.Printf("    %s\n",action.Evidence)
		if action.EstimatedSavingsHigh > 0 { fmt.Printf("    impact: ~%d-%d tokens (non-additive, %s confidence)\n",action.EstimatedSavingsLow,action.EstimatedSavingsHigh,action.Confidence) }
		for _, step := range action.Steps { fmt.Printf("    -> %s\n",step) }
	}
	return true,0
}
