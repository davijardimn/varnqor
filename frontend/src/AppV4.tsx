import {useCallback,useEffect,useMemo,useRef,useState,type ReactNode} from 'react'
import {ProviderBadge,ProviderMark,providerName} from './providerBrand'
import {languageOptions,v4Translator} from './v4i18n'

type View='control'|'missions'|'investigate'|'optimize'|'policies'|'agents'|'settings'
type IconName=View|'search'|'sun'|'moon'|'system'|'copy'|'refresh'|'chevron'|'pulse'|'shield'|'close'|'command'|'arrow'|'check'|'warning'

const ranges=[{m:60,l:'1h'},{m:360,l:'6h'},{m:1440,l:'24h'},{m:10080,l:'7d'}]
const emptyFilter:ExploreFilter={provider:'',project:'',session:'',model:'',rangeMinutes:60,limit:80}

export default function AppV4(){
  const [view,setView]=useState<View>('control')
  const [filter,setFilter]=useState<ExploreFilter>(emptyFilter)
  const [report,setReport]=useState<ControlPlaneResponse|null>(null)
  const [settings,setSettings]=useState<VarnqorSettings|null>(null)
  const [doctor,setDoctor]=useState<DoctorReport|null>(null)
  const [syncing,setSyncing]=useState(false)
  const [saving,setSaving]=useState(false)
  const [error,setError]=useState('')
  const [copied,setCopied]=useState('')
  const [palette,setPalette]=useState(false)
  const [paletteQuery,setPaletteQuery]=useState('')
  const requestRef=useRef(0)

  const t=useMemo(()=>v4Translator(settings?.language||'system'),[settings?.language])
  const app=()=>window.go?.main?.App

  const refresh=useCallback(async(silent=false)=>{
    const api=app();if(!api?.GetControlPlane)return
    const request=++requestRef.current
    if(!silent)setSyncing(true)
    try{
      const value=await api.GetControlPlane(filter)
      if(request===requestRef.current){setReport(value);setError('')}
    }catch(err){if(request===requestRef.current)setError(String(err))}
    finally{if(request===requestRef.current)setSyncing(false)}
  },[filter])

  useEffect(()=>{
    const api=app();if(!api)return
    api.GetSettings?.().then(setSettings).catch(err=>setError(String(err)))
  },[])

  useEffect(()=>{refresh();const id=window.setInterval(()=>refresh(true),(settings?.refreshSeconds||5)*1000);return()=>window.clearInterval(id)},[refresh,settings?.refreshSeconds])

  useEffect(()=>{
    if(!settings)return
    const media=window.matchMedia('(prefers-color-scheme: dark)')
    const apply=()=>{const next=settings.theme==='system'?(media.matches?'dark':'light'):settings.theme;document.documentElement.dataset.theme=next;document.documentElement.style.colorScheme=next}
    apply();media.addEventListener?.('change',apply);return()=>media.removeEventListener?.('change',apply)
  },[settings?.theme])

  useEffect(()=>{
    const down=(event:KeyboardEvent)=>{
      if((event.ctrlKey||event.metaKey)&&event.key.toLowerCase()==='k'){event.preventDefault();setPalette(v=>!v);setPaletteQuery('')}
      if(event.key==='Escape')setPalette(false)
    }
    window.addEventListener('keydown',down);return()=>window.removeEventListener('keydown',down)
  },[])

  const save=async(next=settings)=>{
    const api=app();if(!next||!api?.SaveSettings)return
    setSaving(true)
    try{const saved=await api.SaveSettings(next);setSettings(saved);setError('');setCopied('saved');window.setTimeout(()=>setCopied(''),1300);await refresh(true)}catch(err){setError(String(err))}finally{setSaving(false)}
  }

  const copyText=async(text:string,key='copied')=>{try{await navigator.clipboard.writeText(text);setCopied(key);window.setTimeout(()=>setCopied(''),1500)}catch(err){setError(String(err))}}
  const copyBrief=async()=>{const api=app();if(!api?.ExportControlBrief)return;try{const text=await api.ExportControlBrief(filter);await copyText(text,'brief')}catch(err){setError(String(err))}}
  const copyExplore=async(format:'json'|'csv')=>{const api=app();if(!api?.ExportExplore)return;try{const text=await api.ExportExplore(filter,format);await copyText(text,format)}catch(err){setError(String(err))}}

  const focusMission=(mission:ControlMission,target:View='investigate')=>{setFilter(f=>({...f,provider:mission.provider,project:mission.project,session:mission.session,model:''}));setView(target)}
  const clearScope=()=>setFilter(f=>({...f,provider:'',project:'',session:'',model:''}))
  const setRange=(rangeMinutes:number)=>setFilter(f=>({...f,rangeMinutes}))

  const nav: {id:View;icon:IconName;label:string}[]=[
    {id:'control',icon:'control',label:t('nav.control')},{id:'missions',icon:'missions',label:t('nav.missions')},{id:'investigate',icon:'investigate',label:t('nav.investigate')},
    {id:'optimize',icon:'optimize',label:t('nav.optimize')},{id:'policies',icon:'policies',label:t('nav.policies')},{id:'agents',icon:'agents',label:t('nav.agents')},{id:'settings',icon:'settings',label:t('nav.settings')},
  ]
  const meta:Record<View,[string,string]>={
    control:[t('control.title'),t('control.sub')],missions:[t('missions.title'),t('missions.sub')],investigate:[t('investigate.title'),t('investigate.sub')],optimize:[t('optimize.title'),t('optimize.sub')],policies:[t('policies.title'),t('policies.sub')],agents:[t('agents.title'),t('agents.sub')],settings:[t('settings.title'),t('settings.sub')]
  }
  const scope=[filter.provider&&providerName(filter.provider),filter.project,filter.session&&shortID(filter.session),filter.model].filter(Boolean) as string[]

  return <div className="v4Shell">
    <aside className="appRail">
      <div className="railSpacer"/>
      <nav className="railNav">{nav.map(item=><button key={item.id} className={view===item.id?'active':''} onClick={()=>setView(item.id)} title={item.label} aria-label={item.label}><Icon name={item.icon}/><span>{item.label}</span>{item.id==='control'&&!!report?.control.highIncidents&&<i className="railAlert"/>}</button>)}</nav>
      <div className="railBottom">
        <button onClick={()=>{setPalette(true);setPaletteQuery('')}} title="Command palette"><Icon name="command"/></button>
        <div className={`railLive ${error?'bad':syncing?'sync':''}`} title={error||t('common.live')}><i/></div>
      </div>
    </aside>

    <main className="workspace">
      <header className="topbar">
        <div className="pageIdentity"><h1>{meta[view][0]}</h1><p>{meta[view][1]}</p></div>
        <div className="topbarCenter">{scope.length>0&&<button className="scopeChip" onClick={clearScope} title={t('investigate.clear')}><span>{scope.join(' / ')}</span><Icon name="close"/></button>}</div>
        <div className="topbarTools">
          <div className="rangeSwitch">{ranges.map(r=><button key={r.m} className={filter.rangeMinutes===r.m?'active':''} onClick={()=>setRange(r.m)}>{r.l}</button>)}</div>
          <button className="toolButton" onClick={()=>refresh()} title={t('common.live')}><Icon name="refresh" className={syncing?'spin':''}/></button>
          <ThemeButton settings={settings} onChange={next=>{if(settings){const value={...settings,theme:next};setSettings(value);save(value)}}}/>
          <div className={`liveBadge ${error?'bad':syncing?'sync':''}`}><i/><span>{syncing?t('common.syncing'):t('common.live')}</span></div>
        </div>
      </header>

      <div className="workspaceBody">
        {error&&<div className="errorBar"><Icon name="warning"/><span>{error}</span><button onClick={()=>setError('')}><Icon name="close"/></button></div>}
        {!report?<Loading/>:<>
          {view==='control'&&<ControlRoom report={report} t={t} onMission={focusMission} onBrief={copyBrief} briefCopied={copied==='brief'}/>} 
          {view==='missions'&&<Missions report={report} t={t} onFocus={focusMission}/>} 
          {view==='investigate'&&<Investigate report={report} filter={filter} setFilter={setFilter} t={t} onCopy={copyExplore} copied={copied}/>} 
          {view==='optimize'&&<Optimize report={report} settings={settings} setSettings={setSettings} save={save} t={t} onBrief={copyBrief} briefCopied={copied==='brief'} onPolicies={()=>setView('policies')}/>} 
          {view==='policies'&&<Policies report={report} settings={settings} setSettings={setSettings} save={save} saving={saving} saved={copied==='saved'} t={t}/>} 
          {view==='agents'&&<Agents report={report} t={t}/>} 
          {view==='settings'&&<SettingsView report={report} settings={settings} setSettings={setSettings} save={save} saving={saving} saved={copied==='saved'} doctor={doctor} setDoctor={setDoctor} t={t}/>} 
        </>}
      </div>
    </main>
    {palette&&<CommandPalette query={paletteQuery} setQuery={setPaletteQuery} nav={nav} t={t} onView={id=>{setView(id);setPalette(false)}} onRange={m=>{setRange(m);setPalette(false)}} onClose={()=>setPalette(false)}/>} 
  </div>
}

