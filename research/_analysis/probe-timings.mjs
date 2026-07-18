import { readFileSync } from 'node:fs'
const lines = readFileSync('C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/agent-a9d48684d761cd037.jsonl', 'utf8').split('\n').filter(Boolean)
const evs = []
for (const l of lines) {
  let j; try { j = JSON.parse(l) } catch { continue }
  const ts = j.timestamp && Date.parse(j.timestamp)
  if (!ts) continue
  const m = j.message
  if (m && m.role === 'assistant' && Array.isArray(m.content)) {
    for (const b of m.content) if (b.type === 'tool_use') evs.push({ t: ts, kind: 'use', name: b.name, input: JSON.stringify(b.input || {}).slice(0, 80) })
  }
  if (m && m.role === 'user' && Array.isArray(m.content)) {
    for (const b of m.content) if (b.type === 'tool_result') evs.push({ t: ts, kind: 'result' })
  }
}
evs.sort((a, b) => a.t - b.t)
let lastUse = null
for (const e of evs) {
  if (e.kind === 'use') lastUse = e
  else if (lastUse) { console.log(((e.t - lastUse.t) / 1000).toFixed(1).padStart(6) + 's ', lastUse.name.padEnd(9), lastUse.input); lastUse = null }
}
