package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/davijardimn/varnqor/internal/storage"
)

func runControlCLI(args []string)(bool,int){
	filter,err:=parseExploreArgs(args);if err!=nil{fmt.Fprintln(os.Stderr,"control:",err);return true,2}
	store,err:=storage.OpenDefault();if err!=nil{fmt.Fprintln(os.Stderr,"control:",err);return true,1};defer store.Close()
	if err:=store.PrepareV1();err!=nil{fmt.Fprintln(os.Stderr,"control:",err);return true,1}
	app:=&App{store:store}
	report,err:=app.GetControlPlane(filter);if err!=nil{fmt.Fprintln(os.Stderr,"control:",err);return true,1}
	if wantsJSON(args){printJSON(report);return true,0}
	c:=report.Control
	fmt.Printf("VARNQOR CONTROL · %s · %s\n\n",strings.ToUpper(c.State),exploreRangeLabel(c.RangeMinutes))
	fmt.Printf("Fleet          %d active provider(s)\n",c.ActiveProviders)
	fmt.Printf("Missions       %d active · %d cooling\n",c.ActiveMissions,c.CoolingMissions)
	fmt.Printf("Pace           %d tokens/h\n",c.CurrentPerHour)
	fmt.Printf("Projection     %d tokens / %s\n",c.ProjectedTokens,exploreRangeLabel(c.Policy.DecisionHorizonMinutes))
	fmt.Printf("Policy         %s · burn ceiling %.2fx\n",strings.ToUpper(c.Policy.Profile),c.Policy.BurnCeilingRatio)
	fmt.Printf("Incidents      %d high · %d medium · %d low\n",c.HighIncidents,c.MediumIncidents,c.LowIncidents)
	fmt.Println("External exec  OFF")
	if len(c.Incidents)>0{fmt.Println("\nINCIDENTS");for i,item:=range c.Incidents{if i>=6{break};fmt.Printf("%-6s %s\n",strings.ToUpper(item.Severity),item.Title);fmt.Printf("       %s\n",item.Detail)}}
	if len(c.Decisions)>0{fmt.Println("\nDECISION QUEUE");for i,item:=range c.Decisions{if i>=6{break};fmt.Printf("%02d  %-6s %s\n",i+1,strings.ToUpper(item.Priority),item.Title);fmt.Printf("    %s\n",item.Reason);if item.EstimatedSavingsHigh>0{fmt.Printf("    impact: ~%d-%d tokens (non-additive)\n",item.EstimatedSavingsLow,item.EstimatedSavingsHigh)}}}
	if len(c.Missions)>0{fmt.Println("\nMISSIONS");for i,item:=range c.Missions{if i>=8{break};fmt.Printf("%-8s %-9s %-9s %7.1f%%  %s / %s\n",strings.ToUpper(item.State),strings.ToUpper(item.Provider),strings.ToUpper(item.Pressure),item.SharePercent,trimWidth(item.Project,24),shortCLIControlID(item.Session))}}
	return true,0
}

func runBriefCLI(args []string)(bool,int){
	filter,err:=parseExploreArgs(args);if err!=nil{fmt.Fprintln(os.Stderr,"brief:",err);return true,2}
	store,err:=storage.OpenDefault();if err!=nil{fmt.Fprintln(os.Stderr,"brief:",err);return true,1};defer store.Close()
	if err:=store.PrepareV1();err!=nil{fmt.Fprintln(os.Stderr,"brief:",err);return true,1}
	app:=&App{store:store}
	brief,err:=app.ExportControlBrief(filter);if err!=nil{fmt.Fprintln(os.Stderr,"brief:",err);return true,1}
	fmt.Print(brief);return true,0
}

func shortCLIControlID(value string)string{if len(value)<=12{return value};return value[:10]+"…"}
