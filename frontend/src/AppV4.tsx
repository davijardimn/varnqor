import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type Dispatch,
  type ReactNode,
  type SetStateAction,
} from 'react'
import {ProviderBadge,ProviderMark,providerName} from './providerBrand'
import {languageOptions,v4Translator} from './v4i18n'

type View='control'|'missions'|'investigate'|'optimize'|'policies'|'agents'|'settings'
type IconName=View|'search'|'sun'|'moon'|'system'|'copy'|'refresh'|'pulse'|'shield'|'close'|'command'|'arrow'|'check'|'warning'
type Translate=(key:string)=>string
type SettingsSetter=Dispatch<SetStateAction<VarnqorSettings|null>>

type RangeOption={m:number;l:string}
const ranges:RangeOption[]=[{m:60,l:'1h'},{m:360,l:'6h'},{m:1440,l:'24h'},{m:10080,l:'7d'}]
const initialFilter:ExploreFilter={provider:'',project:'',session:'',model:'',rangeMinutes:60,limit:80}

export default function AppV4(){
  const [view,setView]=useState<View>('control')
  const [filter,setFilter]=useState<ExploreFilter>(initialFilter)
  const [report,setReport]=useState<ControlPlaneResponse|null>(null)
  const [settings,setSettings]=useState<VarnqorSettings|null>(null)
  const [doctor,setDoctor]=useState<DoctorReport|null>(null)
  const [syncing,setSyncing]=useState(false)
  const [saving,setSaving]=useState(false)
  const [error,setError]=useState('')
  const [copied,setCopied]=useState('')
  const [palette,setPalette]=useState(false)
  const [query,setQuery]=useState('')
  const requestRef=useRef(0)

  const t=useMemo(()=>v4Translator(settings?.language||'system'),[settings?.language])
  const api=()=>window.go?.main?.App

  const refresh=useCallback(async(silent=false)=>{
    const bridge=api()
    if(!bridge?.GetControlPlane)return
    const request=++requestRef.current
    if(!silent)setSyncing(true)
    try{
      const value=await bridge.GetControlPlane(filter)
      if(request===requestRef.current){setReport(value);setError('')}
    }catch(reason){if(request===requestRef.current)setError(String(reason))}
    finally{if(request===requestRef.current)setSyncing(false)}
  },[filter])

  useEffect(()=>{
    const bridge=api()
    bridge?.GetSettings?.().then(setSettings).catch(reason=>setError(String(reason)))
  },[])

  useEffect(()=>{
    void refresh()
    const timer=window.setInterval(()=>void refresh(true),(settings?.refreshSeconds||5)*1000)
    return()=>window.clearInterval(timer)
  },[refresh,settings?.refreshSeconds])

  useEffect(()=>{
    if(!settings)return
    const media=window.matchMedia('(prefers-color-scheme: dark)')
    const apply=()=>{
      const resolved=settings.theme==='system'?(media.matches?'dark':'light'):settings.theme
      document.documentElement.dataset.theme=resolved
      document.documentElement.style.colorScheme=resolved
    }
    apply()
    media.addEventListener?.('change',apply)
    return()=>media.removeEventListener?.('change',apply)
  },[settings])

  useEffect(()=>{
    const onKey=(event:KeyboardEvent)=>{
      if((event.ctrlKey||event.metaKey)&&event.key.toLowerCase()==='k'){
        event.preventDefault();setPalette(value=>!value);setQuery('')
      }
      if(event.key==='Escape')setPalette(false)
    }
    window.addEventListener('keydown',onKey)
    return()=>window.removeEventListener('keydown',onKey)
  },[])

  const save=async(value:VarnqorSettings|null=settings)=>{
    const bridge=api()
    if(!value||!bridge?.SaveSettings)return
    setSaving(true)
    try{
      const saved=await bridge.SaveSettings(value)
      setSettings(saved);setError('');setCopied('saved')
      window.setTimeout(()=>setCopied(''),1400)
      await refresh(true)
    }catch(reason){setError(String(reason))}
    finally{setSaving(false)}
  }

  const copy=async(text:string,key:string)=>{
    try{await navigator.clipboard.writeText(text);setCopied(key);window.setTimeout(()=>setCopied(''),1500)}
    catch(reason){setError(String(reason))}
  }
  const copyBrief=async()=>{
    const bridge=api();if(!bridge?.ExportControlBrief)return
    try{await copy(await bridge.ExportControlBrief(filter),'brief')}catch(reason){setError(String(reason))}
  }
  const copyExplore=async(format:'json'|'csv')=>{
    const bridge=api();if(!bridge?.ExportExplore)return
    try{await copy(await bridge.ExportExplore(filter,format),format)}catch(reason){setError(String(reason))}
  }
  const focusMission=(mission:ControlMission,target:View)=>{
    setFilter(current=>({...current,provider:mission.provider,project:mission.project,session:mission.session,model:''}))
    setView(target)
  }
  const clearScope=()=>setFilter(current=>({...current,provider:'',project:'',session:'',model:''}))

  const navigation:{id:View;icon:IconName;label:string}[]=[
    {id:'control',icon:'control',label:t('nav.control')},
    {id:'missions',icon:'missions',label:t('nav.missions')},
    {id:'investigate',icon:'investigate',label:t('nav.investigate')},
    {id:'optimize',icon:'optimize',label:t('nav.optimize')},
    {id:'policies',icon:'policies',label:t('nav.policies')},
    {id:'agents',icon:'agents',label:t('nav.agents')},
    {id:'settings',icon:'settings',label:t('nav.settings')},
  ]
  const pageMeta:Record<View,[string,string]>={
    control:[t('control.title'),t('control.sub')],missions:[t('missions.title'),t('missions.sub')],
    investigate:[t('investigate.title'),t('investigate.sub')],optimize:[t('optimize.title'),t('optimize.sub')],
    policies:[t('policies.title'),t('policies.sub')],agents:[t('agents.title'),t('agents.sub')],settings:[t('settings.title'),t('settings.sub')],
  }
  const scope=[filter.provider&&providerName(filter.provider),filter.project,filter.session&&shortID(filter.session),filter.model].filter(Boolean) as string[]

  return <div className="v4Shell">
    <aside className="appRail">
      <div className="railSpacer"/>
      <nav className="railNav">{navigation.map(item=><button key={item.id} className={view===item.id?'active':''} onClick={()=>setView(item.id)} title={item.label} aria-label={item.label}><Icon name={item.icon}/><span>{item.label}</span>{item.id==='control'&&!!report?.control.highIncidents&&<i className="railAlert"/>}</button>)}</nav>
      <div className="railBottom"><button onClick={()=>{setPalette(true);setQuery('')}} title="Command palette"><Icon name="command"/></button><div className={`railLive ${error?'bad':syncing?'sync':''}`}><i/></div></div>
    </aside>

    <main className="workspace">
      <header className="topbar">
        <div className="pageIdentity"><h1>{pageMeta[view][0]}</h1><p>{pageMeta[view][1]}</p></div>
        <div className="topbarCenter">{scope.length>0&&<button className="scopeChip" onClick={clearScope}><span>{scope.join(' / ')}</span><Icon name="close"/></button>}</div>
        <div className="topbarTools">
          <div className="rangeSwitch">{ranges.map(range=><button key={range.m} className={filter.rangeMinutes===range.m?'active':''} onClick={()=>setFilter(current=>({...current,rangeMinutes:range.m}))}>{range.l}</button>)}</div>
          <button className="toolButton" onClick={()=>void refresh()} title="Refresh"><Icon name="refresh" className={syncing?'spin':''}/></button>
          <ThemeButton settings={settings} setSettings={setSettings} save={save}/>
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

    {palette&&<CommandPalette query={query} setQuery={setQuery} navigation={navigation} onView={next=>{setView(next);setPalette(false)}} onRange={minutes=>{setFilter(current=>({...current,rangeMinutes:minutes}));setPalette(false)}} onClose={()=>setPalette(false)}/>} 
  </div>
}

