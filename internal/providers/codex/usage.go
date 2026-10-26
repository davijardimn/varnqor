package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

type TranscriptFile struct { Path, Project, Session string; Subagent bool; Size int64; ModTime time.Time }
type IndexedUsageCall struct { Key string; SourceOffset int64; Call domain.UsageCall }
type Chunk struct { Usage []IndexedUsageCall; NextOffset int64; Warnings int }
type envelope struct { Timestamp string `json:"timestamp"`; Type string `json:"type"`; Payload json.RawMessage `json:"payload"` }
type sessionMeta struct { ID string `json:"id"`; Cwd string `json:"cwd"`; Source json.RawMessage `json:"source"` }
type turnContext struct { Model string `json:"model"` }
type tokenCount struct { Type string `json:"type"`; Info *struct { Total tokenUsage `json:"total_token_usage"`; Last tokenUsage `json:"last_token_usage"` } `json:"info"` }
type tokenUsage struct { Input int64 `json:"input_tokens"`; CachedInput int64 `json:"cached_input_tokens"`; CacheWrite int64 `json:"cache_write_input_tokens"`; Output int64 `json:"output_tokens"`; Reasoning int64 `json:"reasoning_output_tokens"`; Total int64 `json:"total_tokens"` }

func Discover(d Detection) ([]TranscriptFile, error) {
	files := make([]TranscriptFile, 0)
	for _, base := range []string{d.SessionsPath, d.ArchivedPath} {
		if _, err := os.Stat(base); err != nil { if os.IsNotExist(err) { continue }; return nil, err }
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil { return walkErr }; if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") { return nil }
			info, err := entry.Info(); if err != nil { return err }
			project, session, subagent := readMetadata(path); if session == "" { session = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) }
			files = append(files, TranscriptFile{Path:path,Project:project,Session:session,Subagent:subagent,Size:info.Size(),ModTime:info.ModTime()}); return nil
		}); if err != nil { return nil, err }
	}
	sort.Slice(files, func(i,j int) bool { return files[i].Path < files[j].Path }); return files,nil
}

func ReadChunk(file TranscriptFile, offset int64) (Chunk,error) {
	stream,err:=os.Open(file.Path); if err!=nil{return Chunk{},err}; defer stream.Close(); if _,err:=stream.Seek(offset,io.SeekStart);err!=nil{return Chunk{},err}
	reader:=bufio.NewReaderSize(stream,128*1024); chunk:=Chunk{NextOffset:offset}; cursor:=offset; model:="Codex (unspecified)"
	for { lineStart:=cursor; line,readErr:=reader.ReadBytes('\n'); cursor+=int64(len(line)); if len(line)>0&&line[len(line)-1]=='\n' { chunk.NextOffset=cursor; var env envelope; if err:=json.Unmarshal(line,&env);err!=nil{chunk.Warnings++}else{switch env.Type{case"turn_context":var ctx turnContext;if json.Unmarshal(env.Payload,&ctx)==nil&&strings.TrimSpace(ctx.Model)!=""{model=ctx.Model};case"event_msg":var event tokenCount;if json.Unmarshal(env.Payload,&event)==nil&&event.Type=="token_count"&&event.Info!=nil{call,ok:=usageCall(env.Timestamp,model,file,event.Info.Last);if ok{key:=fmt.Sprintf("%s:%d:%d:%d:%d:%d",file.Session,event.Info.Total.Input,event.Info.Total.CachedInput,event.Info.Total.CacheWrite,event.Info.Total.Output,event.Info.Total.Total);chunk.Usage=append(chunk.Usage,IndexedUsageCall{Key:key,SourceOffset:lineStart,Call:call})}}}}}; if readErr!=nil{if readErr==io.EOF{break};return chunk,readErr} }
	return chunk,nil
}

func usageCall(timestamp,model string,file TranscriptFile,value tokenUsage)(domain.UsageCall,bool){when,err:=parseTime(timestamp);if err!=nil{return domain.UsageCall{},false};uncached:=value.Input-value.CachedInput;if uncached<0{uncached=0};total:=value.Total;if total<=0{total=value.Input+value.CacheWrite+value.Output};if total<=0{return domain.UsageCall{},false};return domain.UsageCall{Timestamp:when,Model:normalizeModel(model),Project:displayProject(file.Project),Session:file.Session,Subagent:file.Subagent,Tokens:domain.TokenTotals{Input:uncached,Output:value.Output,CacheCreation:value.CacheWrite,CacheRead:value.CachedInput,Total:total}},true}

func readMetadata(path string)(project,session string,subagent bool){stream,err:=os.Open(path);if err!=nil{return "","",false};defer stream.Close();scanner:=bufio.NewScanner(stream);scanner.Buffer(make([]byte,0,64*1024),1024*1024);for i:=0;i<32&&scanner.Scan();i++{var env envelope;if json.Unmarshal(scanner.Bytes(),&env)!=nil||env.Type!="session_meta"{continue};var meta sessionMeta;if json.Unmarshal(env.Payload,&meta)!=nil{continue};session=meta.ID;if strings.TrimSpace(meta.Cwd)!=""{project=filepath.Base(filepath.Clean(meta.Cwd))};lower:=strings.ToLower(string(meta.Source));subagent=strings.Contains(lower,"subagent")||strings.Contains(lower,"review");break};return}
func normalizeModel(value string)string{value=strings.TrimSpace(value);if value==""||strings.HasPrefix(value,"<"){return"Codex (unspecified)"};return value}
func displayProject(value string)string{value=strings.TrimSpace(value);if value==""||value=="unknown"{return"Unassigned"};return value}
func parseTime(value string)(time.Time,error){value=strings.TrimSpace(value);for _,layout:=range[]string{time.RFC3339Nano,time.RFC3339}{if parsed,err:=time.Parse(layout,value);err==nil{return parsed,nil}};return time.Time{},fmt.Errorf("invalid timestamp %q",value)}
