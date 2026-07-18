// Was qmd auto-indexing responsible for run 4's Bash floor? Join every run-4 Bash call
// against sc-recall-index firing windows from the hook log; compare durations near vs far.
import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'

// 1. Indexing events from the hook log (AgentOrange was the launching project).
const hookLog = readFileSync('C:/Users/gbloc/Projects/AgentOrange/.claude/prosthetic-conscience-hook.log', 'utf8')
const idxEvents = []
for (const line of hookLog.split('\n')) {
  if (!line.includes('sc-recall-index') || !line.includes('running qmd update')) continue
  const ts = Date.parse(line.slice(0, 20))
  if (ts) idxEvents.push(ts)
}

// 2. Bash calls from run-4 transcripts: duration + start time.
const dir = 'C:/Users/gbloc/AppData/Local/Temp/claude/feov-time-forensics/run4'
const calls = []
for (const f of readdirSync(dir).filter((f) => f.endsWith('.jsonl'))) {
  const evs = []
  for (const l of readFileSync(join(dir, f), 'utf8').split('\n')) {
    let j; try { j = JSON.parse(l) } catch { continue }
    const ts = j.timestamp && Date.parse(j.timestamp)
    if (!ts || !j.message) continue
    if (j.message.role === 'assistant' && Array.isArray(j.message.content)) {
      for (const b of j.message.content) if (b.type === 'tool_use') evs.push({ t: ts, kind: 'use', name: b.name })
    }
    if (j.message.role === 'user' && Array.isArray(j.message.content)) {
      for (const b of j.message.content) if (b.type === 'tool_result') evs.push({ t: ts, kind: 'result' })
    }
  }
  evs.sort((a, b) => a.t - b.t)
  let last = null
  for (const e of evs) {
    if (e.kind === 'use') last = e
    else if (last) { if (last.name === 'Bash') calls.push({ t: last.t, dur: (e.t - last.t) / 1000 }); last = null }
  }
}

// 3. Window join: run-4 window only; near = an indexing event within [t-15s, t+dur+15s].
const t0 = Math.min(...calls.map((c) => c.t)), t1 = Math.max(...calls.map((c) => c.t))
const runIdx = idxEvents.filter((t) => t >= t0 - 60000 && t <= t1 + 60000)
const near = [], far = []
for (const c of calls) {
  const hit = runIdx.some((t) => t >= c.t - 15000 && t <= c.t + c.dur * 1000 + 15000)
  ;(hit ? near : far).push(c.dur)
}
const stats = (a) => {
  if (!a.length) return 'n=0'
  a = [...a].sort((x, y) => x - y)
  const q = (p) => a[Math.min(a.length - 1, Math.floor(p * a.length))]
  return `n=${a.length} med=${q(0.5).toFixed(1)}s p90=${q(0.9).toFixed(1)}s mean=${(a.reduce((s, x) => s + x, 0) / a.length).toFixed(1)}s`
}
console.log('indexing events in hook log (total):', idxEvents.length, '| inside run-4 window:', runIdx.length)
console.log('Bash calls near an indexing window: ', stats(near))
console.log('Bash calls far from any indexing:   ', stats(far))

// 4. Indexing burst structure: gaps between consecutive updates inside the run window.
const gaps = []
for (let i = 1; i < runIdx.length; i++) gaps.push((runIdx[i] - runIdx[i - 1]) / 1000)
const under5 = gaps.filter((g) => g < 5).length
console.log(`update cadence: ${runIdx.length} updates; ${under5} arrived <5s after the previous (overlap candidates)`)
