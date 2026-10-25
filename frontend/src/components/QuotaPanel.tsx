const compact=new Intl.NumberFormat(undefined,{notation:'compact',maximumFractionDigits:2})
function fmt(value:number){return compact.format(value)}
function time(value?:string){if(!value)return'—';const date=new Date(value);return Number.isNaN(date.getTime())?'—':date.toLocaleString([], {weekday:'short',hour:'2-digit',minute:'2-digit'})}
function runway(hours:number){if(hours<1)return `${Math.max(0,Math.round(hours*60))}m`;if(hours<24)return `${hours.toFixed(1)}h`;return `${(hours/24).toFixed(1)}d`}

export default function QuotaPanel({quota,onConfigure}:{quota:QuotaStatus;onConfigure:()=>void}){
 if(!quota.available)return <section className="quotaPanel quotaEmpty"><div><span className="label">M8 · TRUSTED LIMITS</span><h2>Runway needs a trusted limit.</h2><p>Varnqor will not infer quota from your historical baseline. Configure a local daily budget to enable transparent remaining/runway calculations.</p></div><button className="ghostButton" onClick={onConfigure}>CONFIGURE BUDGET</button></section>
 const width=Math.min(100,Math.max(0,quota.utilizationPercent))
 return <section className={`quotaPanel ${quota.willExceedBeforeReset?'quotaPressure':''}`} aria-label="Configured budget and runway">
  <div className="quotaHeader"><div><span className="label">M8 · TRUSTED LIMITS · {quota.source.toUpperCase()}</span><h2>{quota.label}</h2><p>This is a user-configured budget, not an official Claude quota. Runway is calculated from remaining budget and the current rolling pace.</p></div><button className="ghostButton" onClick={onConfigure}>SETTINGS</button></div>
  <div className="quotaMeter"><div style={{width:`${width}%`}}/></div>
  <div className="quotaGrid">
   <div><span className="label">USED</span><strong>{fmt(quota.used)}</strong><small>{quota.utilizationPercent.toFixed(1)}% of {fmt(quota.limit)}</small></div>
   <div><span className="label">REMAINING</span><strong>{fmt(quota.remaining)}</strong><small>configured tokens</small></div>
   <div><span className="label">RUNWAY</span><strong>{quota.runwayAvailable?runway(quota.runwayHours):'PACE IDLE'}</strong><small>at current pace</small></div>
   <div><span className="label">RESET</span><strong>{time(quota.resetAt)}</strong><small>local time</small></div>
   <div><span className="label">AT RESET</span><strong>{fmt(quota.projectedAtReset)}</strong><small>{quota.willExceedBeforeReset?'PROJECTED OVER BUDGET':'within configured budget'}</small></div>
  </div>
 </section>
}
