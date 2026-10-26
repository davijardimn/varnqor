package main

import (
	"context"
	"fmt"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/doctor"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/indexer"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/providers/codex"
	"github.com/davijardimn/varnqor/internal/providers/gemini"
	"github.com/davijardimn/varnqor/internal/settings"
	"github.com/davijardimn/varnqor/internal/storage"
)

type App struct { ctx context.Context; store *storage.Store; storeErr error }
type SystemStatus struct {
	Version string `json:"version"`; GeneratedAt time.Time `json:"generatedAt"`; Providers []domain.ProviderStatus `json:"providers"`; Claude claude.Detection `json:"claude"`
	Usage domain.UsageOverview `json:"usage"`; Hourly []domain.HourUsage `json:"hourly"`; Burn analytics.BurnRate `json:"burn"`; Forecast domain.ForecastReport `json:"forecast"`; Quota domain.QuotaStatus `json:"quota"`; Attribution domain.AttributionReport `json:"attribution"`; Waste domain.WasteReport `json:"waste"`; WasteScope string `json:"wasteScope"`; Advisor domain.AdvisorReport `json:"advisor"`; Sync domain.SyncStats `json:"sync"`; StorageReady bool `json:"storageReady"`; UsageError string `json:"usageError,omitempty"`; SettingsError string `json:"settingsError,omitempty"`
}
func NewApp()*App{return &App{}}
func(a *App)startup(ctx context.Context){a.ctx=ctx;a.store,a.storeErr=storage.OpenDefault();if a.storeErr==nil&&a.store!=nil{a.storeErr=a.store.PrepareV1()}}
func(a *App)shutdown(context.Context){if a.store!=nil{_ = a.store.Close()}}
func(a *App)GetSettings()(settings.Settings,error){return settings.LoadDefault()}
func(a *App)SaveSettings(value settings.Settings)(settings.Settings,error){return settings.SaveDefault(value)}
func(a *App)RunDoctor()domain.DoctorReport{return doctor.Run(a.store,a.storeErr,Version)}
func(a *App)GetSystemStatus()SystemStatus{cfg,cfgErr:=settings.LoadDefault();return collectSystemStatus(a.store,a.storeErr,cfg,cfgErr,time.Now())}

func collectSystemStatus(store *storage.Store,storeErr error,cfg settings.Settings,cfgErr error,now time.Time)SystemStatus{
	cd:=claude.Detect();xd:=codex.Detect();gd:=gemini.Detect();status:=SystemStatus{Version:Version,GeneratedAt:now,Claude:cd,StorageReady:store!=nil&&storeErr==nil,WasteScope:"claude"}
	if cfgErr!=nil{status.SettingsError=cfgErr.Error();cfg=settings.Default()};if storeErr!=nil{status.UsageError=fmt.Sprintf("open local usage store: %v",storeErr);status.Providers=detectedProviderStatuses(nil,now,cd,xd,gd,nil);return status};if store==nil{status.UsageError="local usage store is not initialized";status.Providers=detectedProviderStatuses(nil,now,cd,xd,gd,nil);return status}
	errs:=map[string]error{};if cd.DataDetected{stats,err:=indexer.SyncClaude(store,cd.ProjectsPath);mergeSync(&status.Sync,stats);if err!=nil{errs["claude"]=err}};if xd.DataDetected{stats,err:=indexer.SyncCodex(store,xd);mergeSync(&status.Sync,stats);if err!=nil{errs["codex"]=err}};if gd.DataDetected{stats,err:=indexer.SyncGemini(store,gd);mergeSync(&status.Sync,stats);if err!=nil{errs["gemini"]=err}}
	usage,hourly,err:=store.DailyUsageAll(now);if err!=nil{status.UsageError=fmt.Sprintf("query aggregate local usage: %v",err);status.Providers=detectedProviderStatuses(store,now,cd,xd,gd,errs);return status};files,warnings,err:=store.SourceSummaryAll();if err!=nil{status.UsageError=fmt.Sprintf("query source summary: %v",err);return status};usage.FilesScanned,usage.ParseWarnings=files,warnings;status.Usage,status.Hourly=usage,hourly;status.Providers=detectedProviderStatuses(store,now,cd,xd,gd,errs)
	currentStart:=now.Add(-time.Hour);current,_,err:=store.UsageWindowAll(currentStart,now);if err!=nil{status.UsageError=fmt.Sprintf("query current burn window: %v",err);return status};previous,_,err:=store.UsageWindowAll(now.Add(-2*time.Hour),currentStart);if err!=nil{status.UsageError=fmt.Sprintf("query previous burn window: %v",err);return status};localNow:=now.In(now.Location());currentHour:=time.Date(localNow.Year(),localNow.Month(),localNow.Day(),localNow.Hour(),0,0,0,localNow.Location());baseline:=make([]int64,0,len(hourly));for _,bucket:=range hourly{if bucket.Start.Before(currentHour)&&bucket.Tokens.Total>0{baseline=append(baseline,bucket.Tokens.Total)}};status.Burn=analytics.ComputeBurnRate(current.Total,previous.Total,baseline);status.Forecast=analytics.ComputeForecast(status.Burn,2*time.Hour)
	status.Attribution,err=store.AttributionAll(currentStart,now,8);if err!=nil{status.UsageError=fmt.Sprintf("query burn attribution: %v",err);return status};if toolCalls,toolErr:=store.ToolCalls("claude",currentStart,now);toolErr==nil{status.Waste=analytics.AnalyzeWaste(toolCalls,currentStart,now)};applyConfiguredQuota(&status,store,cfg,now);if status.Quota.Available{status.Forecast.RunwayAvailable=status.Quota.RunwayAvailable;status.Forecast.LimitSource=status.Quota.Source};status.Advisor=analytics.BuildAdvisor(now,status.Burn,status.Forecast,status.Attribution,status.Waste);return status
}

