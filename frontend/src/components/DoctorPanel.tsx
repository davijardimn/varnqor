import { useEffect,useState } from 'react'

export default function DoctorPanel({onClose}:{onClose:()=>void}){
 const[report,setReport]=useState<DoctorReport|null>(null);const[error,setError]=useState('')
 useEffect(()=>{let active=true;const run=async()=>{const fn=window.go?.main?.App?.RunDoctor;if(!fn){setError('Doctor bridge is unavailable.');return}try{const value=await fn();if(active)setReport(value)}catch(e){if(active)setError(String(e))}};void run();return()=>{active=false}},[])
 return <div className="modalBackdrop" onMouseDown={onClose}><section className="modalPanel doctorModal" onMouseDown={e=>e.stopPropagation()} aria-label="Varnqor doctor">
  <div className="modalHeader"><div><span className="label">VARNQOR DOCTOR</span><h2>{report?report.status.toUpperCase():'RUNNING CHECKS…'}</h2>{report&&<p>{report.pass} pass · {report.warn} warn · {report.fail} fail · {report.version}</p>}</div><button className="iconButton" onClick={onClose}>×</button></div>
  {error&&<div className="modalError">{error}</div>}
  <div className="doctorRows">{(report?.checks??[]).map(item=><div className="doctorRow" key={item.id}><span className={`doctorState state-${item.status}`}>{item.status.toUpperCase()}</span><div><strong>{item.label}</strong><small>{item.detail}</small></div></div>)}</div>
  <div className="doctorCli">CLI: <code>varnqor doctor</code> · JSON: <code>varnqor doctor --json</code></div>
 </section></div>
}
