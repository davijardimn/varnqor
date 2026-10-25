import { useState } from 'react'

export default function SettingsPanel({value,onClose,onSaved}:{value:VarnqorSettings;onClose:()=>void;onSaved:(value:VarnqorSettings)=>void}){
 const[draft,setDraft]=useState<VarnqorSettings>(()=>structuredClone(value));const[saving,setSaving]=useState(false);const[error,setError]=useState('')
 const save=async()=>{const fn=window.go?.main?.App?.SaveSettings;if(!fn){setError('Settings bridge is unavailable.');return}setSaving(true);setError('');try{const saved=await fn(draft);onSaved(saved);onClose()}catch(e){setError(String(e))}finally{setSaving(false)}}
 return <div className="modalBackdrop" onMouseDown={onClose}><section className="modalPanel" onMouseDown={e=>e.stopPropagation()} aria-label="Varnqor settings">
  <div className="modalHeader"><div><span className="label">VARNQOR SETTINGS</span><h2>Local monitoring controls.</h2></div><button className="iconButton" onClick={onClose}>×</button></div>
  <div className="settingsGroup"><label><span>Live refresh</span><small>How often the desktop refreshes the incremental local index.</small></label><div className="fieldWithUnit"><input type="number" min={2} max={60} value={draft.refreshSeconds} onChange={e=>setDraft({...draft,refreshSeconds:Number(e.target.value)})}/><span>seconds</span></div></div>
  <div className="settingsDivider"/>
  <div className="settingsToggle"><div><strong>Configured daily budget</strong><small>Enables remaining + runway. This is explicitly your budget, not a provider-reported quota.</small></div><button className={`toggle ${draft.budget.enabled?'on':''}`} onClick={()=>setDraft({...draft,budget:{...draft.budget,enabled:!draft.budget.enabled}})}><i/></button></div>
  <div className={`budgetFields ${draft.budget.enabled?'':'disabled'}`}>
   <label><span>Label</span><input disabled={!draft.budget.enabled} value={draft.budget.label} onChange={e=>setDraft({...draft,budget:{...draft.budget,label:e.target.value}})}/></label>
   <label><span>Token budget / day</span><input disabled={!draft.budget.enabled} type="number" min={1} step={1000} value={draft.budget.limitTokens||''} placeholder="10000000" onChange={e=>setDraft({...draft,budget:{...draft.budget,limitTokens:Number(e.target.value)}})}/></label>
   <label><span>Daily reset hour</span><input disabled={!draft.budget.enabled} type="number" min={0} max={23} value={draft.budget.resetHour} onChange={e=>setDraft({...draft,budget:{...draft.budget,resetHour:Number(e.target.value)}})}/><small>0–23, using your local timezone.</small></label>
  </div>
  {error&&<div className="modalError">{error}</div>}
  <div className="modalActions"><button className="ghostButton" onClick={onClose}>CANCEL</button><button className="primaryButton" onClick={save} disabled={saving}>{saving?'SAVING…':'SAVE LOCALLY'}</button></div>
 </section></div>
}
