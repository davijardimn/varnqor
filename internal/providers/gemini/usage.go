package gemini

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

type TranscriptFile struct { Path,Project,Session string; Subagent,LegacyJSON bool; Size int64; ModTime time.Time }
type IndexedUsageCall struct { Key string; SourceOffset int64; Call domain.UsageCall }
type Chunk struct { Usage []IndexedUsageCall; NextOffset int64; Warnings int }
type tokensSummary struct { Input int64 `json:"input"`; Output int64 `json:"output"`; Cached int64 `json:"cached"`; Thoughts int64 `json:"thoughts"`; Tool int64 `json:"tool"`; Total int64 `json:"total"` }
type messageRecord struct { ID string `json:"id"`; Timestamp string `json:"timestamp"`; Type string `json:"type"`; Model string `json:"model"`; Tokens *tokensSummary `json:"tokens"` }
type conversationRecord struct { SessionID string `json:"sessionId"`; ProjectHash string `json:"projectHash"`; Directories []string `json:"directories"`; Kind string `json:"kind"`; Messages []messageRecord `json:"messages"` }

func Discover(d Detection)([]TranscriptFile,error){files:=make([]TranscriptFile,0);if _,err:=os.Stat(d.SessionsPath);err!=nil{if os.IsNotExist(err){return files,nil};return nil,err};err:=filepath.WalkDir(d.SessionsPath,func(path string,entry os.DirEntry,walkErr error)error{if walkErr!=nil{return walkErr};if entry.IsDir(){return nil};name:=strings.ToLower(entry.Name());if !strings.HasPrefix(name,"session-")||(!strings.HasSuffix(name,".jsonl")&&!strings.HasSuffix(name,".json")){return nil};if !strings.Contains(strings.ToLower(filepath.ToSlash(path)),"/chats/"){return nil};info,err:=entry.Info();if err!=nil{return err};project:=projectFromPath(d.SessionsPath,path);session:=strings.TrimSuffix(strings.TrimSuffix(entry.Name(),".jsonl"),".json");file:=TranscriptFile{Path:path,Project:project,Session:session,LegacyJSON:strings.HasSuffix(name,".json"),Size:info.Size(),ModTime:info.ModTime()};if file.LegacyJSON{if meta,ok:=legacyMetadata(path);ok{file.Session=choose(meta.SessionID,file.Session);if len(meta.Directories)>0&&strings.TrimSpace(meta.Directories[0])!=""{file.Project=filepath.Base(filepath.Clean(meta.Directories[0]))};file.Subagent=strings.EqualFold(meta.Kind,"subagent")}};files=append(files,file);return nil});if err!=nil{return nil,err};sort.Slice(files,func(i,j int)bool{return files[i].Path<files[j].Path});return files,nil}

func ReadChunk(file TranscriptFile,offset int64)(Chunk,error){if file.LegacyJSON{return ReadLegacy(file)};stream,err:=os.Open(file.Path);if err!=nil{return Chunk{},err};defer stream.Close();if _,err:=stream.Seek(offset,io.SeekStart);err!=nil{return Chunk{},err};reader:=bufio.NewReaderSize(stream,128*1024);chunk:=Chunk{NextOffset:offset};cursor:=offset;for{lineStart:=cursor;line,readErr:=reader.ReadBytes('\n');cursor+=int64(len(line));if len(line)>0&&line[len(line)-1]=='\n'{chunk.NextOffset=cursor;var message messageRecord;if err:=json.Unmarshal(line,&message);err!=nil{chunk.Warnings++}else if call,ok:=messageCall(message,file);ok{chunk.Usage=append(chunk.Usage,IndexedUsageCall{Key:choose(message.ID,fmt.Sprintf("%s:%d",file.Session,lineStart)),SourceOffset:lineStart,Call:call})}};if readErr!=nil{if readErr==io.EOF{break};return chunk,readErr}};return chunk,nil}
func ReadLegacy(file TranscriptFile)(Chunk,error){data,err:=os.ReadFile(file.Path);if err!=nil{return Chunk{},err};var conversation conversationRecord;if err:=json.Unmarshal(data,&conversation);err!=nil{return Chunk{NextOffset:int64(len(data)),Warnings:1},nil};if conversation.SessionID!=""{file.Session=conversation.SessionID};if len(conversation.Directories)>0&&strings.TrimSpace(conversation.Directories[0])!=""{file.Project=filepath.Base(filepath.Clean(conversation.Directories[0]))};file.Subagent=strings.EqualFold(conversation.Kind,"subagent");chunk:=Chunk{NextOffset:int64(len(data))};for index,message:=range conversation.Messages{if call,ok:=messageCall(message,file);ok{chunk.Usage=append(chunk.Usage,IndexedUsageCall{Key:choose(message.ID,fmt.Sprintf("%s:%d",file.Session,index)),SourceOffset:int64(index),Call:call})}};return chunk,nil}
func messageCall(message messageRecord,file TranscriptFile)(domain.UsageCall,bool){if !strings.EqualFold(message.Type,"gemini")||message.Tokens==nil{return domain.UsageCall{},false};when,err:=parseTime(message.Timestamp);if err!=nil{return domain.UsageCall{},false};value:=*message.Tokens;uncached:=value.Input-value.Cached;if uncached<0{uncached=0};total:=value.Total;if total<=0{total=value.Input+value.Output};if total<=0{return domain.UsageCall{},false};return domain.UsageCall{MessageID:message.ID,Timestamp:when,Model:normalizeModel(message.Model),Project:displayProject(file.Project),Session:file.Session,Subagent:file.Subagent,Tokens:domain.TokenTotals{Input:uncached,Output:value.Output,CacheRead:value.Cached,Total:total}},true}
func legacyMetadata(path string)(conversationRecord,bool){data,err:=os.ReadFile(path);if err!=nil{return conversationRecord{},false};var value conversationRecord;if json.Unmarshal(data,&value)!=nil{return conversationRecord{},false};return value,true}
func projectFromPath(root,path string)string{rel,err:=filepath.Rel(root,path);if err!=nil{return"Unassigned"};parts:=strings.Split(filepath.ToSlash(rel),"/");if len(parts)>0&&strings.TrimSpace(parts[0])!=""{value:=parts[0];if len(value)>12{value=value[:12]};return value};return"Unassigned"}
func normalizeModel(value string)string{value=strings.TrimSpace(value);if value==""||strings.HasPrefix(value,"<"){return"Gemini (unspecified)"};return value}
func displayProject(value string)string{value=strings.TrimSpace(value);if value==""||value=="unknown"{return"Unassigned"};return value}
func parseTime(value string)(time.Time,error){for _,layout:=range[]string{time.RFC3339Nano,time.RFC3339}{if parsed,err:=time.Parse(layout,strings.TrimSpace(value));err==nil{return parsed,nil}};return time.Time{},fmt.Errorf("invalid timestamp %q",value)}
func choose(value,fallback string)string{if strings.TrimSpace(value)!=""{return value};return fallback}