function ControlRoom({report,t,onMission,onBrief,briefCopied}:{report:ControlPlaneResponse;t:(k:string)=>string;onMission:(m:ControlMission,v?:View)=>void;onBrief:()=>void;briefCopied:boolean}){
  const c=report.control,e=report.explore,s=report.status
  const incidents=c.incidents??[],decisions=c.decisions??[],missions=c.missions??[],providers=s.providers??[]
  return <div className="page controlPage">
    <section className="commandStrip">
      <div className={`commandState state-${c.state}`}><div className="stateGlyph"><ControlGlyph state={c.state}/></div><div><span>{t('control.state')}</span><strong>{stateLabel(c.state,t)}</strong><small>{c.highIncidents>0?`${c.highIncidents} high-priority incident${c.highIncidents===1?'':'s'}`:c.mediumIncidents>0?`${c.mediumIncidents} policy condition${c.mediumIncidents===1?'':'s'} to review`:t('control.externalOff')}</small></div></div>
      <Metric label={t('control.pace')} value={`${formatTokens(c.currentPerHour)} ${t('common.perHour')}`} sub={s.burn.ready?`${s.burn.ratio.toFixed(2)}× baseline`:t('common.learning')}/>
      <Metric label={t('control.missions')} value={String(c.activeMissions)} sub={`${c.coolingMissions} ${t('missions.cooling').toLowerCase()}`}/>
      <Metric label={t('control.providers')} value={String(c.activeProviders)} sub={`${providers.length} known`}/>
      <Metric label={t('control.projection')} value={formatTokens(c.projectedTokens)} sub={`${c.policy.decisionHorizonMinutes} ${t('common.minutes')}`}/>
      <button className="briefButton" onClick={onBrief}><Icon name={briefCopied?'check':'copy'}/><span>{briefCopied?t('control.briefCopied'):t('control.copyBrief')}</span></button>
    </section>

    <section className="controlMainGrid">
      <Panel title={t('control.flow')} sub={t('control.flowSub')} className="flowPanel"><Timeline data={e.timeline??[]} baseline={s.burn.baselinePerHour} rangeMinutes={e.filter.rangeMinutes}/><div className="flowLegend"><span><i className="legendPrimary"/>Current window</span><span><i className="legendBaseline"/>Learned baseline</span><span>{formatTokens(e.current.tokens.total)} tokens · {e.current.calls} calls · {e.current.cacheReadShare.toFixed(1)}% cache read</span></div></Panel>
      <Panel title={t('control.decisions')} sub={t('control.decisionsSub')} className="decisionPanel" action={<span className="panelCount">{decisions.length}</span>}>
        {decisions.length===0?<Empty icon="check" text={t('control.noDecision')}/>:<div className="decisionQueue">{decisions.slice(0,5).map((d,i)=><DecisionRow key={d.id} item={d} index={i}/>)}</div>}
      </Panel>
    </section>

    <section className="controlLowerGrid">
      <Panel title={t('control.incidents')} sub={t('control.incidentsSub')} action={<span className={`signalTotal ${c.highIncidents?'hot':''}`}>{c.highIncidents+c.mediumIncidents+c.lowIncidents}</span>}>
        {incidents.length===0?<Empty icon="shield" text={t('control.noIncident')}/>:<div className="incidentList">{incidents.slice(0,6).map(item=><IncidentRow key={item.id} item={item}/>)}</div>}
      </Panel>
      <Panel title={t('control.fleet')} sub={t('control.fleetSub')}><div className="fleetList">{providers.map(p=><div className="fleetRow" key={p.id}><ProviderMark id={p.id} size={22}/><div className="fleetIdentity"><strong>{p.name}</strong><span><i className={`stateDot state-${p.state}`}/>{providerState(p.state,t)}</span></div><div className="fleetMetric"><strong>{formatTokens(p.tokensToday)}</strong><span>tokens</span></div><div className="fleetMetric"><strong>{p.observedSessions}</strong><span>sessions</span></div></div>)}</div></Panel>
      <Panel title={t('control.missionBoard')} sub={t('control.missionBoardSub')}><div className="missionMiniList">{missions.length===0?<Empty icon="missions" text={t('missions.none')}/>:missions.slice(0,5).map(m=><button key={m.id} onClick={()=>onMission(m,'missions')}><ProviderMark id={m.provider} size={17}/><div><strong>{cleanProject(m.project)}</strong><span>{shortID(m.session)} · {m.state}</span></div><div className="missionShare"><strong>{m.sharePercent.toFixed(1)}%</strong><i><b style={{width:`${Math.min(100,m.sharePercent)}%`}}/></i></div><span className={`pressure pressure-${m.pressure}`}>{m.pressure}</span></button>)}</div></Panel>
    </section>

    <section className="controlFoot">
      <div><span>Policy</span><strong>{c.policy.profile}</strong><small>Burn ceiling {c.policy.burnCeilingRatio.toFixed(2)}× · session {c.policy.sessionDominancePercent.toFixed(0)}% · project {c.policy.projectDominancePercent.toFixed(0)}%</small></div>
      <div><span>Runway</span><strong>{s.quota.available?(s.quota.runwayAvailable?`${s.quota.runwayHours.toFixed(1)}h`:`${s.quota.utilizationPercent.toFixed(1)}% used`):'No trusted limit'}</strong><small>{s.quota.available?s.quota.label:'Configure a local budget to enable runway'}</small></div>
      <div><span>Execution boundary</span><strong>Manual / advisory</strong><small>No silent prompt rewriting, model switching or external agent execution.</small></div>
    </section>
  </div>
}