function ControlRoom({report,t,onMission,onBrief,briefCopied}:{report:ControlPlaneResponse;t:Translate;onMission:(mission:ControlMission,target:View)=>void;onBrief:()=>void;briefCopied:boolean}){
  const {control,explore,status}=report
  const decisions=control.decisions??[],incidents=control.incidents??[],missions=control.missions??[],providers=status.providers??[]
  const stateDetail=control.highIncidents>0?`${control.highIncidents} high-priority incident${control.highIncidents===1?'':'s'}`:control.mediumIncidents>0?`${control.mediumIncidents} condition${control.mediumIncidents===1?'':'s'} to review`:t('control.externalOff')
  return <div className="page controlPage">
    <section className="commandStrip">
      <div className={`commandState state-${control.state}`}><div className="stateGlyph"><ControlGlyph state={control.state}/></div><div><span>{t('control.state')}</span><strong>{stateLabel(control.state,t)}</strong><small>{stateDetail}</small></div></div>
      <Metric label={t('control.pace')} value={`${formatTokens(control.currentPerHour)}${t('common.perHour')}`} sub={status.burn.ready?`${status.burn.ratio.toFixed(2)}× baseline`:t('common.learning')}/>
      <Metric label={t('control.missions')} value={String(control.activeMissions)} sub={`${control.coolingMissions} ${t('missions.cooling').toLowerCase()}`}/>
      <Metric label={t('control.providers')} value={String(control.activeProviders)} sub={`${providers.length} known`}/>
      <Metric label={t('control.projection')} value={formatTokens(control.projectedTokens)} sub={`${control.policy.decisionHorizonMinutes} ${t('common.minutes')}`}/>
      <button className="briefButton" onClick={onBrief}><Icon name={briefCopied?'check':'copy'}/><span>{briefCopied?t('control.briefCopied'):t('control.copyBrief')}</span></button>
    </section>

    <section className="controlMainGrid">
      <Panel title={t('control.flow')} sub={t('control.flowSub')} className="flowPanel"><Timeline data={explore.timeline??[]} baseline={status.burn.baselinePerHour} rangeMinutes={explore.filter.rangeMinutes}/><div className="flowLegend"><span><i className="legendPrimary"/>Current window</span><span><i className="legendBaseline"/>Learned baseline</span><span>{formatTokens(explore.current.tokens.total)} tokens · {explore.current.calls} calls · {explore.current.cacheReadShare.toFixed(1)}% cache read</span></div></Panel>
      <Panel title={t('control.decisions')} sub={t('control.decisionsSub')} className="decisionPanel" action={<span className="panelCount">{decisions.length}</span>}>{decisions.length===0?<Empty icon="check" text={t('control.noDecision')}/>:<div className="decisionQueue">{decisions.slice(0,5).map((decision,index)=><DecisionRow key={decision.id} item={decision} index={index}/>)}</div>}</Panel>
    </section>

    <section className="controlLowerGrid">
      <Panel title={t('control.incidents')} sub={t('control.incidentsSub')} action={<span className={`signalTotal ${control.highIncidents?'hot':''}`}>{control.highIncidents+control.mediumIncidents+control.lowIncidents}</span>}>{incidents.length===0?<Empty icon="shield" text={t('control.noIncident')}/>:<div className="incidentList">{incidents.slice(0,6).map(incident=><IncidentRow key={incident.id} item={incident}/>)}</div>}</Panel>
      <Panel title={t('control.fleet')} sub={t('control.fleetSub')}><div className="fleetList">{providers.map(provider=><div className="fleetRow" key={provider.id}><ProviderMark id={provider.id} size={22}/><div className="fleetIdentity"><strong>{provider.name}</strong><span><i className={`stateDot state-${provider.state}`}/>{providerState(provider.state,t)}</span></div><div className="fleetMetric"><strong>{formatTokens(provider.tokensToday)}</strong><span>tokens</span></div><div className="fleetMetric"><strong>{provider.observedSessions}</strong><span>sessions</span></div></div>)}</div></Panel>
      <Panel title={t('control.missionBoard')} sub={t('control.missionBoardSub')}><div className="missionMiniList">{missions.length===0?<Empty icon="missions" text={t('missions.none')}/>:missions.slice(0,5).map(mission=><button key={mission.id} onClick={()=>onMission(mission,'missions')}><ProviderMark id={mission.provider} size={17}/><div><strong>{cleanProject(mission.project)}</strong><span>{shortID(mission.session)} · {mission.state}</span></div><div className="missionShare"><strong>{mission.sharePercent.toFixed(1)}%</strong><i><b style={{width:`${Math.min(100,mission.sharePercent)}%`}}/></i></div><span className={`pressure pressure-${mission.pressure}`}>{mission.pressure}</span></button>)}</div></Panel>
    </section>

    <section className="controlFoot">
      <div><span>Policy</span><strong>{control.policy.profile}</strong><small>Burn {control.policy.burnCeilingRatio.toFixed(2)}× · session {control.policy.sessionDominancePercent.toFixed(0)}% · project {control.policy.projectDominancePercent.toFixed(0)}%</small></div>
      <div><span>Runway</span><strong>{status.quota.available?(status.quota.runwayAvailable?`${status.quota.runwayHours.toFixed(1)}h`:`${status.quota.utilizationPercent.toFixed(1)}% used`):'No trusted limit'}</strong><small>{status.quota.available?status.quota.label:'Configure a local budget to enable runway'}</small></div>
      <div><span>Execution boundary</span><strong>Manual / advisory</strong><small>No silent prompt rewriting, model switching or external agent execution.</small></div>
    </section>
  </div>
}

