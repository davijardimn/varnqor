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
		value, err := settings.LoadDefault(); if err != nil { fmt.Fprintln(os.Stderr,"load settings:",err); return true,1 }
		if wantsJSON(args) { printJSON(value) } else { printConfig(value) }; return true,0
	}
	value, err := settings.LoadDefault(); if err != nil { fmt.Fprintln(os.Stderr,"load settings:",err); return true,1 }

	switch args[0] {
	case "refresh":
		if len(args)!=2{fmt.Fprintln(os.Stderr,"usage: varnqor config refresh <2-60 seconds>");return true,2};seconds,err:=strconv.Atoi(args[1]);if err!=nil{fmt.Fprintln(os.Stderr,"refresh interval must be an integer");return true,2};value.RefreshSeconds=seconds
	case "language":
		if len(args)!=2{fmt.Fprintln(os.Stderr,"usage: varnqor config language <system|en|pt-BR|es|de|fr>");return true,2};value.Language=args[1]
	case "theme":
		if len(args)!=2{fmt.Fprintln(os.Stderr,"usage: varnqor config theme <system|light|dark>");return true,2};value.Theme=strings.ToLower(args[1])
	case "budget":
		if len(args)<2{fmt.Fprintln(os.Stderr,"usage: varnqor config budget <tokens|off> [--reset-hour 0-23]");return true,2}
		if args[1]=="off"{value.Budget.Enabled=false}else{amount,err:=settings.ParseTokenAmount(args[1]);if err!=nil{fmt.Fprintln(os.Stderr,"budget:",err);return true,2};value.Budget.Enabled=true;value.Budget.LimitTokens=amount;for index:=2;index<len(args);index++{if args[index]!="--reset-hour"||index+1>=len(args){fmt.Fprintln(os.Stderr,"usage: varnqor config budget <tokens|off> [--reset-hour 0-23]");return true,2};hour,err:=strconv.Atoi(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"reset hour must be an integer from 0 to 23");return true,2};value.Budget.ResetHour=hour;index++}}
	case "optimize":
		if len(args)<2{fmt.Fprintln(os.Stderr,"usage: varnqor config optimize <balanced|economy|throughput|off> [--target 5-60] [--horizon 30-1440] [--mode preview|manual]");return true,2}
		if strings.EqualFold(args[1],"off"){value.Optimization.Enabled=false}else{value.Optimization.Enabled=true;value.Optimization.Goal=strings.ToLower(args[1]);for index:=2;index<len(args);index++{switch args[index]{case"--target":if index+1>=len(args){fmt.Fprintln(os.Stderr,"optimize: --target requires a value");return true,2};parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"optimize: target must be numeric");return true,2};value.Optimization.TargetReductionPercent=parsed;index++;case"--horizon":if index+1>=len(args){fmt.Fprintln(os.Stderr,"optimize: --horizon requires minutes");return true,2};parsed,err:=strconv.Atoi(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"optimize: horizon must be integer minutes");return true,2};value.Optimization.HorizonMinutes=parsed;index++;case"--mode":if index+1>=len(args){fmt.Fprintln(os.Stderr,"optimize: --mode requires preview or manual");return true,2};value.Optimization.ExecutionMode=strings.ToLower(args[index+1]);index++;default:fmt.Fprintf(os.Stderr,"optimize: unknown option %q\n",args[index]);return true,2}}}
	case "control":
		if len(args)<2{fmt.Fprintln(os.Stderr,"usage: varnqor config control <balanced|economy|throughput|off> [policy options]");return true,2}
		if strings.EqualFold(args[1],"off"){value.Control.Enabled=false}else{value.Control.Enabled=true;applyControlProfile(&value,args[1]);for index:=2;index<len(args);index++{if index+1>=len(args){fmt.Fprintf(os.Stderr,"control: %s requires a value\n",args[index]);return true,2};switch args[index]{case"--burn":parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"control: --burn must be numeric");return true,2};value.Control.BurnCeilingRatio=parsed;case"--session":parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"control: --session must be numeric");return true,2};value.Control.SessionDominancePercent=parsed;case"--project":parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"control: --project must be numeric");return true,2};value.Control.ProjectDominancePercent=parsed;case"--cache":parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"control: --cache must be numeric");return true,2};value.Control.CacheFloorPercent=parsed;case"--budget-warning":parsed,err:=strconv.ParseFloat(args[index+1],64);if err!=nil{fmt.Fprintln(os.Stderr,"control: --budget-warning must be numeric");return true,2};value.Control.BudgetWarningPercent=parsed;case"--active":parsed,err:=strconv.Atoi(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"control: --active must be integer minutes");return true,2};value.Control.MissionActiveMinutes=parsed;case"--cooling":parsed,err:=strconv.Atoi(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"control: --cooling must be integer minutes");return true,2};value.Control.MissionCoolingMinutes=parsed;case"--horizon":parsed,err:=strconv.Atoi(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"control: --horizon must be integer minutes");return true,2};value.Control.DecisionHorizonMinutes=parsed;case"--minimum":amount,err:=settings.ParseTokenAmount(args[index+1]);if err!=nil{fmt.Fprintln(os.Stderr,"control: --minimum:",err);return true,2};value.Control.MinimumEvaluationTokens=amount;default:fmt.Fprintf(os.Stderr,"control: unknown option %q\n",args[index]);return true,2};index++}}
	default: fmt.Fprintf(os.Stderr,"unknown config command %q\n",args[0]);return true,2
	}
	saved,err:=settings.SaveDefault(value);if err!=nil{fmt.Fprintln(os.Stderr,"save settings:",err);return true,1};printConfig(saved);return true,0
}

