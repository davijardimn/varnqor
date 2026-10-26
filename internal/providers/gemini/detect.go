package gemini

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Detection struct {
	Detected bool `json:"detected"`; BinaryDetected bool `json:"binaryDetected"`; DataDetected bool `json:"dataDetected"`
	Home string `json:"home"`; SessionsPath string `json:"sessionsPath"`; Sessions int `json:"sessions"`; LastActivity time.Time `json:"lastActivity,omitempty"`
}
func Detect() Detection { home,err:=os.UserHomeDir(); if err!=nil{return Detection{}}; return DetectAt(home,exec.LookPath) }
func DetectAt(userHome string, lookup func(string)(string,error)) Detection {
	root:=filepath.Join(userHome,".gemini"); result:=Detection{Home:root,SessionsPath:filepath.Join(root,"tmp")}
	if lookup!=nil{if _,err:=lookup("gemini");err==nil{result.BinaryDetected=true}}
	_ = filepath.WalkDir(result.SessionsPath,func(path string,entry os.DirEntry,err error)error{if err!=nil||entry.IsDir(){return nil};name:=strings.ToLower(entry.Name());if !strings.HasPrefix(name,"session-")||(!strings.HasSuffix(name,".jsonl")&&!strings.HasSuffix(name,".json")){return nil};if !strings.Contains(strings.ToLower(filepath.ToSlash(path)),"/chats/"){return nil};result.Sessions++;result.DataDetected=true;if info,infoErr:=entry.Info();infoErr==nil&&info.ModTime().After(result.LastActivity){result.LastActivity=info.ModTime()};return nil})
	result.Detected=result.BinaryDetected||result.DataDetected;return result
}