function Missions({report,t,onFocus}:{report:ControlPlaneResponse;t:Translate;onFocus:(mission:ControlMission,target:View)=>void}){
  const missions=report.control.missions??[]
  const groups:{state:string;label:string;items:ControlMission[]}[]=[
    {state:'active',label:t('missions.active'),items:missions.filter(item=>item.state==='active')},
    {state:'cooling',label:t('missions.cooling'),items:missions.filter(item=>item.state==='cooling')},
    {state:'idle',label:t('missions.idle'),items:missions.filter(item=>item.state==='idle')},
  ]
  return <div className="page missionsPage">
    <div className="missionSummary"><Metric label={t('missions.active')} value={String(report.control.activeMissions)} sub={`${formatTokens(missions.filter(item=>item.state==='active').reduce((sum,item)=>sum+item.tokens.total,0))} tokens`}/><Metric label={t('missions.cooling')} value={String(report.control.coolingMissions)} sub={`≤ ${report.control.policy.missionCoolingMinutes} min`}/><Metric label={t('missions.pressure')} value={(missions[0]?.pressure||'normal').toUpperCase()} sub={missions[0]?`${missions[0].sharePercent.toFixed(1)}% top share`:'No active mission'}/><Metric label={t('control.pace')} value={`${formatTokens(report.control.currentPerHour)}/h`} sub={`${report.explore.current.calls} calls`}/></div>
    <div className="missionColumns">{groups.map(group=><section key={group.state} className={`missionColumn missionColumn-${group.state}`}><header><div><i/><strong>{group.label}</strong><span>{group.items.length}</span></div><small>{group.state==='active'?`≤ ${report.control.policy.missionActiveMinutes} min`:group.state==='cooling'?`≤ ${report.control.policy.missionCoolingMinutes} min`:'outside cooling window'}</small></header><div className="missionCards">{group.items.length===0?<Empty icon="missions" text={t('missions.none')}/>:group.items.map(mission=><article className="missionCard" key={mission.id}><div className="missionCardHead"><ProviderBadge id={mission.provider}/><span className={`pressure pressure-${mission.pressure}`}>{mission.pressure}</span></div><h3>{cleanProject(mission.project)}</h3><code>{shortID(mission.session)}</code><div className="missionBar"><i><b style={{width:`${Math.min(100,mission.sharePercent)}%`}}/></i><span>{mission.sharePercent.toFixed(1)}%</span></div><div className="missionStats"><div><span>{t('missions.velocity')}</span><strong>{formatTokens(mission.velocityHour)}/h</strong></div><div><span>{t('missions.calls')}</span><strong>{mission.calls}</strong></div><div><span>{t('missions.last')}</span><strong>{relativeAge(mission.ageMinutes)}</strong></div></div><div className="missionActions"><button onClick={()=>onFocus(mission,'investigate')}>{t('missions.focus')} <Icon name="arrow"/></button><button onClick={()=>onFocus(mission,'optimize')}>Optimize</button></div></article>)}</div></section>)}</div>
  </div>
}

