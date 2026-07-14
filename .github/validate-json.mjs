// CI guard: every tracked .json must parse — a malformed plugin.json / marketplace.json /
// requirements.json / .mcp.json fails silently at load time, so fail loudly here instead.
import { execSync } from 'node:child_process'
import { readFileSync } from 'node:fs'

const files = execSync('git ls-files *.json **/*.json').toString().trim().split('\n').filter(Boolean)
let bad = 0
for (const f of files) {
  try { JSON.parse(readFileSync(f, 'utf8')) }
  catch (e) { console.error(`${f}: ${e.message}`); bad++ }
}
if (bad) process.exit(1)
console.log(`${files.length} JSON files valid`)
