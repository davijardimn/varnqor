package storage

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func NormalizeExploreFilter(value domain.ExploreFilter) domain.ExploreFilter {
	value.Provider = strings.TrimSpace(strings.ToLower(value.Provider))
	value.Project = strings.TrimSpace(value.Project)
	value.Session = strings.TrimSpace(value.Session)
	value.Model = strings.TrimSpace(value.Model)
	if value.Provider == "all" { value.Provider = "" }
	if value.RangeMinutes <= 0 { value.RangeMinutes = 60 }
	if value.RangeMinutes < 15 { value.RangeMinutes = 15 }
	if value.RangeMinutes > 10080 { value.RangeMinutes = 10080 }
	if value.Limit <= 0 { value.Limit = 50 }
	if value.Limit > 200 { value.Limit = 200 }
	return value
}

func (s *Store) Explore(value domain.ExploreFilter, now time.Time) (domain.ExploreReport, error) {
	filter := NormalizeExploreFilter(value)
	duration := time.Duration(filter.RangeMinutes) * time.Minute
	currentStart := now.Add(-duration)
	previousStart := currentStart.Add(-duration)

	current, err := s.exploreSummary(filter, currentStart, now)
	if err != nil { return domain.ExploreReport{}, err }
	previous, err := s.exploreSummary(filter, previousStart, currentStart)
	if err != nil { return domain.ExploreReport{}, err }

	bucketMinutes := exploreBucketMinutes(filter.RangeMinutes)
	timeline, err := s.exploreTimeline(filter, currentStart, now, time.Duration(bucketMinutes)*time.Minute)
	if err != nil { return domain.ExploreReport{}, err }
	providers, err := s.exploreDimension(filter, currentStart, now, "provider", current.Tokens.Total, 20)
	if err != nil { return domain.ExploreReport{}, err }
	projects, err := s.exploreDimension(filter, currentStart, now, "project", current.Tokens.Total, filter.Limit)
	if err != nil { return domain.ExploreReport{}, err }
	models, err := s.exploreDimension(filter, currentStart, now, "model", current.Tokens.Total, filter.Limit)
	if err != nil { return domain.ExploreReport{}, err }
	sessions, err := s.exploreSessions(filter, currentStart, now, current.Tokens.Total, filter.Limit)
	if err != nil { return domain.ExploreReport{}, err }
	recent, err := s.exploreRecent(filter, currentStart, now, filter.Limit)
	if err != nil { return domain.ExploreReport{}, err }
	facets, err := s.exploreFacets(filter)
	if err != nil { return domain.ExploreReport{}, err }

	tokensComparable := previous.Tokens.Total > 0
	callsComparable := previous.Calls > 0
	cacheComparable := promptTokens(current.Tokens) > 0 && promptTokens(previous.Tokens) > 0
	cacheDelta := float64(0)
	if cacheComparable { cacheDelta = math.Round((current.CacheReadShare-previous.CacheReadShare)*10)/10 }

	return domain.ExploreReport{
		GeneratedAt: now,
		Filter: filter,
		BucketMinutes: bucketMinutes,
		Current: current,
		Previous: previous,
		Comparison: domain.CompareReport{
			TokensChangePercent: percentDelta(current.Tokens.Total, previous.Tokens.Total),
			CallsChangePercent: percentDelta(int64(current.Calls), int64(previous.Calls)),
			CacheChangePoints: cacheDelta,
			TokensComparable: tokensComparable,
			CallsComparable: callsComparable,
			CacheComparable: cacheComparable,
			Direction: compareDirection(current.Tokens.Total, previous.Tokens.Total),
		},
		Timeline: timeline,
		Providers: providers,
		Projects: projects,
		Models: models,
		Sessions: sessions,
		Recent: recent,
		Facets: facets,
	}, nil
}

func (s *Store) exploreSummary(filter domain.ExploreFilter, start, end time.Time) (domain.WindowSummary, error) {
	where, args := exploreWhere(filter, start, end)
	query := `SELECT COUNT(*),COUNT(DISTINCT provider||char(31)||project),COUNT(DISTINCT provider||char(31)||project||char(31)||session),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),COALESCE(SUM(cache_creation_tokens),0),COALESCE(SUM(cache_read_tokens),0),COALESCE(SUM(cache_creation_5m_tokens),0),COALESCE(SUM(cache_creation_1h_tokens),0),COALESCE(SUM(total_tokens),0) FROM usage_calls WHERE ` + where
	result := domain.WindowSummary{Start:start, End:end}
	if err := s.db.QueryRow(query,args...).Scan(&result.Calls,&result.Projects,&result.Sessions,&result.Tokens.Input,&result.Tokens.Output,&result.Tokens.CacheCreation,&result.Tokens.CacheRead,&result.Tokens.CacheCreation5m,&result.Tokens.CacheCreation1h,&result.Tokens.Total); err != nil { return result, err }
	prompt := promptTokens(result.Tokens)
	if prompt > 0 { result.CacheReadShare = float64(result.Tokens.CacheRead)/float64(prompt)*100 }
	return result,nil
}