function Investigate({report,filter,setFilter,t,onCopy,copied}:{report:ControlPlaneResponse;filter:ExploreFilter;setFilter:Dispatch<SetStateAction<ExploreFilter>>;t:Translate;onCopy:(format:'json'|'csv')=>void;copied:string}){
  const explore=report.explore
  return <div className="page investigatePage">
    <section className="filterDock"><span className="filterDockLabel"><Icon name="search"/>{t('investigate.filters')}</span><Select value={filter.provider} onChange={provider=>setFilter(current=>({...current,provider,session:''}))} options={['',...(explore.facets.providers??[])]} label={t('common.provider')}/><Select value={filter.project} onChange={project=>setFilter(current=>({...current,project,session:''}))} options={['',...(explore.facets.projects??[])]} label={t('common.project')}/><Select value={filter.model} onChange={model=>setFilter(current=>({...current,model,session:''}))} options={['',...(explore.facets.models??[])]} label={t('common.model')}/><div className="dockSpacer"/><button className="dockButton" onClick={()=>setFilter(current=>({...current,provider:'',project:'',session:'',model:''}))}>{t('investigate.clear')}</button><button className="dockButton" onClick={()=>onCopy('json')}><Icon name={copied==='json'?'check':'copy'}/>{t('investigate.copyJson')}</button><button className="dockButton" onClick={()=>onCopy('csv')}><Icon name={copied==='csv'?'check':'copy'}/>{t('investigate.copyCsv')}</button></section>
    <section className="compareStrip"><Compare label="Tokens" value={formatTokens(explore.current.tokens.total)} delta={explore.comparison.tokensChangePercent} comparable={explore.comparison.tokensComparable}/><Compare label="Calls" value={String(explore.current.calls)} delta={explore.comparison.callsChangePercent} comparable={explore.comparison.callsComparable}/><Compare label={t('investigate.cache')} value={`${explore.current.cacheReadShare.toFixed(1)}%`} delta={explore.comparison.cacheChangePoints} comparable={explore.comparison.cacheComparable} points/><Compare label="Sessions" value={String(explore.current.sessions)} sub={`${explore.current.projects} projects`}/></section>
    <div className="investigateGrid"><Panel title={t('investigate.timeline')} sub={`${explore.bucketMinutes} min buckets`} className="investigateTimeline"><Timeline data={explore.timeline??[]} rangeMinutes={explore.filter.rangeMinutes}/></Panel><Panel title={t('investigate.projects')}><DimensionRows items={explore.projects??[]}/></Panel><Panel title={t('investigate.models')}><DimensionRows items={explore.models??[]}/></Panel></div>
    <Panel title={t('investigate.recent')} sub={`${explore.recent?.length||0} token-bearing calls`}><RecentTable items={explore.recent??[]}/></Panel>
  </div>
}

function Optimize({report,settings,setSettings,save,t,onBrief,briefCopied,onPolicies}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:SettingsSetter;save:(value?:VarnqorSettings|null)=>Promise<void>;t:Translate;onBrief:()=>void;briefCopied:boolean;onPolicies:()=>void}){
  const optimization=report.optimization
  const actions=optimization.actions??[]
  const update=(changes:Partial<OptimizationSettings>)=>setSettings(current=>current?({...current,optimization:{...current.optimization,...changes}}):current)
  return <div className="page optimizePage">
    <section className="scenarioGrid"><Scenario label={t('optimize.projected')} value={formatTokens(optimization.scenario.projectedTokens)} sub={`${optimization.scenario.horizonMinutes} min`}/><Scenario label={t('optimize.target')} value={formatTokens(optimization.scenario.targetTokens)} sub={optimization.scenario.feasibility}/><Scenario label={t('optimize.reduction')} value={`${optimization.scenario.targetReductionPercent.toFixed(0)}%`} sub={`~${formatTokens(optimization.scenario.targetSavings)} tokens`}/><Scenario label="Control state" value={stateLabel(report.control.state,t)} sub={`${report.control.highIncidents} high · ${report.control.mediumIncidents} medium`}/></section>
    <section className="optimizerDock"><div className="optimizerField"><span>Goal</span><select value={settings?.optimization.goal||'balanced'} onChange={event=>update({goal:event.target.value})}><option value="balanced">Balanced</option><option value="economy">Economy</option><option value="throughput">Throughput</option></select></div><div className="optimizerField targetField"><span>{t('optimize.reduction')}</span><input type="range" min="5" max="60" step="5" value={settings?.optimization.targetReductionPercent||20} onChange={event=>update({targetReductionPercent:Number(event.target.value)})}/><b>{settings?.optimization.targetReductionPercent||20}%</b></div><div className="optimizerField"><span>{t('optimize.horizon')}</span><select value={settings?.optimization.horizonMinutes||120} onChange={event=>update({horizonMinutes:Number(event.target.value)})}><option value={60}>1h</option><option value={120}>2h</option><option value={360}>6h</option><option value={720}>12h</option><option value={1440}>24h</option></select></div><button className="primaryAction" disabled={!settings} onClick={()=>void save()}>{t('common.save')}</button><button className="secondaryAction" onClick={onPolicies}>{t('optimize.configure')}</button><button className="secondaryAction" onClick={onBrief}><Icon name={briefCopied?'check':'copy'}/>{t('optimize.copyBrief')}</button></section>
    <div className="executionBoundary"><Icon name="shield"/><div><strong>{t('control.externalOff')}</strong><span>Plans and evidence stay local; no silent prompt rewrite, model switch, session termination or external tool execution.</span></div></div>
    <section className="interventionGrid">{actions.length===0?<div className="emptyLarge"><Icon name="check"/><strong>{t('optimize.noActions')}</strong><span>The selected scope remains inside the bounded optimization evidence thresholds.</span></div>:actions.map((action,index)=><article className={`interventionCard priority-${action.priority}`} key={action.id}><header><span>{String(index+1).padStart(2,'0')}</span><div><small>{action.kind}</small><h3>{action.title}</h3></div><b>{action.priority}</b></header><p>{action.summary}</p><div className="evidenceBox"><span>Evidence</span><p>{action.evidence}</p></div>{action.estimatedSavingsHigh>0&&<div className="impactRow"><div><span>{t('optimize.impact')}</span><strong>~{formatTokens(action.estimatedSavingsLow)}–{formatTokens(action.estimatedSavingsHigh)}</strong></div><small>{action.confidence} {t('optimize.confidence')} · {t('optimize.nonAdditive')}</small></div>}<div className="stepList"><span>{t('optimize.steps')}</span>{(action.steps??[]).map((step,stepIndex)=><div key={stepIndex}><i>{stepIndex+1}</i><p>{step}</p></div>)}</div>{(action.provider||action.project||action.session)&&<footer>{action.provider&&<ProviderBadge id={action.provider}/>} {action.project&&<span>{cleanProject(action.project)}</span>} {action.session&&<code>{shortID(action.session)}</code>}</footer>}</article>)}</section>
  </div>
}