function Missions({report,t,onFocus}:{report:ControlPlaneResponse;t:(k:string)=>string;onFocus:(m:ControlMission,v?:View)=>void}){
  const missions=report.control.missions??[]
  const groups:[string,string,ControlMission[]][]=[['active',t('missions.active'),missions.filter(m=>m.state==='active')],['cooling',t('missions.cooling'),missions.filter(m=>m.state==='cooling')],['idle',t('missions.idle'),missions.filter(m=>m.state==='idle')]]
  return <div className="page missionsPage">
    <div className="missionSummary"><Metric label={t('missions.active')} value={String(report.control.activeMissions)} sub={`${formatTokens(missions.filter(m=>m.state==='active').reduce((a,m)=>a+m.tokens.total,0))} tokens`}/><Metric label={t('missions.cooling')} value={String(report.control.coolingMissions)} sub={`≤ ${report.control.policy.missionCoolingMinutes} min`}/><Metric label={t('missions.pressure')} value={(missions[0]?.pressure||'normal').toUpperCase()} sub={missions[0]?`${missions[0].sharePercent.toFixed(1)}% top share`:'No active mission'}/><Metric label={t('control.pace')} value={`${formatTokens(report.control.currentPerHour)}/h`} sub={`${report.explore.current.calls} calls`}/></div>
    <div className="missionColumns">{groups.map(([state,label,items])=><section key={state} className={`missionColumn missionColumn-${state}`}><header><div><i/><strong>{label}</strong><span>{items.length}</span></div><small>{state==='active'?`≤ ${report.control.policy.missionActiveMinutes} min`:state==='cooling'?`≤ ${report.control.policy.missionCoolingMinutes} min`:'outside cooling window'}</small></header><div className="missionCards">{items.length===0?<Empty icon="missions" text={t('missions.none')}/>:items.map(m=><article className="missionCard" key={m.id}><div className="missionCardHead"><ProviderBadge id={m.provider}/><span className={`pressure pressure-${m.pressure}`}>{m.pressure}</span></div><h3>{cleanProject(m.project)}</h3><code>{shortID(m.session)}</code><div className="missionBar"><i><b style={{width:`${Math.min(100,m.sharePercent)}%`}}/></i><span>{m.sharePercent.toFixed(1)}%</span></div><div className="missionStats"><div><span>{t('missions.velocity')}</span><strong>{formatTokens(m.velocityHour)}/h</strong></div><div><span>{t('missions.calls')}</span><strong>{m.calls}</strong></div><div><span>{t('missions.last')}</span><strong>{relativeAge(m.ageMinutes)}</strong></div></div><div className="missionActions"><button onClick={()=>onFocus(m,'investigate')}>{t('missions.focus')} <Icon name="arrow"/></button><button onClick={()=>onFocus(m,'optimize')}>Optimize</button></div></article>)}</div></section>)}</div>
  </div>
}

