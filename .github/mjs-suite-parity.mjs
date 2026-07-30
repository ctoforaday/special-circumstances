// The `.mjs` suites are ENUMERATED in .github/workflows/hooks.yml rather than globbed,
// because `node --test <dir>` does not work on Windows. Nothing checked that the
// enumeration matched the tree, and `node --test` exits 0 on a path that does not exist
// — it does not warn, and it does not count the file it could not find. So the list
// could name a deleted suite (CI stays green, testing nothing) or omit a live one (the
// suite never runs once), and both read exactly like a healthy build.
//
// Measured on 2026-07-30, before this guard existed: of four enumerated suites, TWO did
// not exist (read-batching.test.mjs, run-dashboard.test.mjs). This is the third instance
// of the class — record.test.mjs went unlisted for its whole life and never ran in CI —
// and the previous two were fixed as instances, which is why it recurred. This closes it
// with a mechanism instead.
import { readFileSync, globSync } from 'node:fs'

const fail = (msg) => { console.error(msg); process.exitCode = 1 }

const WORKFLOW = '.github/workflows/hooks.yml'

// Every suite in the tree. The working corpus (ideas/, research/, projects/) is not
// swept: it is seeded at runtime and is not CI's business.
const onDisk = [
  ...globSync('plugins/*/tests/**/*.test.mjs'),
  ...globSync('scripts/**/*.test.mjs'),
].filter((p) => !p.includes('node_modules')).map((p) => p.replaceAll('\\', '/'))

// Every suite the workflow hands to `node --test`, wherever in the file it appears.
const workflow = readFileSync(WORKFLOW, 'utf8')
const listed = [...workflow.matchAll(/^\s*(plugins\/[^\s]+\.test\.mjs)\s*$/gm)].map((m) => m[1])

if (!workflow.includes('node --test')) {
  fail(`${WORKFLOW}: no \`node --test\` step found at all — this guard greps for one, so it has been passing without reading anything. If the suites moved to another runner, move this check with them.`)
} else if (listed.length === 0) {
  fail(`${WORKFLOW}: found \`node --test\` but could not extract a single suite path — the pattern this guard greps for has moved, so it has been passing without reading anything.`)
}

const want = [...onDisk].sort()
const have = [...listed].sort()

for (const p of have.filter((p) => !want.includes(p))) {
  fail(`${WORKFLOW} runs ${p}, which does not exist. \`node --test\` exits 0 on a missing path, so this has been counting as a pass while testing nothing.`)
}
for (const p of want.filter((p) => !have.includes(p))) {
  fail(`${p} exists and is not in ${WORKFLOW} — it has never run in CI. Add it to the \`node --test\` list (enumerated, not globbed: \`node --test <dir>\` fails on Windows).`)
}
for (const p of have.filter((p, i) => have.indexOf(p) !== i)) {
  fail(`${WORKFLOW} lists ${p} more than once.`)
}

if (!process.exitCode) {
  console.log(`mjs suite list matches the tree: ${want.length} suite(s) — ${want.join(', ')}`)
}
