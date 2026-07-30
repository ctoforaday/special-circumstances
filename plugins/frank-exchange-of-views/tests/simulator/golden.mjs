// golden.mjs — golden-file assertions for the mjs suites.
//
// Deliberately mirrors the Go helper in tools/internal/difftest so the two
// suites share ONE convention: goldens live in testdata/<name>.golden beside the
// test, an update run rewrites them, a normal run compares, and a golden nobody
// owns is a failure rather than litter.
//
// WHY A CONVENTION AND NOT A LIBRARY: the risk with golden suites is not diff
// quality, it is REVIEW EROSION at scale — a wave regenerates everything, the
// diff is 400 lines, it gets rubber-stamped, and real drift ships inside the
// noise. No library fixes that; what fixes it is (a) one command that updates
// everything so nobody hand-picks, (b) a change REPORT at update time so the
// author sees what moved before committing, (c) staleness failing in CI, and
// (d) orphan detection so renamed tests do not leave stale goldens that look
// authoritative. All four are cheap and live here.
import { readFileSync, writeFileSync, mkdirSync, existsSync, readdirSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import assert from 'node:assert/strict'

const UPDATING = process.env.UPDATE_GOLDENS === '1'

// Every golden touched this process, so orphan detection can subtract.
const claimed = new Set()
const changed = []

export function goldenDir(importMetaUrl) {
  return join(dirname(fileURLToPath(importMetaUrl)), 'testdata')
}

// normalize keeps goldens portable: line endings, absolute run dirs, temp paths,
// and anything else that differs by machine rather than by behaviour.
export function normalize(text, { runDir = null, extra = [] } = {}) {
  let s = String(text).replace(/\r\n/g, '\n')
  if (runDir) {
    s = s.split(runDir).join('{RUN}')
    s = s.split(runDir.replace(/\\/g, '/')).join('{RUN}')
  }
  for (const [pattern, replacement] of extra) s = s.replace(pattern, replacement)
  return s
}

// assertGolden compares (or rewrites) one golden.
//
// The failure message names the update command rather than describing the diff:
// a golden mismatch is either a bug the author must read, or an intended change
// the author must re-record — and in both cases the next action is the same.
export function assertGolden(importMetaUrl, name, actual) {
  const dir = goldenDir(importMetaUrl)
  const path = join(dir, `${name}.golden`)
  claimed.add(path)
  const got = normalize(actual)

  if (UPDATING) {
    mkdirSync(dir, { recursive: true })
    const before = existsSync(path) ? normalize(readFileSync(path, 'utf8')) : null
    if (before !== got) {
      writeFileSync(path, got)
      changed.push({ name, added: before === null, lines: lineDelta(before ?? '', got) })
    }
    return
  }

  if (!existsSync(path)) {
    assert.fail(`missing golden ${name} — record it with: (cd scripts && go run ./golden -update)`)
  }
  const want = normalize(readFileSync(path, 'utf8'))
  if (got !== want) {
    assert.fail(`${name} differs from its golden.\n${firstDifference(want, got)}\n` +
      `If this change is INTENTIONAL: (cd scripts && go run ./golden -update), then commit the testdata change ON ITS OWN so the review surface stays readable.`)
  }
}

// reportGoldens is called from an at-exit test so an update run says what moved.
export function goldenReport() {
  if (!UPDATING) return null
  return changed
}

// orphans finds goldens in the directory that no test claimed this run. A
// renamed scenario otherwise leaves its old golden behind, where it reads as an
// authoritative record of behaviour nothing actually asserts.
export function orphanGoldens(importMetaUrl) {
  const dir = goldenDir(importMetaUrl)
  if (!existsSync(dir)) return []
  return readdirSync(dir)
    .filter((f) => f.endsWith('.golden'))
    .map((f) => join(dir, f))
    .filter((p) => !claimed.has(p))
}

function lineDelta(before, after) {
  const b = before.split('\n').length
  const a = after.split('\n').length
  return a - b
}

function firstDifference(want, got) {
  const w = want.split('\n')
  const g = got.split('\n')
  for (let i = 0; i < Math.max(w.length, g.length); i++) {
    if (w[i] !== g[i]) {
      const start = Math.max(0, i - 3)
      const context = w.slice(start, i).map((l) => `  ${l}`).join('\n')
      return `first difference at line ${i + 1}:\n${context}\n- want: ${w[i] ?? '(end of file)'}\n+ got:  ${g[i] ?? '(end of file)'}`
    }
  }
  return '(differ only in trailing whitespace)'
}