function Investigate({report,filter,setFilter,t,onCopy,copied}:{report:ControlPlaneResponse;filter:ExploreFilter;setFilter:(v:ExploreFilter|((p:ExploreFilter)=>ExploreFilter))=>void;t:(k:string)=>string;onCopy:(f:'json'|'csv')=>void;copied:string}){
  const e=report.explore
  return <div className="page investigatePage">
    <section className="filterDock"><span className="filterDockLabel"><Icon name="search"/>{t('investigate.filters')}</span><Select value={filter.provider} onChange={v=>setFilter(f=>({...f,provider:v,session:''}))} options={['',...(e.facets.providers??[])]} label={t('common.provider')}/><Select value={filter.project} onChange={v=>setFilter(f=>({...f,project:v,session:''}))} options={['',...(e.facets.projects??[])]} label={t('common.project')}/><Select value={filter.model} onChange={v=>setFilter(f=>({...f,model:v,session:''}))} options={['',...(e.facets.models??[])]} label={t('common.model')}/><div className="dockSpacer"/><button className="dockButton" onClick={()=>setFilter(f=>({...f,provider:'',project:'',session:'',model:''}))}>{t('investigate.clear')}</button><button className="dockButton" onClick={()=>onCopy('json')}><Icon name={copied==='json'?'check':'copy'}/>{t('investigate.copyJson')}</button><button className="dockButton" onClick={()=>onCopy('csv')}><Icon name={copied==='csv'?'check':'copy'}/>{t('investigate.copyCsv')}</button></section>
    <section className="compareStrip"><Compare label="Tokens" value={formatTokens(e.current.tokens.total)} delta={e.comparison.tokensChangePercent} comparable={e.comparison.tokensComparable}/><Compare label="Calls" value={String(e.current.calls)} delta={e.comparison.callsChangePercent} comparable={e.comparison.callsComparable}/><Compare label={t('investigate.cache')} value={`${e.current.cacheReadShare.toFixed(1)}%`} delta={e.comparison.cacheChangePoints} comparable={e.comparison.cacheComparable} points/><Compare label="Sessions" value={String(e.current.sessions)} sub={`${e.current.projects} projects`}/></section>
    <div className="investigateGrid"><Panel title={t('investigate.timeline')} sub={`${e.bucketMinutes} min buckets`} className="investigateTimeline"><Timeline data={e.timeline??[]} rangeMinutes={e.filter.rangeMinutes}/></Panel><Panel title={t('investigate.projects')}><DimensionRows items={e.projects??[]}/></Panel><Panel title={t('investigate.models')}><DimensionRows items={e.models??[]}/></Panel></div>
    <Panel title={t('investigate.recent')} sub={`${e.recent?.length||0} token-bearing calls`}><RecentTable items={e.recent??[]}/></Panel>
  </div>
}

function Optimize({report,settings,setSettings,save,t,onBrief,briefCopied,onPolicies}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:(v:VarnqorSettings|null|((p:VarnqorSettings|null)=>VarnqorSettings|null))=>void;save:(v?:VarnqorSettings|null)=>Promise<void>;t:(k:string)=>string;onBrief:()=>void;briefCopied:boolean;onPolicies:()=>void}){
  const o=report.optimization,c=report.control,actions=o.actions??[]
  const updateOpt=(key:keyof OptimizationSettings,value:string|number|boolean)=>setSettings(s=>s?({...s,optimization:{...s.optimization,[key]:value}}):s)
  return <div className="page optimizePage">
    <section className="scenarioGrid"><Scenario label={t('optimize.projected')} value={formatTokens(o.scenario.projectedTokens)} sub={`${o.scenario.horizonMinutes} min`}/><Scenario label={t('optimize.target')} value={formatTokens(o.scenario.targetTokens)} sub={o.scenario.feasibility}/><Scenario label={t('optimize.reduction')} value={`${o.scenario.targetReductionPercent.toFixed(0)}%`} sub={`~${formatTokens(o.scenario.targetSavings)} tokens`}/><Scenario label="Control state" value={stateLabel(c.state,t)} sub={`${c.highIncidents} high · ${c.mediumIncidents} medium`}/></section>
    <section className="optimizerDock"><div className="optimizerField"><span>Goal</span><select value={settings?.optimization.goal||'balanced'} onChange={e=>updateOpt('goal',e.target.value)}><option value="balanced">Balanced</option><option value="economy">Economy</option><option value="throughput">Throughput</option></select></div><div className="optimizerField targetField"><span>{t('optimize.reduction')}</span><input type="range" min="5" max="60" step="5" value={settings?.optimization.targetReductionPercent||20} onChange={e=>updateOpt('targetReductionPercent',Number(e.target.value))}/><b>{settings?.optimization.targetReductionPercent||20}%</b></div><div className="optimizerField"><span>{t('optimize.horizon')}</span><select value={settings?.optimization.horizonMinutes||120} onChange={e=>updateOpt('horizonMinutes',Number(e.target.value))}><option value={60}>1h</option><option value={120}>2h</option><option value={360}>6h</option><option value={720}>12h</option><option value={1440}>24h</option></select></div><button className="primaryAction" disabled={!settings} onClick={()=>save()}>{t('common.save')}</button><button className="secondaryAction" onClick={onPolicies}>{t('optimize.configure')}</button><button className="secondaryAction" onClick={onBrief}><Icon name={briefCopied?'check':'copy'}/>{t('optimize.copyBrief')}</button></section>
    <div className="executionBoundary"><Icon name="shield"/><div><strong>{t('control.externalOff')}</strong><span>Varnqor creates evidence-backed interventions, but does not rewrite prompts, switch models, terminate sessions or execute provider tools silently.</span></div></div>
    <section className="interventionGrid">{actions.length===0?<div className="emptyLarge"><Icon name="check"/><strong>{t('optimize.noActions')}</strong><span>The current window remains inside the evidence thresholds for bounded optimization.</span></div>:actions.map((a,i)=><article className={`interventionCard priority-${a.priority}`} key={a.id}><header><span>{String(i+1).padStart(2,'0')}</span><div><small>{a.kind}</small><h3>{a.title}</h3></div><b>{a.priority}</b></header><p>{a.summary}</p><div className="evidenceBox"><span>Evidence</span><p>{a.evidence}</p></div>{a.estimatedSavingsHigh>0&&<div className="impactRow"><div><span>{t('optimize.impact')}</span><strong>~{formatTokens(a.estimatedSavingsLow)}–{formatTokens(a.estimatedSavingsHigh)}</strong></div><small>{a.confidence} {t('optimize.confidence')} · {t('optimize.nonAdditive')}</small></div>}<div className="stepList"><span>{t('optimize.steps')}</span>{(a.steps??[]).map((step,n)=><div key={n}><i>{n+1}</i><p>{step}</p></div>)}</div>{(a.provider||a.project||a.session)&&<footer>{a.provider&&<ProviderBadge id={a.provider}/>} {a.project&&<span>{cleanProject(a.project)}</span>} {a.session&&<code>{shortID(a.session)}</code>}</footer>}</article>)}</section>
  </div>
}

