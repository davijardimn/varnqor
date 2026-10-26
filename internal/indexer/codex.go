package indexer

import (
	"fmt"

	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/providers/codex"
	"github.com/davijardimn/varnqor/internal/storage"
)

const codexProvider = "codex"

func SyncCodex(store *storage.Store, detection codex.Detection) (domain.SyncStats, error) {
	files, err := codex.Discover(detection); if err != nil { return domain.SyncStats{}, err }
	stats := domain.SyncStats{FilesSeen:len(files)}; states,err:=store.SourceStates(codexProvider);if err!=nil{return stats,err};stateByPath:=make(map[string]storage.SourceState,len(states));for _,state:=range states{stateByPath[state.Path]=state}
	for _,file:=range files{state,exists:=stateByPath[file.Path];offset:=int64(0);if exists{offset=state.Offset;if state.Offset==file.Size&&state.Size==file.Size&&state.ModTimeNS==file.ModTime.UnixNano(){continue};if file.Size<state.Offset{if err:=store.ResetSource(codexProvider,file.Path);err!=nil{return stats,err};offset=0;stats.Rebuilt=true}}
		chunk,err:=codex.ReadChunk(file,offset);if err!=nil{return stats,fmt.Errorf("index Codex %s: %w",file.Path,err)};checkpoint,err:=fileCheckpoint(file.Path,chunk.NextOffset);if err!=nil{return stats,err};records:=make([]storage.UsageRecord,0,len(chunk.Usage));for _,indexed:=range chunk.Usage{records=append(records,storage.UsageRecord{Key:indexed.Key,Provider:codexProvider,SourcePath:file.Path,SourceOffset:indexed.SourceOffset,Call:indexed.Call})};if err:=store.ApplyFile(storage.SourceState{Provider:codexProvider,Path:file.Path,Size:file.Size,ModTimeNS:file.ModTime.UnixNano(),Offset:chunk.NextOffset,Checkpoint:checkpoint,ParseWarnings:state.ParseWarnings+chunk.Warnings},records);err!=nil{return stats,err};stats.FilesChanged++;stats.BytesIndexed+=chunk.NextOffset-offset;stats.RecordsObserved+=len(records);stats.ParseWarnings+=chunk.Warnings}
	return stats,nil
}