function Policies({report,settings,setSettings,save,saving,saved,t}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:SettingsSetter;save:(value?:VarnqorSettings|null)=>Promise<void>;saving:boolean;saved:boolean;t:Translate}){
  if(!settings)return <Loading/>
  const control=settings.control
  const update=(changes:Partial<ControlSettings>)=>setSettings(current=>current?({...current,control:{...current.control,...changes}}):current)
  const applyProfile=(name:'balanced'|'economy'|'throughput')=>update(controlProfile(name))
  return <div className="page policiesPage">
    <Panel title={t('policies.profile')} sub={t('policies.profileSub')}><div className="profileCards">{(['balanced','economy','throughput'] as const).map(name=>{const preset=controlProfile(name);return <button key={name} className={control.profile===name?'active':''} onClick={()=>applyProfile(name)}><div className="profileTop"><span>{t(`policies.${name}`)}</span>{control.profile===name&&<Icon name="check"/>}</div><strong>{preset.burnCeilingRatio.toFixed(2)}×</strong><small>burn ceiling</small><div><span>Session {preset.sessionDominancePercent}%</span><span>Project {preset.projectDominancePercent}%</span><span>Budget {preset.budgetWarningPercent}%</span></div></button>})}</div></Panel>
    <div className="policyLayout"><Panel title="Control thresholds" sub={`Current state: ${stateLabel(report.control.state,t)}`}><div className="policyRows"><PolicyNumber label={t('policies.burn')} value={control.burnCeilingRatio} min={1} max={10} step={.05} suffix="×" onChange={value=>update({burnCeilingRatio:value})}/><PolicyNumber label={t('policies.session')} value={control.sessionDominancePercent} min={20} max={100} step={1} suffix="%" onChange={value=>update({sessionDominancePercent:value})}/><PolicyNumber label={t('policies.project')} value={control.projectDominancePercent} min={20} max={100} step={1} suffix="%" onChange={value=>update({projectDominancePercent:value})}/><PolicyNumber label={t('policies.cache')} value={control.cacheFloorPercent} min={0} max={100} step={1} suffix="%" onChange={value=>update({cacheFloorPercent:value})}/><PolicyNumber label={t('policies.budget')} value={control.budgetWarningPercent} min={10} max={100} step={1} suffix="%" onChange={value=>update({budgetWarningPercent:value})}/><PolicyNumber label={t('policies.minimum')} value={control.minimumEvaluationTokens} min={100} max={10000000} step={100} suffix=" tokens" onChange={value=>update({minimumEvaluationTokens:value})}/></div></Panel><Panel title="Mission lifecycle" sub="Recent sessions become active, cooling, then idle missions"><div className="policyRows"><PolicyNumber label={t('policies.active')} value={control.missionActiveMinutes} min={1} max={120} step={1} suffix=" min" onChange={value=>update({missionActiveMinutes:value})}/><PolicyNumber label={t('policies.cooling')} value={control.missionCoolingMinutes} min={control.missionActiveMinutes} max={720} step={1} suffix=" min" onChange={value=>update({missionCoolingMinutes:value})}/><PolicyNumber label={t('policies.horizon')} value={control.decisionHorizonMinutes} min={30} max={1440} step={30} suffix=" min" onChange={value=>update({decisionHorizonMinutes:value})}/><div className="policySwitch"><div><strong>Control Plane</strong><span>Evaluate missions, incidents and decisions locally</span></div><button className={control.enabled?'on':''} onClick={()=>update({enabled:!control.enabled})}><i/></button></div></div></Panel></div>
    <Panel title={t('policies.guardrails')} sub={t('policies.guardrailsSub')}><div className="guardrailStrip"><div><span>Spike</span><strong>{settings.guardrails.spikePercent}%</strong></div><div><span>Project</span><strong>{settings.guardrails.projectConcentrationPercent}%</strong></div><div><span>Session</span><strong>{settings.guardrails.sessionConcentrationPercent}%</strong></div><div><span>Cache floor</span><strong>{settings.guardrails.cacheReadMinimumPercent}%</strong></div><div><span>Minimum</span><strong>{formatTokens(settings.guardrails.minimumTokens)}</strong></div></div></Panel>
    <div className="saveDock"><div><Icon name="shield"/><span>Every threshold is visible, local and deterministic. External execution remains disabled.</span></div><button className="primaryAction" disabled={saving} onClick={()=>void save()}><Icon name={saved?'check':'shield'}/>{saved?t('policies.saved'):saving?'Saving…':t('policies.save')}</button></div>
  </div>
}

function Agents({report,t}:{report:ControlPlaneResponse;t:Translate}){
  const providers=report.status.providers??[]
  return <div className="page agentsPage"><div className="agentCards">{providers.map(provider=><article key={provider.id} className={`agentCard agent-${provider.state}`}><header><div className="agentLogo"><ProviderMark id={provider.id} size={34}/></div><div><h2>{provider.name}</h2><span><i className={`stateDot state-${provider.state}`}/>{providerState(provider.state,t)}</span></div><b>{provider.capabilities.tools?'deep':'usage'}</b></header><div className="agentNumbers"><div><span>{t('agents.sessions')}</span><strong>{provider.observedSessions}</strong></div><div><span>{t('agents.calls')}</span><strong>{provider.indexedCalls}</strong></div><div><span>{t('agents.tokens')}</span><strong>{formatTokens(provider.tokensToday)}</strong></div></div><div className="capabilities"><span>{t('agents.capabilities')}</span><div><b className={provider.capabilities.usage?'yes':''}>{t('agents.usage')}</b><b className={provider.capabilities.tools?'yes':''}>{t('agents.tools')}</b><b className={provider.capabilities.limits?'yes':''}>{t('agents.limits')}</b></div></div><div className="agentMeta"><div><span>{t('agents.last')}</span><strong>{provider.lastActivity?formatDateTime(provider.lastActivity):'—'}</strong></div><div><span>Local data</span><code>{provider.dataPath||'—'}</code></div>{provider.note&&<p>{provider.note}</p>}{provider.error&&<p className="agentError">{provider.error}</p>}</div></article>)}</div></div>
}