function Policies({report,settings,setSettings,save,saving,saved,t}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:(v:VarnqorSettings|null|((p:VarnqorSettings|null)=>VarnqorSettings|null))=>void;save:(v?:VarnqorSettings|null)=>Promise<void>;saving:boolean;saved:boolean;t:(k:string)=>string}){
  if(!settings)return <Loading/>
  const update=(key:keyof ControlSettings,value:string|number|boolean)=>setSettings(s=>s?({...s,control:{...s.control,[key]:value}}):s)
  const profile=(name:'balanced'|'economy'|'throughput')=>setSettings(s=>s?({...s,control:{...s.control,...profileValues(name)}}):s)
  const c=settings.control
  return <div className="page policiesPage">
    <Panel title={t('policies.profile')} sub={t('policies.profileSub')}><div className="profileCards">{(['balanced','economy','throughput'] as const).map(name=>{const p=profileValues(name);return <button key={name} className={c.profile===name?'active':''} onClick={()=>profile(name)}><div className="profileTop"><span>{t(`policies.${name}`)}</span>{c.profile===name&&<Icon name="check"/>}</div><strong>{p.burnCeilingRatio.toFixed(2)}×</strong><small>burn ceiling</small><div><span>Session {p.sessionDominancePercent}%</span><span>Project {p.projectDominancePercent}%</span><span>Budget {p.budgetWarningPercent}%</span></div></button>})}</div></Panel>
    <div className="policyLayout"><Panel title="Control thresholds" sub={`Current state: ${stateLabel(report.control.state,t)}`}><div className="policyRows"><PolicyNumber label={t('policies.burn')} value={c.burnCeilingRatio} min={1} max={10} step={.05} suffix="×" onChange={v=>update('burnCeilingRatio',v)}/><PolicyNumber label={t('policies.session')} value={c.sessionDominancePercent} min={20} max={100} step={1} suffix="%" onChange={v=>update('sessionDominancePercent',v)}/><PolicyNumber label={t('policies.project')} value={c.projectDominancePercent} min={20} max={100} step={1} suffix="%" onChange={v=>update('projectDominancePercent',v)}/><PolicyNumber label={t('policies.cache')} value={c.cacheFloorPercent} min={0} max={100} step={1} suffix="%" onChange={v=>update('cacheFloorPercent',v)}/><PolicyNumber label={t('policies.budget')} value={c.budgetWarningPercent} min={10} max={100} step={1} suffix="%" onChange={v=>update('budgetWarningPercent',v)}/><PolicyNumber label={t('policies.minimum')} value={c.minimumEvaluationTokens} min={100} max={10000000} step={100} suffix=" tokens" onChange={v=>update('minimumEvaluationTokens',v)}/></div></Panel><Panel title="Mission lifecycle" sub="How recent activity becomes an operational mission"><div className="policyRows"><PolicyNumber label={t('policies.active')} value={c.missionActiveMinutes} min={1} max={120} step={1} suffix=" min" onChange={v=>update('missionActiveMinutes',v)}/><PolicyNumber label={t('policies.cooling')} value={c.missionCoolingMinutes} min={c.missionActiveMinutes} max={720} step={1} suffix=" min" onChange={v=>update('missionCoolingMinutes',v)}/><PolicyNumber label={t('policies.horizon')} value={c.decisionHorizonMinutes} min={30} max={1440} step={30} suffix=" min" onChange={v=>update('decisionHorizonMinutes',v)}/><div className="policySwitch"><div><strong>Control Plane</strong><span>Evaluate missions, incidents and decisions locally</span></div><button className={c.enabled?'on':''} onClick={()=>update('enabled',!c.enabled)}><i/></button></div></div></Panel></div>
    <Panel title={t('policies.guardrails')} sub={t('policies.guardrailsSub')}><div className="guardrailStrip"><div><span>Spike</span><strong>{settings.guardrails.spikePercent}%</strong></div><div><span>Project</span><strong>{settings.guardrails.projectConcentrationPercent}%</strong></div><div><span>Session</span><strong>{settings.guardrails.sessionConcentrationPercent}%</strong></div><div><span>Cache floor</span><strong>{settings.guardrails.cacheReadMinimumPercent}%</strong></div><div><span>Minimum</span><strong>{formatTokens(settings.guardrails.minimumTokens)}</strong></div></div></Panel>
    <div className="saveDock"><div><Icon name="shield"/><span>Every threshold is local and explainable. External execution remains disabled.</span></div><button className="primaryAction" disabled={saving} onClick={()=>save()}><Icon name={saved?'check':'shield'}/>{saved?t('policies.saved'):saving?'Saving…':t('policies.save')}</button></div>
  </div>
}

function Agents({report,t}:{report:ControlPlaneResponse;t:(k:string)=>string}){
  const providers=report.status.providers??[]
  return <div className="page agentsPage"><div className="agentCards">{providers.map(p=><article key={p.id} className={`agentCard agent-${p.state}`}><header><div className="agentLogo"><ProviderMark id={p.id} size={34}/></div><div><h2>{p.name}</h2><span><i className={`stateDot state-${p.state}`}/>{providerState(p.state,t)}</span></div><b>{p.capabilities.tools?'deep':'usage'}</b></header><div className="agentNumbers"><div><span>{t('agents.sessions')}</span><strong>{p.observedSessions}</strong></div><div><span>{t('agents.calls')}</span><strong>{p.indexedCalls}</strong></div><div><span>{t('agents.tokens')}</span><strong>{formatTokens(p.tokensToday)}</strong></div></div><div className="capabilities"><span>{t('agents.capabilities')}</span><div><b className={p.capabilities.usage?'yes':''}>{t('agents.usage')}</b><b className={p.capabilities.tools?'yes':''}>{t('agents.tools')}</b><b className={p.capabilities.limits?'yes':''}>{t('agents.limits')}</b></div></div><div className="agentMeta"><div><span>{t('agents.last')}</span><strong>{p.lastActivity?formatDateTime(p.lastActivity):'—'}</strong></div><div><span>Local data</span><code>{p.dataPath||'—'}</code></div>{p.note&&<p>{p.note}</p>}{p.error&&<p className="agentError">{p.error}</p>}</div></article>)}</div></div>
}

