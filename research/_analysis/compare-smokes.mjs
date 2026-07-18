import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'

function analyze(dir) {
  const out = { rounds: 0, bash: [], mdWrite: [], tools: {}, batched: 0, msgs: 0, agentMin: 0 }
  for (const f of readdirSync(dir).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))) {
    const evs = []
    let first = null, last = null
    for (const l of readFileSync(join(dir, f), 'utf8').split('\n')) {
      let j; try { j = JSON.parse(l) } catch { continue }
      const ts = j.timestamp && Date.parse(j.timestamp)
      if (!ts || !j.message) continue
      first = first ?? ts; last = ts
      if (j.message.role === 'assistant' && Array.isArray(j.message.content)) {
        const uses = j.message.content.filter((b) => b.type === 'tool_use')
        if (j.message.usage) out.rounds++
        if (uses.length > 1) out.batched++
        if (uses.length >= 1) out.msgs++
        for (const b of uses) evs.push({ t: ts, kind: 'use', name: b.name, input: b.input || {} })
      }
      if (j.message.role === 'user' && Array.isArray(j.message.content)) {
        for (const b of j.message.content) if (b.type === 'tool_result') evs.push({ t: ts, kind: 'result' })
      }
    }
    if (first && last) out.agentMin += (last - first) / 60000
    evs.sort((a, b) => a.t - b.t)
    let lastUse = null
    for (const e of evs) {
      if (e.kind === 'use') lastUse = e
      else if (lastUse) {
        const dur = (e.t - lastUse.t) / 1000
        out.tools[lastUse.name] = (out.tools[lastUse.name] || 0) + 1
        if (lastUse.name === 'Bash') out.bash.push(dur)
        if ((lastUse.name === 'Write' || lastUse.name === 'Edit') && String(lastUse.input.file_path || '').endsWith('.md')) out.mdWrite.push(dur)
        lastUse = null
      }
    }
  }
  return out
}
const med = (a) => { if (!a.length) return NaN; a = [...a].sort((x, y) => x - y); return a[Math.floor(a.length / 2)] }
const base = analyze('C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_5a9b945d-200')
const ab = analyze('C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_bfb5cee2-a15')
const row = (k, b, a) => console.log(k.padEnd(30), String(b).padStart(10), String(a).padStart(10))
row('metric', 'baseline', 'A/B')
row('API rounds (usage msgs)', base.rounds, ab.rounds)
row('agent-minutes (sum)', base.agentMin.toFixed(1), ab.agentMin.toFixed(1))
row('tool msgs w/ >1 call', `${base.batched}/${base.msgs}`, `${ab.batched}/${ab.msgs}`)
row('Bash calls', base.bash.length, ab.bash.length)
row('Bash median s', med(base.bash).toFixed(1), med(ab.bash).toFixed(1))
row('md Write/Edit calls', base.mdWrite.length, ab.mdWrite.length)
row('md write median s', med(base.mdWrite).toFixed(1), med(ab.mdWrite).toFixed(1))
console.log('tools baseline:', JSON.stringify(base.tools))
console.log('tools A/B     :', JSON.stringify(ab.tools))