func applyControlProfile(value *settings.Settings, profile string){
	profile=strings.ToLower(strings.TrimSpace(profile));value.Control.Profile=profile
	switch profile{
	case"economy":value.Control.BurnCeilingRatio=1.45;value.Control.SessionDominancePercent=55;value.Control.ProjectDominancePercent=65;value.Control.CacheFloorPercent=20;value.Control.BudgetWarningPercent=72;value.Control.MissionActiveMinutes=10;value.Control.MissionCoolingMinutes=40;value.Control.DecisionHorizonMinutes=120;value.Control.MinimumEvaluationTokens=800
	case"throughput":value.Control.BurnCeilingRatio=2.25;value.Control.SessionDominancePercent=78;value.Control.ProjectDominancePercent=85;value.Control.CacheFloorPercent=8;value.Control.BudgetWarningPercent=92;value.Control.MissionActiveMinutes=15;value.Control.MissionCoolingMinutes=60;value.Control.DecisionHorizonMinutes=180;value.Control.MinimumEvaluationTokens=1500
	default:value.Control.Profile="balanced";value.Control.BurnCeilingRatio=1.70;value.Control.SessionDominancePercent=62;value.Control.ProjectDominancePercent=72;value.Control.CacheFloorPercent=15;value.Control.BudgetWarningPercent=80;value.Control.MissionActiveMinutes=12;value.Control.MissionCoolingMinutes=45;value.Control.DecisionHorizonMinutes=120;value.Control.MinimumEvaluationTokens=1000
	}
}

func printConfig(value settings.Settings) {
	fmt.Println("VARNQOR CONFIG")
	fmt.Printf("Live refresh   %ds\n",value.RefreshSeconds)
	fmt.Printf("Language       %s\n",value.Language)
	fmt.Printf("Theme          %s\n",strings.ToUpper(value.Theme))
	if value.Budget.Enabled{fmt.Printf("Daily budget   %d tokens · reset %02d:00 local · CONFIGURED\n",value.Budget.LimitTokens,value.Budget.ResetHour)}else{fmt.Println("Daily budget   OFF")}
	if value.Guardrails.Enabled{fmt.Printf("Pulse          ON · spike %.0f%% · project %.0f%% · session %.0f%% · min %d\n",value.Guardrails.SpikePercent,value.Guardrails.ProjectConcentrationPercent,value.Guardrails.SessionConcentrationPercent,value.Guardrails.MinimumTokens)}else{fmt.Println("Pulse          OFF")}
	if value.Optimization.Enabled{fmt.Printf("Optimize       %s · target %.0f%% · horizon %dm · %s\n",strings.ToUpper(value.Optimization.Goal),value.Optimization.TargetReductionPercent,value.Optimization.HorizonMinutes,strings.ToUpper(value.Optimization.ExecutionMode))}else{fmt.Println("Optimize       OFF")}
	if value.Control.Enabled{fmt.Printf("Control        %s · burn %.2fx · session %.0f%% · project %.0f%% · budget %.0f%%\n",strings.ToUpper(value.Control.Profile),value.Control.BurnCeilingRatio,value.Control.SessionDominancePercent,value.Control.ProjectDominancePercent,value.Control.BudgetWarningPercent)}else{fmt.Println("Control        OFF")}
	fmt.Println("Configured budgets are local user budgets, not official provider quotas.")
	fmt.Println("Control and Optimize remain advisory toward external agents; silent execution is disabled.")
}