function SettingsView({report,settings,setSettings,save,saving,saved,doctor,setDoctor,t}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:(v:VarnqorSettings|null|((p:VarnqorSettings|null)=>VarnqorSettings|null))=>void;save:(v?:VarnqorSettings|null)=>Promise<void>;saving:boolean;saved:boolean;doctor:DoctorReport|null;setDoctor:(v:DoctorReport|null)=>void;t:(k:string)=>string}){
  if(!settings)return <Loading/>
  const update=<K extends keyof VarnqorSettings>(key:K,value:VarnqorSettings[K])=>setSettings(s=>s?({...s,[key]:value}):s)
  const runDoctor=async()=>{const api=window.go?.main?.App;if(!api?.RunDoctor)return;setDoctor(await api.RunDoctor())}
  return <div className="page settingsPage"><div className="settingsGrid"><Panel title={t('settings.appearance')}><div className="settingRows"><label><span>{t('settings.theme')}</span><select value={settings.theme} onChange={e=>update('theme',e.target.value)}><option value="dark">{t('settings.theme.dark')}</option><option value="light">{t('settings.theme.light')}</option><option value="system">{t('settings.theme.system')}</option></select></label><label><span>{t('settings.language')}</span><select value={settings.language} onChange={e=>update('language',e.target.value)}>{languageOptions.map(l=><option key={l} value={l}>{t(`language.${l}`)}</option>)}</select></label></div></Panel><Panel title={t('settings.monitoring')}><div className="settingRows"><label><span>{t('settings.refresh')}</span><div className="numberUnit"><input type="number" min="2" max="60" value={settings.refreshSeconds} onChange={e=>update('refreshSeconds',Number(e.target.value))}/><b>sec</b></div></label></div></Panel><Panel title={t('settings.budget')} sub={t('settings.budgetSub')}><div className="settingRows"><label className="toggleRow"><span>{t('settings.enabled')}</span><button className={`switch ${settings.budget.enabled?'on':''}`} onClick={()=>update('budget',{...settings.budget,enabled:!settings.budget.enabled})}><i/></button></label><label><span>{t('settings.limit')}</span><input type="number" min="0" step="1000" value={settings.budget.limitTokens} onChange={e=>update('budget',{...settings.budget,limitTokens:Number(e.target.value)})}/></label><label><span>{t('settings.reset')}</span><div className="numberUnit"><input type="number" min="0" max="23" value={settings.budget.resetHour} onChange={e=>update('budget',{...settings.budget,resetHour:Number(e.target.value)})}/><b>:00</b></div></label></div></Panel><Panel title={t('settings.diagnostics')}><div className="doctorBlock"><button className="secondaryAction" onClick={runDoctor}>{t('settings.doctor')}</button>{doctor?<div className={`doctorResult doctor-${doctor.status}`}><strong>{doctor.status.toUpperCase()}</strong><span>{doctor.pass} pass · {doctor.warn} warn · {doctor.fail} fail</span></div>:<span className="mutedText">Not checked in this session</span>}</div></Panel><Panel title={t('settings.about')}><div className="aboutRows"><div><span>{t('settings.version')}</span><strong>{report.status.version}</strong></div><div><span>Storage</span><strong>Local SQLite</strong></div><div><span>Control</span><strong>{settings.control.enabled?'Local decision support':'Disabled'}</strong></div><div><span>Privacy</span><strong>{t('settings.privacy')}</strong></div></div></Panel></div><div className="saveDock"><div><span>V4 settings schema · stored locally</span></div><button className="primaryAction" disabled={saving} onClick={()=>save()}><Icon name={saved?'check':'settings'}/>{saved?t('settings.saved'):saving?'Saving…':t('settings.save')}</button></div></div>
}

function CommandPalette({query,setQuery,nav,t,onView,onRange,onClose}:{query:string;setQuery:(v:string)=>void;nav:{id:View;icon:IconName;label:string}[];t:(k:string)=>string;onView:(v:View)=>void;onRange:(m:number)=>void;onClose:()=>void}){
  const q=query.trim().toLowerCase();const items=nav.filter(n=>!q||n.label.toLowerCase().includes(q));const rangeItems=ranges.filter(r=>!q||r.l.includes(q)||'range period'.includes(q))
  return <div className="paletteOverlay" onMouseDown={e=>{if(e.target===e.currentTarget)onClose()}}><div className="commandPalette"><header><Icon name="search"/><input autoFocus value={query} onChange={e=>setQuery(e.target.value)} placeholder="Go to a control surface or set a range…"/><kbd>Esc</kbd></header><div className="paletteBody"><small>GO TO</small>{items.map(item=><button key={item.id} onClick={()=>onView(item.id)}><Icon name={item.icon}/><span>{item.label}</span><Icon name="arrow"/></button>)}<small>RANGE</small>{rangeItems.map(r=><button key={r.m} onClick={()=>onRange(r.m)}><Icon name="pulse"/><span>{r.l} window</span><b>{r.m} min</b></button>)}</div><footer><span>Varnqor Control Plane</span><kbd>Ctrl K</kbd></footer></div></div>
}

function Panel({title,sub,action,className='',children}:{title:string;sub?:string;action?:ReactNode;className?:string;children:ReactNode}){return <section className={`panel ${className}`}><header className="panelHeader"><div><h2>{title}</h2>{sub&&<p>{sub}</p>}</div>{action}</header>{children}</section>}
function Metric({label,value,sub}:{label:string;value:string;sub?:string}){return <div className="metric"><span>{label}</span><strong>{value}</strong>{sub&&<small>{sub}</small>}</div>}
function Scenario({label,value,sub}:{label:string;value:string;sub:string}){return <div className="scenario"><span>{label}</span><strong>{value}</strong><small>{sub}</small></div>}
function Empty({icon,text}:{icon:IconName;text:string}){return <div className="empty"><Icon name={icon}/><span>{text}</span></div>}