func detectedProviderStatuses(store *storage.Store,now time.Time,c claude.Detection,x codex.Detection,g gemini.Detection,errs map[string]error)[]domain.ProviderStatus{
	items:=[]domain.ProviderStatus{{ID:"claude",Name:"Claude Code",Detected:c.Detected,BinaryDetected:c.BinaryDetected,DataDetected:c.DataDetected,HomePath:c.ClaudeHome,DataPath:c.ProjectsPath,ObservedSessions:c.Sessions,ObservedProjects:c.Projects,LastActivity:c.LastActivity,Capabilities:domain.ProviderCapabilities{Usage:true,Tools:true}},{ID:"codex",Name:"Codex",Detected:x.Detected,BinaryDetected:x.BinaryDetected,DataDetected:x.DataDetected,HomePath:x.Home,DataPath:x.SessionsPath,ObservedSessions:x.Sessions,UnsupportedFiles:x.UnsupportedFiles,LastActivity:x.LastActivity,Capabilities:domain.ProviderCapabilities{Usage:true},Note:"JSONL usage is indexed locally. Compressed session archives are reported but not indexed."},{ID:"gemini",Name:"Gemini CLI",Detected:g.Detected,BinaryDetected:g.BinaryDetected,DataDetected:g.DataDetected,HomePath:g.Home,DataPath:g.SessionsPath,ObservedSessions:g.Sessions,LastActivity:g.LastActivity,Capabilities:domain.ProviderCapabilities{Usage:true},Note:"Usage is read from local chat sessions. Tool-level waste analysis is not enabled for Gemini in 1.0."}}
	for i:=range items{item:=&items[i];switch{case errs!=nil&&errs[item.ID]!=nil:item.State="error";item.Error=errs[item.ID].Error();case item.DataDetected:item.State="connected";case item.BinaryDetected:item.State="ready";default:item.State="unavailable"};if store!=nil{calls,tokens,files,warnings,last,err:=store.ProviderTodaySummary(item.ID,now);if err==nil{item.IndexedCalls=calls;item.TokensToday=tokens;item.FilesIndexed=files;item.ParseWarnings=warnings;if last.After(item.LastActivity){item.LastActivity=last}}};if item.UnsupportedFiles>0&&item.State=="connected"{item.State="partial"}}
	return items
}
func mergeSync(target *domain.SyncStats,value domain.SyncStats){target.FilesSeen+=value.FilesSeen;target.FilesChanged+=value.FilesChanged;target.BytesIndexed+=value.BytesIndexed;target.RecordsObserved+=value.RecordsObserved;target.ParseWarnings+=value.ParseWarnings;target.Rebuilt=target.Rebuilt||value.Rebuilt}
func applyConfiguredQuota(status *SystemStatus,store *storage.Store,cfg settings.Settings,now time.Time){if !cfg.Budget.Enabled||cfg.Budget.LimitTokens<=0||store==nil{return};start,resetAt:=analytics.ConfiguredBudgetWindow(now,cfg.Budget.ResetHour);used,_,err:=store.UsageWindowAll(start,now);if err!=nil{if status.UsageError==""{status.UsageError=fmt.Sprintf("query configured budget window: %v",err)};return};status.Quota=analytics.ComputeConfiguredQuota(now,start,resetAt,cfg.Budget.Label,cfg.Budget.LimitTokens,used.Total,status.Burn.CurrentPerHour)}
