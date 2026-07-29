// The set of plugins is written down in three places that cannot import each other:
// the marketplace manifest (JSON), the bootstrap script (bash array), and the
// pasteable snippet in docs/setup-script.md (a bash for-loop inside a fenced block).
// A setup script cannot read the manifest — it runs before any checkout exists in the
// general case — so the duplication is structural. This makes it loud instead.
//
// The failure it catches is silent by construction: an unnamed plugin is simply never
// installed, and a session missing one behaves exactly like a session predating it.
import { readFileSync } from 'node:fs'

const fail = (msg) => { console.error(msg); process.exitCode = 1 }

const declared = JSON.parse(readFileSync('.claude-plugin/marketplace.json', 'utf8'))
  .plugins.map((p) => p.name)

// PLUGINS=(a b c) in the bootstrap script.
const bootstrap = readFileSync('scripts/bootstrap-plugins.sh', 'utf8')
  .match(/^PLUGINS=\(([^)]*)\)/m)?.[1].trim().split(/\s+/) ?? null

// `for p in a b c; do` in the documented snippet.
const snippet = readFileSync('docs/setup-script.md', 'utf8')
  .match(/^for p in ([^;]+); do$/m)?.[1].trim().split(/\s+/) ?? null

const sources = {
  'scripts/bootstrap-plugins.sh (PLUGINS=…)': bootstrap,
  'docs/setup-script.md (for p in …)': snippet,
}

const want = [...declared].sort()
for (const [where, got] of Object.entries(sources)) {
  if (got === null) {
    fail(`${where}: could not find the plugin list at all — the pattern this check greps for has moved, so it has been passing without reading anything.`)
    continue
  }
  const have = [...got].sort()
  const missing = want.filter((p) => !have.includes(p))
  const extra = have.filter((p) => !want.includes(p))
  if (missing.length) fail(`${where}: never installs ${missing.join(', ')} — declared in .claude-plugin/marketplace.json but not here.`)
  if (extra.length) fail(`${where}: installs ${extra.join(', ')}, which the marketplace does not offer — the install will fail at setup time.`)
}

if (!process.exitCode) console.log(`plugin lists agree: ${want.join(', ')}`)