function DecisionRow({item,index}:{item:ControlDecision;index:number}){return <article className={`decisionRow priority-${item.priority}`}><span className="decisionIndex">{String(index+1).padStart(2,'0')}</span><div><div className="decisionMeta"><b>{item.priority}</b><small>{item.kind}</small></div><strong>{item.title}</strong><p>{item.reason}</p>{(item.provider||item.project)&&<footer>{item.provider&&<ProviderBadge id={item.provider}/>} {item.project&&<span>{cleanProject(item.project)}</span>}</footer>}</div>{item.estimatedSavingsHigh>0&&<aside><span>impact</span><strong>{formatTokens(item.estimatedSavingsLow)}–{formatTokens(item.estimatedSavingsHigh)}</strong></aside>}</article>}
function IncidentRow({item}:{item:ControlIncident}){return <article className={`incidentRow severity-${item.severity}`}><i/><div><div><b>{item.severity}</b><span>{item.source}</span></div><strong>{item.title}</strong><p>{item.detail}</p></div>{item.value>0&&<aside><strong>{formatIncidentValue(item)}</strong><span>/ {formatThreshold(item)}</span></aside>}</article>}
function DimensionRows({items}:{items:DimensionActivity[]}){return <div className="dimensionRows">{items.slice(0,8).map((item,i)=><div key={`${item.provider}-${item.name}-${i}`}><span>{item.provider&&<ProviderMark id={item.provider} size={14}/>}<strong>{cleanProject(item.name)}</strong></span><b>{formatTokens(item.tokens.total)}</b><em>{item.sharePercent.toFixed(1)}%</em><i><b style={{width:`${Math.min(100,item.sharePercent)}%`}}/></i></div>)}</div>}
function RecentTable({items}:{items:CallActivity[]}){return <div className="recentTable"><div className="recentHead"><span>Time</span><span>Agent</span><span>Project</span><span>Model</span><span>Tokens</span></div>{items.slice(0,40).map((item,i)=><div className="recentRow" key={`${item.timestamp}-${i}`}><span>{formatClock(item.timestamp)}</span><span><ProviderBadge id={item.provider}/></span><span title={item.project}>{cleanProject(item.project)}</span><span title={item.model}>{item.model||'—'}</span><strong>{formatTokens(item.tokens.total)}</strong></div>)}</div>}
function Compare({label,value,delta,comparable=true,points=false,sub}:{label:string;value:string;delta?:number;comparable?:boolean;points?:boolean;sub?:string}){const d=delta??0;return <div className="compare"><span>{label}</span><strong>{value}</strong>{sub?<small>{sub}</small>:comparable?<small className={d>0?'up':d<0?'down':''}>{d>0?'+':''}{d.toFixed(1)}{points?' pts':'%'}</small>:<small>new window</small>}</div>}
function Select({value,onChange,options,label}:{value:string;onChange:(v:string)=>void;options:string[];label:string}){return <label className="selectWrap"><span>{label}</span><select value={value} onChange={e=>onChange(e.target.value)}>{options.map((o,i)=><option key={`${o}-${i}`} value={o}>{o||'All'}</option>)}</select></label>}
function PolicyNumber({label,value,min,max,step,suffix,onChange}:{label:string;value:number;min:number;max:number;step:number;suffix:string;onChange:(v:number)=>void}){return <label className="policyNumber"><div><strong>{label}</strong><span>{min}–{max}{suffix}</span></div><div><input type="number" value={value} min={min} max={max} step={step} onChange={e=>onChange(Number(e.target.value))}/><b>{suffix}</b></div></label>}

function Timeline({data,baseline=0,rangeMinutes}:{data:TimelineBucket[];baseline?:number;rangeMinutes:number}){
  const max=Math.max(1,...data.map(d=>d.tokens.total));const baselineBucket=baseline>0&&data.length?baseline*(rangeMinutes/data.length)/60:0;const baselinePct=Math.min(100,baselineBucket/max*100)
  return <div className="timeline"><div className="timelineGrid"><i/><i/><i/><i/></div>{baseline>0&&<div className="baselineLine" style={{bottom:`calc(30px + ${baselinePct*.72}%)`}}><span>{formatTokens(baseline)}/h baseline</span></div>}<div className="timelineBars">{data.map((bucket,i)=>{const h=Math.max(2,bucket.tokens.total/max*100);return <div className="timelineSlot" key={`${bucket.start}-${i}`} title={`${formatDateTime(bucket.start)} · ${formatTokens(bucket.tokens.total)} tokens`}><i style={{height:`${h}%`}}/><span>{i===0||i===data.length-1||i%Math.max(1,Math.floor(data.length/5))===0?formatTimelineLabel(bucket.start,rangeMinutes):''}</span></div>})}</div></div>
}

function ThemeButton({settings,onChange}:{settings:VarnqorSettings|null;onChange:(v:string)=>void}){const current=settings?.theme||'dark';const next=current==='dark'?'light':current==='light'?'system':'dark';return <button className="toolButton" onClick={()=>onChange(next)} title={`Theme: ${current}`}><Icon name={current==='dark'?'moon':current==='light'?'sun':'system'}/></button>}
function ControlGlyph({state}:{state:string}){return <span className={`controlGlyph glyph-${state}`}><i/><b/><em/></span>}
function Loading(){return <div className="loading"><div className="loadingMark"><i/><i/><i/></div><span>Building local control picture…</span></div>}

