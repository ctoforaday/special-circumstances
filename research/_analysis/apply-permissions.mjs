import { readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

// 1. User settings: defaultMode auto (value-source rules require user scope).
const userP = join(process.env.USERPROFILE, '.claude', 'settings.json')
const user = JSON.parse(readFileSync(userP, 'utf8'))
user.permissions = user.permissions || {}
const prevMode = user.permissions.defaultMode
user.permissions.defaultMode = 'auto'
writeFileSync(userP, JSON.stringify(user, null, 2) + '\n')
console.log(`user settings: defaultMode ${prevMode === undefined ? '(unset)' : prevMode} -> auto`)

// 2. Local settings: remove the approved over-broad allow rules (exact string match).
const localP = 'C:/Users/gbloc/Projects/AgentOrange/.claude/settings.local.json'
const local = JSON.parse(readFileSync(localP, 'utf8'))
const remove = new Set([
  'Bash(gh api *)',
  "Bash(python -c ' *)",
  'Bash(DST_DIR="/c/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange" *)',
  'Bash(gh pr *)',
  'Bash(echo "=== Does the OneDrive PROJECT folder still exist? ===" *)',
  'Bash(git -C /c/Users/gbloc/Projects/special-circumstances fetch -q origin)',
])
const before = (local.permissions && local.permissions.allow || []).length
local.permissions.allow = local.permissions.allow.filter((r) => !remove.has(r))
writeFileSync(localP, JSON.stringify(local, null, 2) + '\n')
console.log(`local allow rules: ${before} -> ${local.permissions.allow.length} (removed ${before - local.permissions.allow.length} of ${remove.size} targeted)`)
