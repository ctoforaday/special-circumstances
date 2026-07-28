// Every agent, command and skill declares its capabilities in YAML frontmatter, and
// Claude Code's loader responds to a parse failure by dropping the WHOLE block —
// name, tools, skills, model, all of it — and loading the file anyway with empty
// metadata. Nothing raises. An agent that declared `tools: Read, Write, Bash` and
// `skills: [...]` runs with neither, behaving like an agent that never declared them,
// which is indistinguishable from an agent whose declarations simply do not work.
//
// probe.md sat broken this way: an unquoted description containing "skills: " ended
// the scalar early and broke the block. `claude plugin validate` catches it, and
// nothing in CI ran `claude plugin validate` — the declaration was policy with no
// mechanism. This is the mechanism. It is dependency-free because the repo is.
//
// It does NOT reimplement YAML. It rejects the constructs that make an unquoted
// scalar mean something other than it looks like, and demands quoting there. A
// frontmatter block outside that subset is treated as a defect even if some parser
// somewhere would accept it: the loader's verdict is the one that matters, and a
// value whose meaning depends on which parser reads it is already a bug.
import { readFileSync, globSync } from 'node:fs'

const files = [...globSync('plugins/**/*.md'), ...globSync('*.md')].sort()
const problems = []
let scanned = 0

for (const file of files) {
  const text = readFileSync(file, 'utf8')
  if (!text.startsWith('---\n')) continue // no frontmatter is fine; a partial one is not
  const end = text.indexOf('\n---', 3)
  if (end < 0) {
    problems.push(`${file}: frontmatter opens with --- and never closes; the loader reads the entire file as YAML.`)
    continue
  }
  scanned++
  const lines = text.slice(4, end).split('\n')
  for (const [i, line] of lines.entries()) {
    const at = `${file}:${i + 2}`
    const m = line.match(/^([A-Za-z_][\w-]*):\s+(\S.*)$/)
    if (!m) continue // blank, comment, list item, or a block-scalar continuation
    const [, key, value] = m
    if (/^["'[{>|]/.test(value)) continue // already quoted, flow-style, or a block scalar
    if (value.includes(': ')) {
      problems.push(`${at}: ${key} is an unquoted value containing ": " — YAML ends the scalar there and the whole block fails to parse, so every field in it is silently dropped at load time. Wrap the value in double quotes.`)
    } else if (/ #/.test(value)) {
      problems.push(`${at}: ${key} is an unquoted value containing " #" — YAML reads the rest as a comment, so the field loads truncated with nothing to show for it. Wrap the value in double quotes.`)
    }
  }
}

if (problems.length) {
  for (const p of problems) console.error(p)
  console.error(`\n${problems.length} frontmatter problem(s). These do not fail loudly at runtime; that is why they are checked here.`)
  process.exit(1)
}
console.log(`${scanned} frontmatter blocks valid`)