function Icon({name,className=''}:{name:IconName;className?:string}){
 const paths:Record<IconName,ReactNode>={
  control:<><path d="M4 5h16v14H4z"/><path d="M8 9h3v6H8zM14 7h2v8h-2z"/></>,missions:<><circle cx="12" cy="12" r="8"/><path d="M12 7v5l3 2"/></>,investigate:<><circle cx="10.5" cy="10.5" r="6"/><path d="m15 15 5 5"/></>,optimize:<><path d="M4 17h3l3-5 3 2 3-7 4 2"/><path d="M17 5h3v3"/></>,policies:<><path d="M12 3 5 6v5c0 4.4 2.8 7.8 7 10 4.2-2.2 7-5.6 7-10V6z"/><path d="m9 12 2 2 4-5"/></>,agents:<><circle cx="8" cy="9" r="3"/><circle cx="16.5" cy="10" r="2.5"/><path d="M3 19c.8-3 2.5-4.5 5-4.5S12.3 16 13 19M14 15c2.8-.6 5 .8 6 3"/></>,settings:<><circle cx="12" cy="12" r="3"/><path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6 7 7M17 17l1.4 1.4M18.4 5.6 17 7M7 17l-1.4 1.4"/></>,search:<><circle cx="10.5" cy="10.5" r="6"/><path d="m15 15 5 5"/></>,sun:<><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M19.1 4.9l-1.4 1.4M6.3 17.7l-1.4 1.4"/></>,moon:<path d="M19 15.5A8 8 0 0 1 8.5 5 8 8 0 1 0 19 15.5Z"/>,system:<><rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8M12 17v4"/></>,copy:<><rect x="8" y="8" width="11" height="11" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/></>,refresh:<><path d="M20 6v5h-5"/><path d="M18.5 15a7 7 0 1 1-.7-7.7L20 11"/></>,chevron:<path d="m9 6 6 6-6 6"/>,pulse:<path d="M3 12h4l2-6 4 12 2-6h6"/>,shield:<><path d="M12 3 5 6v5c0 4.4 2.8 7.8 7 10 4.2-2.2 7-5.6 7-10V6z"/><path d="m9 12 2 2 4-5"/></>,close:<path d="m7 7 10 10M17 7 7 17"/>,command:<><rect x="4" y="4" width="6" height="6" rx="2"/><rect x="14" y="4" width="6" height="6" rx="2"/><rect x="4" y="14" width="6" height="6" rx="2"/><rect x="14" y="14" width="6" height="6" rx="2"/></>,arrow:<path d="M5 12h14M14 7l5 5-5 5"/>,check:<path d="m5 12 4 4L19 6"/>,warning:<><path d="M12 3 2.8 20h18.4z"/><path d="M12 9v5M12 17h.01"/></>
 }
 return <svg className={className} viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{paths[name]}</svg>
}

function profileValues(name:'balanced'|'economy'|'throughput'):Partial<ControlSettings>{if(name==='economy')return{profile:name,burnCeilingRatio:1.45,sessionDominancePercent:55,projectDominancePercent:65,cacheFloorPercent:20,budgetWarningPercent:72,missionActiveMinutes:10,missionCoolingMinutes:40,decisionHorizonMinutes:120,minimumEvaluationTokens:800};if(name==='throughput')return{profile:name,burnCeilingRatio:2.25,sessionDominancePercent:78,projectDominancePercent:85,cacheFloorPercent:8,budgetWarningPercent:92,missionActiveMinutes:15,missionCoolingMinutes:60,decisionHorizonMinutes:180,minimumEvaluationTokens:1500};return{profile:'balanced',burnCeilingRatio:1.7,sessionDominancePercent:62,projectDominancePercent:72,cacheFloorPercent:15,budgetWarningPercent:80,missionActiveMinutes:12,missionCoolingMinutes:45,decisionHorizonMinutes:120,minimumEvaluationTokens:1000}}
function formatTokens(value:number){const n=Math.abs(value);if(n>=1e9)return`${(value/1e9).toFixed(n>=1e10?1:2)}B`;if(n>=1e6)return`${(value/1e6).toFixed(n>=1e7?1:2)}M`;if(n>=1e3)return`${(value/1e3).toFixed(n>=1e4?1:2)}K`;return Math.round(value).toLocaleString()}
function shortID(value:string){if(!value)return'—';return value.length<=14?value:`${value.slice(0,7)}…${value.slice(-4)}`}
function cleanProject(value:string){if(!value)return'Unknown';const normalized=value.replace(/\\/g,'/');const parts=normalized.split('/').filter(Boolean);return parts[parts.length-1]||value}
function relativeAge(minutes:number){if(minutes<1)return'now';if(minutes<60)return`${minutes}m`;if(minutes<1440)return`${Math.floor(minutes/60)}h`;return`${Math.floor(minutes/1440)}d`}
function formatDateTime(value:string){if(!value)return'—';const d=new Date(value);return Number.isNaN(d.getTime())?'—':d.toLocaleString([], {month:'short',day:'2-digit',hour:'2-digit',minute:'2-digit'})}
function formatClock(value:string){const d=new Date(value);return Number.isNaN(d.getTime())?'—':d.toLocaleTimeString([], {hour:'2-digit',minute:'2-digit',second:'2-digit'})}
function formatTimelineLabel(value:string,range:number){const d=new Date(value);if(Number.isNaN(d.getTime()))return'';return range>=1440?d.toLocaleDateString([],{month:'short',day:'numeric'}):d.toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'})}
function stateLabel(state:string,t:(k:string)=>string){return t(`common.${state}`)===`common.${state}`?state.charAt(0).toUpperCase()+state.slice(1):t(`common.${state}`)}
function providerState(state:string,t:(k:string)=>string){const key=`agents.${state}`;const translated=t(key);return translated===key?state:translated}
function formatIncidentValue(i:ControlIncident){if(i.kind.includes('dominance')||i.kind==='cache-floor'||i.kind==='budget-risk')return`${i.value.toFixed(1)}%`;if(i.kind==='burn-ceiling')return`${i.value.toFixed(2)}×`;return String(i.value)}
function formatThreshold(i:ControlIncident){if(i.kind.includes('dominance')||i.kind==='cache-floor'||i.kind==='budget-risk')return`${i.threshold.toFixed(0)}%`;if(i.kind==='burn-ceiling')return`${i.threshold.toFixed(2)}×`;return String(i.threshold)}