function SettingsView({report,settings,setSettings,save,saving,saved,doctor,setDoctor,t}:{report:ControlPlaneResponse;settings:VarnqorSettings|null;setSettings:SettingsSetter;save:(value?:VarnqorSettings|null)=>Promise<void>;saving:boolean;saved:boolean;doctor:DoctorReport|null;setDoctor:Dispatch<SetStateAction<DoctorReport|null>>;t:Translate}){
  if(!settings)return <Loading/>
  const runDoctor=async()=>{const bridge=window.go?.main?.App;if(bridge?.RunDoctor)setDoctor(await bridge.RunDoctor())}
  const setTheme=(theme:string)=>setSettings(current=>current?({...current,theme}):current)
  const setLanguage=(language:string)=>setSettings(current=>current?({...current,language}):current)
  const setRefresh=(refreshSeconds:number)=>setSettings(current=>current?({...current,refreshSeconds}):current)
  const setBudget=(budget:ConfiguredBudgetSettings)=>setSettings(current=>current?({...current,budget}):current)
  return <div className="page settingsPage"><div className="settingsGrid">
    <Panel title={t('settings.appearance')}><div className="settingRows"><label><span>{t('settings.theme')}</span><select value={settings.theme} onChange={event=>setTheme(event.target.value)}><option value="dark">{t('settings.theme.dark')}</option><option value="light">{t('settings.theme.light')}</option><option value="system">{t('settings.theme.system')}</option></select></label><label><span>{t('settings.language')}</span><select value={settings.language} onChange={event=>setLanguage(event.target.value)}>{languageOptions.map(language=><option key={language} value={language}>{t(`language.${language}`)}</option>)}</select></label></div></Panel>
    <Panel title={t('settings.monitoring')}><div className="settingRows"><label><span>{t('settings.refresh')}</span><div className="numberUnit"><input type="number" min="2" max="60" value={settings.refreshSeconds} onChange={event=>setRefresh(Number(event.target.value))}/><b>sec</b></div></label></div></Panel>
    <Panel title={t('settings.budget')} sub={t('settings.budgetSub')}><div className="settingRows"><label className="toggleRow"><span>{t('settings.enabled')}</span><button className={`switch ${settings.budget.enabled?'on':''}`} onClick={()=>setBudget({...settings.budget,enabled:!settings.budget.enabled})}><i/></button></label><label><span>{t('settings.limit')}</span><input type="number" min="0" step="1000" value={settings.budget.limitTokens} onChange={event=>setBudget({...settings.budget,limitTokens:Number(event.target.value)})}/></label><label><span>{t('settings.reset')}</span><div className="numberUnit"><input type="number" min="0" max="23" value={settings.budget.resetHour} onChange={event=>setBudget({...settings.budget,resetHour:Number(event.target.value)})}/><b>:00</b></div></label></div></Panel>
    <Panel title={t('settings.diagnostics')}><div className="doctorBlock"><button className="secondaryAction" onClick={()=>void runDoctor()}>{t('settings.doctor')}</button>{doctor?<div className={`doctorResult doctor-${doctor.status}`}><strong>{doctor.status.toUpperCase()}</strong><span>{doctor.pass} pass · {doctor.warn} warn · {doctor.fail} fail</span></div>:<span className="mutedText">Not checked in this session</span>}</div></Panel>
    <Panel title={t('settings.about')}><div className="aboutRows"><div><span>{t('settings.version')}</span><strong>{report.status.version}</strong></div><div><span>Storage</span><strong>Local SQLite</strong></div><div><span>Control</span><strong>{settings.control.enabled?'Local decision support':'Disabled'}</strong></div><div><span>Privacy</span><strong>{t('settings.privacy')}</strong></div></div></Panel>
  </div><div className="saveDock"><div><span>V4 settings schema · stored locally</span></div><button className="primaryAction" disabled={saving} onClick={()=>void save()}><Icon name={saved?'check':'settings'}/>{saved?t('settings.saved'):saving?'Saving…':t('settings.save')}</button></div></div>
}

function ThemeButton({settings,setSettings,save}:{settings:VarnqorSettings|null;setSettings:SettingsSetter;save:(value?:VarnqorSettings|null)=>Promise<void>}){
  const current=settings?.theme||'dark'
  const next=current==='dark'?'light':current==='light'?'system':'dark'
  return <button className="toolButton" onClick={()=>{if(!settings)return;const value={...settings,theme:next};setSettings(value);void save(value)}} title={`Theme: ${current}`}><Icon name={current==='dark'?'moon':current==='light'?'sun':'system'}/></button>
}

function CommandPalette({query,setQuery,navigation,onView,onRange,onClose}:{query:string;setQuery:(value:string)=>void;navigation:{id:View;icon:IconName;label:string}[];onView:(view:View)=>void;onRange:(minutes:number)=>void;onClose:()=>void}){
  const normalized=query.trim().toLowerCase()
  const pages=navigation.filter(item=>!normalized||item.label.toLowerCase().includes(normalized))
  return <div className="paletteOverlay" onMouseDown={event=>{if(event.target===event.currentTarget)onClose()}}><div className="commandPalette"><header><Icon name="search"/><input autoFocus value={query} onChange={event=>setQuery(event.target.value)} placeholder="Go to a control surface or set a range…"/><kbd>Esc</kbd></header><div className="paletteBody"><small>GO TO</small>{pages.map(item=><button key={item.id} onClick={()=>onView(item.id)}><Icon name={item.icon}/><span>{item.label}</span><Icon name="arrow"/></button>)}<small>RANGE</small>{ranges.map(range=><button key={range.m} onClick={()=>onRange(range.m)}><Icon name="pulse"/><span>{range.l} window</span><b>{range.m} min</b></button>)}</div><footer><span>Varnqor Control Plane</span><kbd>Ctrl K</kbd></footer></div></div>
}