func (s *Store) exploreTimeline(filter domain.ExploreFilter, start,end time.Time,bucket time.Duration)([]domain.TimelineBucket,error){
	count:=int(math.Ceil(float64(end.Sub(start))/float64(bucket)));if count<1{count=1}
	items:=make([]domain.TimelineBucket,count);for i:=range items{items[i].Start=start.Add(time.Duration(i)*bucket)}
	where,args:=exploreWhere(filter,start,end)
	rows,err:=s.db.Query(`SELECT timestamp_ms,input_tokens,output_tokens,cache_creation_tokens,cache_read_tokens,cache_creation_5m_tokens,cache_creation_1h_tokens,total_tokens FROM usage_calls WHERE `+where+` ORDER BY timestamp_ms`,args...);if err!=nil{return nil,err};defer rows.Close()
	for rows.Next(){var ms int64;var t domain.TokenTotals;if err:=rows.Scan(&ms,&t.Input,&t.Output,&t.CacheCreation,&t.CacheRead,&t.CacheCreation5m,&t.CacheCreation1h,&t.Total);err!=nil{return nil,err};idx:=int(time.UnixMilli(ms).Sub(start)/bucket);if idx<0||idx>=len(items){continue};items[idx].Calls++;domain.AddTokens(&items[idx].Tokens,t)}
	return items,rows.Err()
}

func (s *Store) exploreDimension(filter domain.ExploreFilter,start,end time.Time,kind string,total int64,limit int)([]domain.DimensionActivity,error){
	where,args:=exploreWhere(filter,start,end);args=append(args,limit)
	var selectExpr,groupExpr string
	switch kind{case"provider":selectExpr="provider,provider";groupExpr="provider";case"project":selectExpr="provider,project";groupExpr="provider,project";case"model":selectExpr="provider,model";groupExpr="provider,model";default:return nil,fmt.Errorf("unsupported dimension %q",kind)}
	query:=fmt.Sprintf(`SELECT %s,COUNT(*),COUNT(DISTINCT project||char(31)||session),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),COALESCE(SUM(cache_creation_tokens),0),COALESCE(SUM(cache_read_tokens),0),COALESCE(SUM(cache_creation_5m_tokens),0),COALESCE(SUM(cache_creation_1h_tokens),0),COALESCE(SUM(total_tokens),0),MAX(timestamp_ms) FROM usage_calls WHERE %s GROUP BY %s ORDER BY SUM(total_tokens) DESC LIMIT ?`,selectExpr,where,groupExpr)
	rows,err:=s.db.Query(query,args...);if err!=nil{return nil,err};defer rows.Close();items:=make([]domain.DimensionActivity,0)
	for rows.Next(){var raw string;var last sql.NullInt64;var item domain.DimensionActivity;if err:=rows.Scan(&item.Provider,&raw,&item.Calls,&item.Sessions,&item.Tokens.Input,&item.Tokens.Output,&item.Tokens.CacheCreation,&item.Tokens.CacheRead,&item.Tokens.CacheCreation5m,&item.Tokens.CacheCreation1h,&item.Tokens.Total,&last);err!=nil{return nil,err};switch kind{case"provider":item.Name=providerDisplayName(raw);case"project":item.Name=displayProject(raw);case"model":item.Name=displayModel(item.Provider,raw)};item.SharePercent=globalShare(item.Tokens.Total,total);if last.Valid{item.LastActivity=time.UnixMilli(last.Int64)};items=append(items,item)}
	return items,rows.Err()
}

func (s *Store) exploreSessions(filter domain.ExploreFilter,start,end time.Time,total int64,limit int)([]domain.SessionActivity,error){
	where,args:=exploreWhere(filter,start,end);args=append(args,limit)
	rows,err:=s.db.Query(`SELECT provider,project,session,COUNT(*),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),COALESCE(SUM(cache_creation_tokens),0),COALESCE(SUM(cache_read_tokens),0),COALESCE(SUM(cache_creation_5m_tokens),0),COALESCE(SUM(cache_creation_1h_tokens),0),COALESCE(SUM(total_tokens),0),MIN(timestamp_ms),MAX(timestamp_ms),MAX(subagent) FROM usage_calls WHERE `+where+` GROUP BY provider,project,session ORDER BY SUM(total_tokens) DESC LIMIT ?`,args...);if err!=nil{return nil,err};defer rows.Close();items:=make([]domain.SessionActivity,0)
	for rows.Next(){var first,last int64;var sub int;var item domain.SessionActivity;if err:=rows.Scan(&item.Provider,&item.Project,&item.Session,&item.Calls,&item.Tokens.Input,&item.Tokens.Output,&item.Tokens.CacheCreation,&item.Tokens.CacheRead,&item.Tokens.CacheCreation5m,&item.Tokens.CacheCreation1h,&item.Tokens.Total,&first,&last,&sub);err!=nil{return nil,err};item.Project=displayProject(item.Project);item.FirstActivity=time.UnixMilli(first);item.LastActivity=time.UnixMilli(last);item.Subagent=sub!=0;item.SharePercent=globalShare(item.Tokens.Total,total);items=append(items,item)}
	return items,rows.Err()
}

