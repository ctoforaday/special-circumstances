// Judicial analytics over a run's journal: ruling-type distribution, dispute traffic,
// gap longevity (rounds alive), and grade migrations (the downgrade process).
import { readFileSync } from 'node:fs'

const MASS = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }

function analyze(journalPath, label) {
  const results = readFileSync(journalPath, 'utf8').split('\n').filter(Boolean)
    .map((l) => { try { return JSON.parse(l) } catch { return null } })
    .filter((j) => j && j.result && typeof j.result === 'object')
    .map((j) => j.result)

  const rulings = {}
  let judgeSittings = 0
  const disputes = { raised: 0, accepted: 0, rejected: 0 }
  const gapLife = new Map() // id -> { first, last, grades: [{round, severity}] }
  let redRound = 0

  for (const r of results) {
    if (Array.isArray(r.resolutions)) {
      judgeSittings++
      for (const x of r.resolutions) rulings[x.resolution] = (rulings[x.resolution] || 0) + 1
    }
    if (Array.isArray(r.grade_disputes)) disputes.raised += r.grade_disputes.length
    if (Array.isArray(r.dispute_responses)) for (const d of r.dispute_responses) disputes[d.response] = (disputes[d.response] || 0) + 1
    if (r.verdict && Array.isArray(r.gaps)) {
      redRound++
      for (const g of r.gaps) {
        const e = gapLife.get(g.id) || { first: redRound, grades: [] }
        e.last = redRound
        e.grades.push({ round: redRound, severity: g.severity, likelihood: g.likelihood, impact: g.impact })
        gapLife.set(g.id, e)
      }
    }
  }

  // Longevity histogram + grade migrations across a gap's life.
  const lifespan = {}
  let down = 0, up = 0, flat = 0
  for (const e of gapLife.values()) {
    const span = e.last - e.first + 1
    lifespan[span] = (lifespan[span] || 0) + 1
    if (e.grades.length > 1) {
      const a = e.grades[0], b = e.grades[e.grades.length - 1]
      const m = (g) => (MASS[g.likelihood] ?? 0) * (MASS[g.impact] ?? 0)
      const d = m(b) - m(a)
      if (d < 0) down++; else if (d > 0) up++; else flat++
    }
  }
  console.log(`== ${label} ==`)
  console.log('judge sittings:', judgeSittings, '| rulings by type:', JSON.stringify(rulings))
  console.log('grade disputes: raised', disputes.raised, '| accepted', disputes.accepted || 0, '| rejected', disputes.rejected || 0)
  console.log('gap lifespan histogram (rounds alive -> count):', JSON.stringify(lifespan))
  console.log('grade migrations over multi-round gaps: down', down, '| up', up, '| flat', flat)
  console.log('total distinct gaps:', gapLife.size, '| red rounds seen:', redRound)
}

analyze('C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_c9dc1f42-7bd/journal.jsonl', 'RUN 5 (through pause)')
analyze('C:/Users/gbloc/Projects/special-circumstances/research/2026-07-14_efficiency-investigation/trajectories/journal.jsonl', 'RUN 4 (complete)')