function Panel({title,sub,action,className='',children}:{title:string;sub?:string;action?:ReactNode;className?:string;children:ReactNode}){return <section className={`panel ${className}`}><header className="panelHeader"><div><h2>{title}</h2>{sub&&<p>{sub}</p>}</div>{action}</header>{children}</section>}
function Metric({label,value,sub}:{label:string;value:string;sub?:string}){return <div className="metric"><span>{label}</span><strong>{value}</strong>{sub&&<small>{sub}</small>}</div>}
function Scenario({label,value,sub}:{label:string;value:string;sub:string}){return <div className="scenario"><span>{label}</span><strong>{value}</strong><small>{sub}</small></div>}
function Empty({icon,text}:{icon:IconName;text:string}){return <div className="empty"><Icon name={icon}/><span>{text}</span></div>}

function DecisionRow({item,index}:{item:ControlDecision;index:number}){return <article className={`decisionRow priority-${item.priority}`}><span className="decisionIndex">{String(index+1).padStart(2,'0')}</span><div><div className="decisionMeta"><b>{item.priority}</b><small>{item.kind}</small></div><strong>{item.title}</strong><p>{item.reason}</p>{(item.provider||item.project)&&<footer>{item.provider&&<ProviderBadge id={item.provider}/>} {item.project&&<span>{cleanProject(item.project)}</span>}</footer>}</div>{item.estimatedSavingsHigh>0&&<aside><span>impact</span><strong>{formatTokens(item.estimatedSavingsLow)}–{formatTokens(item.estimatedSavingsHigh)}</strong></aside>}</article>}
function IncidentRow({item}:{item:ControlIncident}){return <article className={`incidentRow severity-${item.severity}`}><i/><div><div><b>{item.severity}</b><span>{item.source}</span></div><strong>{item.title}</strong><p>{item.detail}</p></div>{item.value>0&&<aside><strong>{incidentValue(item)}</strong><span>/ {incidentThreshold(item)}</span></aside>}</article>}
function DimensionRows({items}:{items:DimensionActivity[]}){return <div className="dimensionRows">{items.slice(0,8).map((item,index)=><div key={`${item.provider}-${item.name}-${index}`}><span>{item.provider&&<ProviderMark id={item.provider} size={14}/>}<strong>{cleanProject(item.name)}</strong></span><b>{formatTokens(item.tokens.total)}</b><em>{item.sharePercent.toFixed(1)}%</em><i><b style={{width:`${Math.min(100,item.sharePercent)}%`}}/></i></div>)}</div>}
function RecentTable({items}:{items:CallActivity[]}){return <div className="recentTable"><div className="recentHead"><span>Time</span><span>Agent</span><span>Project</span><span>Model</span><span>Tokens</span></div>{items.slice(0,40).map((item,index)=><div className="recentRow" key={`${item.timestamp}-${index}`}><span>{formatClock(item.timestamp)}</span><span><ProviderBadge id={item.provider}/></span><span title={item.project}>{cleanProject(item.project)}</span><span title={item.model}>{item.model||'—'}</span><strong>{formatTokens(item.tokens.total)}</strong></div>)}</div>}
function Compare({label,value,delta,comparable=true,points=false,sub}:{label:string;value:string;delta?:number;comparable?:boolean;points?:boolean;sub?:string}){const change=delta??0;return <div className="compare"><span>{label}</span><strong>{value}</strong>{sub?<small>{sub}</small>:comparable?<small className={change>0?'up':change<0?'down':''}>{change>0?'+':''}{change.toFixed(1)}{points?' pts':'%'}</small>:<small>new window</small>}</div>}
function Select({value,onChange,options,label}:{value:string;onChange:(value:string)=>void;options:string[];label:string}){return <label className="selectWrap"><span>{label}</span><select value={value} onChange={event=>onChange(event.target.value)}>{options.map((option,index)=><option key={`${option}-${index}`} value={option}>{option||'All'}</option>)}</select></label>}
function PolicyNumber({label,value,min,max,step,suffix,onChange}:{label:string;value:number;min:number;max:number;step:number;suffix:string;onChange:(value:number)=>void}){return <label className="policyNumber"><div><strong>{label}</strong><span>{min}–{max}{suffix}</span></div><div><input type="number" value={value} min={min} max={max} step={step} onChange={event=>onChange(Number(event.target.value))}/><b>{suffix}</b></div></label>}

function Timeline({data,baseline=0,rangeMinutes}:{data:TimelineBucket[];baseline?:number;rangeMinutes:number}){
  const max=Math.max(1,...data.map(bucket=>bucket.tokens.total))
  const baselineBucket=baseline>0&&data.length>0?baseline*(rangeMinutes/data.length)/60:0
  const baselinePct=Math.min(100,baselineBucket/max*100)
  return <div className="timeline"><div className="timelineGrid"><i/><i/><i/><i/></div>{baseline>0&&<div className="baselineLine" style={{bottom:`calc(30px + ${baselinePct*.72}%)`}}><span>{formatTokens(baseline)}/h baseline</span></div>}<div className="timelineBars">{data.map((bucket,index)=>{const height=Math.max(2,bucket.tokens.total/max*100);const showLabel=index===0||index===data.length-1||index%Math.max(1,Math.floor(data.length/5))===0;return <div className="timelineSlot" key={`${bucket.start}-${index}`} title={`${formatDateTime(bucket.start)} · ${formatTokens(bucket.tokens.total)} tokens`}><i style={{height:`${height}%`}}/><span>{showLabel?timelineLabel(bucket.start,rangeMinutes):''}</span></div>})}</div></div>
}

function Loading(){return <div className="loading"><div className="loadingMark"><i/><i/><i/></div><span>Building local control picture…</span></div>}
function ControlGlyph({state}:{state:string}){return <span className={`controlGlyph glyph-${state}`}><i/><b/><em/></span>}

