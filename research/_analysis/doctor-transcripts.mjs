import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
const root = join(process.env.USERPROFILE, '.claude', 'projects')
const files = []
for (const proj of readdirSync(root)) {
  const dir = join(root, proj)
  let entries; try { entries = readdirSync(dir) } catch { continue }
  for (const f of entries) if (f.endsWith('.jsonl')) {
    try { files.push({ p: join(dir, f), m: statSync(join(dir, f)).mtimeMs }) } catch {}
  }
}
files.sort((a, b) => b.m - a.m)
const scan = files.slice(0, 50)
const mcp = {}, skills = {}, cmds = {}, hooks = {}, denials = {}
let oldest = Infinity, newest = 0
for (const { p, m } of scan) {
  oldest = Math.min(oldest, m); newest = Math.max(newest, m)
  let txt; try { txt = readFileSync(p, 'utf8') } catch { continue }
  for (const line of txt.split('\n')) {
    if (!line) continue
    // cheap prefilters before JSON.parse
    if (line.includes('"tool_use"')) {
      let j; try { j = JSON.parse(line) } catch { continue }
      const c = j.message && j.message.content
      if (Array.isArray(c)) for (const b of c) {
        if (b.type !== 'tool_use') continue
        if (b.name && b.name.startsWith('mcp__')) {
          const server = b.name.split('__')[1]
          mcp[server] = (mcp[server] || 0) + 1
        }
        if (b.name === 'Skill' && b.input && b.input.skill) skills[b.input.skill] = (skills[b.input.skill] || 0) + 1
      }
    } else if (line.includes('hook_success') || line.includes('hook_non_blocking_error') || line.includes('hook_cancelled') || line.includes('hook_error')) {
      let j; try { j = JSON.parse(line) } catch { continue }
      const a = j.attachment
      if (a && a.hookName !== undefined) {
        const key = (a.hookName || '?') + '|' + (a.hookEvent || '?')
        hooks[key] = hooks[key] || { n: 0, durs: [], timeouts: 0 }
        hooks[key].n++
        if (typeof a.durationMs === 'number') hooks[key].durs.push(a.durationMs)
        if (a.timedOut) hooks[key].timeouts++
      }
    } else if (line.includes('command-name')) {
      const m2 = line.match(/<command-name>\/?([\w:-]+)<\/command-name>/)
      if (m2) cmds[m2[1]] = (cmds[m2[1]] || 0) + 1
    } else if (line.includes('toolDenialKind')) {
      let j; try { j = JSON.parse(line) } catch { continue }
      if (j.toolDenialKind) {
        denials[j.toolDenialKind] = denials[j.toolDenialKind] || []
        denials[j.toolDenialKind].push(p.split(/[\\/]/).pop().slice(0, 8))
      }
    }
  }
}
const med = (a) => { if (!a.length) return null; a = [...a].sort((x, y) => x - y); return a[Math.floor(a.length / 2)] }
const hookOut = {}
for (const [k, v] of Object.entries(hooks)) hookOut[k] = { runs: v.n, medMs: med(v.durs), maxMs: v.durs.length ? Math.max(...v.durs) : null, timeouts: v.timeouts }
console.log(JSON.stringify({
  window: { sessions: scan.length, days: ((newest - oldest) / 86400000).toFixed(1) },
  mcpCalls: mcp, skillCalls: skills, slashCmds: cmds, hooks: hookOut,
  denialKinds: Object.fromEntries(Object.entries(denials).map(([k, v]) => [k, v.length])),
}, null, 1))
