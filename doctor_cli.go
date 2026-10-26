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

func runCLI(args []string) (bool, int) { if len(args)==0{return false,0}; switch args[0]{case "help","--help","-h":printCLIHelp();return true,0;case "version","--version","-v":fmt.Println("varnqor",Version);return true,0;case "doctor":return runDoctorCLI(args[1:]);case "providers","agents":return runProvidersCLI(args[1:]);case "status":return runStatusCLI(args[1:]);case "explore":return runExploreCLI(args[1:]);case "pulse":return runPulseCLI(args[1:]);case "export":return runExportCLI(args[1:]);case "analyze":return runAnalyzeCLI(args[1:]);case "forecast":return runForecastCLI(args[1:]);case "config":return runConfigCLI(args[1:]);default:fmt.Fprintf(os.Stderr,"unknown command %q\n\n",args[0]);printCLIHelp();return true,2} }
func runDoctorCLI(args []string)(bool,int){store,storeErr:=storage.OpenDefault();if store!=nil{defer store.Close()};report:=doctor.Run(store,storeErr,Version);if wantsJSON(args){printJSON(report)}else{fmt.Printf("VARNQOR DOCTOR %s\n\n",Version);for _,item:=range report.Checks{fmt.Printf("%-5s %-26s %s\n",strings.ToUpper(item.Status),item.Label,item.Detail)};fmt.Printf("\n%s · %d pass · %d warn · %d fail\n",strings.ToUpper(report.Status),report.Pass,report.Warn,report.Fail)};if report.Fail>0{return true,1};return true,0}
func runProvidersCLI(args []string)(bool,int){status:=loadCLIStatus();if wantsJSON(args){printJSON(status.Providers);return true,statusExitCode(status.UsageError)};fmt.Println("VARNQOR PROVIDERS\n");for _,p:=range status.Providers{caps:=[]string{};if p.Capabilities.Usage{caps=append(caps,"usage")};if p.Capabilities.Tools{caps=append(caps,"tools")};if p.Capabilities.Limits{caps=append(caps,"limits")};fmt.Printf("%-12s %-11s %4d sessions · %4d calls · %10d tokens · %s\n",p.Name,strings.ToUpper(p.State),p.ObservedSessions,p.IndexedCalls,p.TokensToday,strings.Join(caps,","));if p.Error!=""{fmt.Printf("             error: %s\n",p.Error)};if p.UnsupportedFiles>0{fmt.Printf("             %d unsupported compressed session file(s)\n",p.UnsupportedFiles)}};return true,statusExitCode(status.UsageError)}
func runStatusCLI(args []string)(bool,int){status:=loadCLIStatus();if wantsJSON(args){printJSON(status);return true,statusExitCode(status.UsageError)};active:=0;for _,p:=range status.Providers{if p.DataDetected&&p.Error==""{active++}};fmt.Printf("VARNQOR %s\n",Version);fmt.Printf("Providers: %d active · %d known\n",active,len(status.Providers));fmt.Printf("Today: %d tokens · %d token-bearing calls · %d token-bearing sessions\n",status.Usage.Tokens.Total,status.Usage.Calls,status.Usage.Sessions);if status.Burn.Ready{fmt.Printf("Burn: %.2fx %s · pressure: %s\n",status.Burn.Ratio,strings.ToUpper(status.Burn.Level),strings.ToUpper(status.Forecast.Pressure))}else{fmt.Printf("Burn: LEARNING · %d/2 active-hour samples\n",status.Burn.BaselineSamples)};fmt.Printf("Advisor: %d recommendations · Waste scope: %s\n",len(status.Advisor.Recommendations),strings.ToUpper(status.WasteScope));if status.Quota.Available{fmt.Printf("Budget: %.1f%% used · %d remaining · source: %s\n",status.Quota.UtilizationPercent,status.Quota.Remaining,status.Quota.Source)};if status.UsageError!=""{fmt.Fprintln(os.Stderr,"error:",status.UsageError)};return true,statusExitCode(status.UsageError)}
func runAnalyzeCLI(args []string)(bool,int){status:=loadCLIStatus();if wantsJSON(args){payload:=struct{Burn any `json:"burn"`;Attribution any `json:"attribution"`;Waste any `json:"waste"`;WasteScope string `json:"wasteScope"`;Advisor any `json:"advisor"`}{status.Burn,status.Attribution,status.Waste,status.WasteScope,status.Advisor};printJSON(payload);return true,statusExitCode(status.UsageError)};fmt.Println("VARNQOR ANALYSIS\n");if status.Burn.Ready{fmt.Printf("Burn rate      %.2fx · %s · %s\n",status.Burn.Ratio,strings.ToUpper(status.Burn.Level),strings.ToUpper(status.Burn.Trend))}else{fmt.Printf("Burn rate      LEARNING · %d/2 active-hour samples\n",status.Burn.BaselineSamples)};fmt.Printf("Current pace   %d tokens/h\n",status.Burn.CurrentPerHour);fmt.Printf("Waste signals  %d Claude-tool findings · ~%d heuristic context tokens\n",len(status.Waste.Findings),status.Waste.EstimatedSignalTokens);if len(status.Attribution.Projects)>0{top:=status.Attribution.Projects[0];fmt.Printf("Top project    %s · %s · %.1f%% of current window\n",top.Label,strings.ToUpper(top.Provider),top.SharePercent)};fmt.Println();if len(status.Advisor.Recommendations)==0{fmt.Println("Advisor: no recommendation is currently strong enough to surface.")}else{fmt.Printf("ADVISOR · %d recommendation(s)\n",len(status.Advisor.Recommendations));for index,item:=range status.Advisor.Recommendations{fmt.Printf("%02d  %-6s %s\n",index+1,strings.ToUpper(item.Priority),item.Title);fmt.Printf("    %s\n",item.Evidence);fmt.Printf("    -> %s\n",item.Recommendation)}};if status.UsageError!=""{fmt.Fprintln(os.Stderr,"\nerror:",status.UsageError)};return true,statusExitCode(status.UsageError)}
func runForecastCLI(args []string)(bool,int){status:=loadCLIStatus();if wantsJSON(args){payload:=struct{Forecast any `json:"forecast"`;Quota any `json:"quota"`}{status.Forecast,status.Quota};printJSON(payload);return true,statusExitCode(status.UsageError)};fmt.Println("VARNQOR FORECAST\n");fmt.Printf("Pressure       %s\n",strings.ToUpper(status.Forecast.Pressure));fmt.Printf("Current pace   %d tokens/h\n",status.Forecast.CurrentPerHour);fmt.Printf("Next %d min    %d projected activity\n",status.Forecast.HorizonMinutes,status.Forecast.ProjectedTokens);fmt.Printf("Envelope       %d-%d\n",status.Forecast.EnvelopeLow,status.Forecast.EnvelopeHigh);if status.Quota.Available{fmt.Printf("Limit source   %s (%s)\n",strings.ToUpper(status.Quota.Source),status.Quota.Label);fmt.Printf("Remaining      %d\n",status.Quota.Remaining);if status.Quota.RunwayAvailable{fmt.Printf("Runway         %.2f hours at current pace\n",status.Quota.RunwayHours)}else{fmt.Println("Runway         unavailable while current pace is idle")};fmt.Printf("At reset       %d projected · over budget: %v\n",status.Quota.ProjectedAtReset,status.Quota.WillExceedBeforeReset)}else{fmt.Println("Runway         NO TRUSTED LIMIT");fmt.Println("               configure a local budget in Desktop or `varnqor config budget 10m`")};if status.UsageError!=""{fmt.Fprintln(os.Stderr,"\nerror:",status.UsageError)};return true,statusExitCode(status.UsageError)}
func loadCLIStatus()SystemStatus{store,storeErr:=storage.OpenDefault();if store!=nil{defer store.Close();if storeErr==nil{storeErr=store.PrepareV1()}};cfg,cfgErr:=settings.LoadDefault();return collectSystemStatus(store,storeErr,cfg,cfgErr,time.Now())}
func wantsJSON(args []string)bool{for _,arg:=range args{if arg=="--json"{return true}};return false}
func printJSON(value any){data,err:=json.MarshalIndent(value,"","  ");if err!=nil{fmt.Fprintln(os.Stderr,"encode JSON:",err);return};fmt.Println(string(data))}
func statusExitCode(message string)int{if message!=""{return 1};return 0}
func printCLIHelp(){fmt.Printf(`Varnqor %s — resource intelligence for AI coding agents

Usage:
  varnqor                              Open Varnqor Desktop
  varnqor status [--json]              Aggregate local usage summary
  varnqor providers [--json]           Claude, Codex and Gemini provider status
  varnqor explore [options]            Investigate a time window
  varnqor pulse [options]              Evaluate local guardrail signals
  varnqor export [options]             Export the selected window as JSON or CSV
  varnqor analyze [--json]             Burn drivers, waste signals and Advisor guidance
  varnqor forecast [--json]            Resource pressure, projection and trusted runway
  varnqor doctor [--json]              Diagnose providers, storage and settings
  varnqor config show                  Show local settings
  varnqor config refresh 5             Set live refresh interval (2-60s)
  varnqor config budget 10m            Enable a configured daily budget
  varnqor config budget off            Disable the configured daily budget
  varnqor version                      Print the application version
  varnqor help                         Show this help

Explore options:
  --range 1h|6h|24h|7d                 Window length (15m to 7d also accepted)
  --provider claude|codex|gemini       Filter provider
  --project NAME                       Filter project
  --model NAME                         Filter model
  --limit N                            Maximum rows (1-200)
  --json                               JSON output

Export options:
  --format json|csv                    Output format (default json)

Varnqor is local-first. A configured budget is not an official provider quota.
`,Version)}