function Icon({name,className=''}:{name:IconName;className?:string}){
  const paths:Record<IconName,ReactNode>={
    control:<><path d="M4 5h16v14H4z"/><path d="M8 9h3v6H8zM14 7h2v8h-2z"/></>,
    missions:<><circle cx="12" cy="12" r="8"/><path d="M12 7v5l3 2"/></>,
    investigate:<><circle cx="10.5" cy="10.5" r="6"/><path d="m15 15 5 5"/></>,
    optimize:<><path d="M4 17h3l3-5 3 2 3-7 4 2"/><path d="M17 5h3v3"/></>,
    policies:<><path d="M12 3 5 6v5c0 4.4 2.8 7.8 7 10 4.2-2.2 7-5.6 7-10V6z"/><path d="m9 12 2 2 4-5"/></>,
    agents:<><circle cx="8" cy="9" r="3"/><circle cx="16.5" cy="10" r="2.5"/><path d="M3 19c.8-3 2.5-4.5 5-4.5S12.3 16 13 19M14 15c2.8-.6 5 .8 6 3"/></>,
    settings:<><circle cx="12" cy="12" r="3"/><path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6 7 7M17 17l1.4 1.4M18.4 5.6 17 7M7 17l-1.4 1.4"/></>,
    search:<><circle cx="10.5" cy="10.5" r="6"/><path d="m15 15 5 5"/></>,
    sun:<><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M19.1 4.9l-1.4 1.4M6.3 17.7l-1.4 1.4"/></>,
    moon:<path d="M19 15.5A8 8 0 0 1 8.5 5 8 8 0 1 0 19 15.5Z"/>,
    system:<><rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8M12 17v4"/></>,
    copy:<><rect x="8" y="8" width="11" height="11" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/></>,
    refresh:<><path d="M20 6v5h-5"/><path d="M18.5 15a7 7 0 1 1-.7-7.7L20 11"/></>,
    pulse:<path d="M3 12h4l2-6 4 12 2-6h6"/>,shield:<><path d="M12 3 5 6v5c0 4.4 2.8 7.8 7 10 4.2-2.2 7-5.6 7-10V6z"/><path d="m9 12 2 2 4-5"/></>,
    close:<path d="m7 7 10 10M17 7 7 17"/>,command:<><rect x="4" y="4" width="6" height="6" rx="2"/><rect x="14" y="4" width="6" height="6" rx="2"/><rect x="4" y="14" width="6" height="6" rx="2"/><rect x="14" y="14" width="6" height="6" rx="2"/></>,
    arrow:<path d="M5 12h14M14 7l5 5-5 5"/>,check:<path d="m5 12 4 4L19 6"/>,warning:<><path d="M12 3 2.8 20h18.4z"/><path d="M12 9v5M12 17h.01"/></>,
  }
  return <svg className={className} viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{paths[name]}</svg>
}

function controlProfile(name:'balanced'|'economy'|'throughput'):ControlSettings{
  if(name==='economy')return{enabled:true,profile:name,burnCeilingRatio:1.45,sessionDominancePercent:55,projectDominancePercent:65,cacheFloorPercent:20,budgetWarningPercent:72,missionActiveMinutes:10,missionCoolingMinutes:40,decisionHorizonMinutes:120,minimumEvaluationTokens:800}
  if(name==='throughput')return{enabled:true,profile:name,burnCeilingRatio:2.25,sessionDominancePercent:78,projectDominancePercent:85,cacheFloorPercent:8,budgetWarningPercent:92,missionActiveMinutes:15,missionCoolingMinutes:60,decisionHorizonMinutes:180,minimumEvaluationTokens:1500}
  return{enabled:true,profile:'balanced',burnCeilingRatio:1.7,sessionDominancePercent:62,projectDominancePercent:72,cacheFloorPercent:15,budgetWarningPercent:80,missionActiveMinutes:12,missionCoolingMinutes:45,decisionHorizonMinutes:120,minimumEvaluationTokens:1000}
}
function formatTokens(value:number){const amount=Math.abs(value);if(amount>=1e9)return`${(value/1e9).toFixed(amount>=1e10?1:2)}B`;if(amount>=1e6)return`${(value/1e6).toFixed(amount>=1e7?1:2)}M`;if(amount>=1e3)return`${(value/1e3).toFixed(amount>=1e4?1:2)}K`;return Math.round(value).toLocaleString()}
function shortID(value:string){if(!value)return'—';return value.length<=14?value:`${value.slice(0,7)}…${value.slice(-4)}`}
function cleanProject(value:string){if(!value)return'Unknown';const parts=value.replace(/\\/g,'/').split('/').filter(Boolean);return parts[parts.length-1]||value}
function relativeAge(minutes:number){if(minutes<1)return'now';if(minutes<60)return`${minutes}m`;if(minutes<1440)return`${Math.floor(minutes/60)}h`;return`${Math.floor(minutes/1440)}d`}
function formatDateTime(value:string){if(!value)return'—';const date=new Date(value);return Number.isNaN(date.getTime())?'—':date.toLocaleString([],{month:'short',day:'2-digit',hour:'2-digit',minute:'2-digit'})}
function formatClock(value:string){const date=new Date(value);return Number.isNaN(date.getTime())?'—':date.toLocaleTimeString([],{hour:'2-digit',minute:'2-digit',second:'2-digit'})}
function timelineLabel(value:string,range:number){const date=new Date(value);if(Number.isNaN(date.getTime()))return'';return range>=1440?date.toLocaleDateString([],{month:'short',day:'numeric'}):date.toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'})}
function stateLabel(state:string,t:Translate){const key=`common.${state}`,translated=t(key);return translated===key?state.charAt(0).toUpperCase()+state.slice(1):translated}
function providerState(state:string,t:Translate){const key=`agents.${state}`,translated=t(key);return translated===key?state:translated}
function incidentValue(item:ControlIncident){if(item.kind.includes('dominance')||item.kind==='cache-floor'||item.kind==='budget-risk')return`${item.value.toFixed(1)}%`;if(item.kind==='burn-ceiling')return`${item.value.toFixed(2)}×`;return String(item.value)}
function incidentThreshold(item:ControlIncident){if(item.kind.includes('dominance')||item.kind==='cache-floor'||item.kind==='budget-risk')return`${item.threshold.toFixed(0)}%`;if(item.kind==='burn-ceiling')return`${item.threshold.toFixed(2)}×`;return String(item.threshold)}