func (s *Store) exploreRecent(filter domain.ExploreFilter,start,end time.Time,limit int)([]domain.CallActivity,error){
	where,args:=exploreWhere(filter,start,end);args=append(args,limit)
	rows,err:=s.db.Query(`SELECT provider,project,session,model,timestamp_ms,input_tokens,output_tokens,cache_creation_tokens,cache_read_tokens,cache_creation_5m_tokens,cache_creation_1h_tokens,total_tokens,subagent FROM usage_calls WHERE `+where+` ORDER BY timestamp_ms DESC LIMIT ?`,args...);if err!=nil{return nil,err};defer rows.Close();items:=make([]domain.CallActivity,0)
	for rows.Next(){var ms int64;var sub int;var rawModel string;var item domain.CallActivity;if err:=rows.Scan(&item.Provider,&item.Project,&item.Session,&rawModel,&ms,&item.Tokens.Input,&item.Tokens.Output,&item.Tokens.CacheCreation,&item.Tokens.CacheRead,&item.Tokens.CacheCreation5m,&item.Tokens.CacheCreation1h,&item.Tokens.Total,&sub);err!=nil{return nil,err};item.Project=displayProject(item.Project);item.Model=displayModel(item.Provider,rawModel);item.Timestamp=time.UnixMilli(ms);item.Subagent=sub!=0;items=append(items,item)}
	return items,rows.Err()
}

func (s *Store) exploreFacets(filter domain.ExploreFilter)(domain.ExploreFacets,error){
	facets:=domain.ExploreFacets{}
	providers,err:=s.distinctStrings(`SELECT DISTINCT provider FROM usage_calls WHERE total_tokens>0 ORDER BY provider`,nil);if err!=nil{return facets,err};facets.Providers=providers
	projectWhere:="total_tokens>0";projectArgs:=[]any{};if filter.Provider!=""{projectWhere+=" AND provider=?";projectArgs=append(projectArgs,filter.Provider)}
	projects,err:=s.distinctStrings(`SELECT DISTINCT project FROM usage_calls WHERE `+projectWhere+` AND TRIM(project)<>'' AND project<>'unknown' ORDER BY project`,projectArgs);if err!=nil{return facets,err};facets.Projects=uniqueSorted(projects)
	modelWhere:=projectWhere;modelArgs:=append([]any{},projectArgs...);if filter.Project!=""{modelWhere+=" AND project=?";modelArgs=append(modelArgs,filter.Project)};if filter.Session!=""{modelWhere+=" AND session=?";modelArgs=append(modelArgs,filter.Session)}
	models,err:=s.distinctStrings(`SELECT DISTINCT model FROM usage_calls WHERE `+modelWhere+` AND TRIM(model)<>'' AND model NOT LIKE '<%>' ORDER BY model`,modelArgs);if err!=nil{return facets,err};facets.Models=models
	return facets,nil
}

func (s *Store) distinctStrings(query string,args []any)([]string,error){rows,err:=s.db.Query(query,args...);if err!=nil{return nil,err};defer rows.Close();items:=[]string{};for rows.Next(){var value string;if err:=rows.Scan(&value);err!=nil{return nil,err};items=append(items,value)};return items,rows.Err()}

func exploreWhere(filter domain.ExploreFilter,start,end time.Time)(string,[]any){parts:=[]string{"timestamp_ms>=?","timestamp_ms<?","total_tokens>0"};args:=[]any{start.UnixMilli(),end.UnixMilli()};if filter.Provider!=""{parts=append(parts,"provider=?");args=append(args,filter.Provider)};if filter.Project!=""{parts=append(parts,"project=?");args=append(args,filter.Project)};if filter.Session!=""{parts=append(parts,"session=?");args=append(args,filter.Session)};if filter.Model!=""{parts=append(parts,"model=?");args=append(args,filter.Model)};return strings.Join(parts," AND "),args}
func exploreBucketMinutes(minutes int)int{switch{case minutes<=120:return 5;case minutes<=720:return 30;case minutes<=2880:return 60;default:return 360}}
func promptTokens(tokens domain.TokenTotals)int64{return tokens.Input+tokens.CacheCreation+tokens.CacheRead}
func percentDelta(current,previous int64)float64{if previous==0{return 0};return math.Round((float64(current-previous)/float64(previous)*100)*10)/10}
func compareDirection(current,previous int64)string{if previous==0&&current>0{return"new"};if current>previous{return"up"};if current<previous{return"down"};return"steady"}
func providerDisplayName(value string)string{switch value{case"claude":return"Claude Code";case"codex":return"Codex";case"gemini":return"Gemini CLI";default:return value}}
func uniqueSorted(values []string)[]string{seen:=map[string]bool{};out:=[]string{};for _,v:=range values{v=strings.TrimSpace(v);if v==""||seen[v]{continue};seen[v]=true;out=append(out,v)};sort.Strings(out);return out}
