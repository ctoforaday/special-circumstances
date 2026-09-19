# NextToken AI Code Audit Rubric — Source Document (captured)

> **Provenance:** This is a captured copy of the source document referenced by
> [`nexttoken_audit_rubric_hook_plugin_input.md`](nexttoken_audit_rubric_hook_plugin_input.md).
> Original (may require access):
> <https://www.perplexity.ai/computer/a/97317781-4c97-4de0-88e9-e7c28186c99a>
> Captured 2026-07-11 so the research task is self-contained and not dependent on
> an access-gated link. Reference markers of the form `[^n]` are preserved from
> the source; the resolved reference list (with full source URLs) is at the end.
>
> **Note on scope:** the rubric defines **PASS 0 through PASS 6** (seven passes).
> PASS 6 (Iterative Regression Audit) is the "feedback loop security degradation"
> content that the iteration-degradation hook in the research prompt targets.

---

# AI-Generated Code Audit Framework: Architectural Flaws, Security Vulnerabilities & Code Review Rubric for NextToken

## Executive Summary

AI-generated codebases — particularly those produced through iterative sessions
with Claude models — exhibit a consistent taxonomy of structural, logical, and
security defects that differ qualitatively from human-authored code. Research
across 100+ large language models found that only 55% of AI-generated code was
secure, meaning nearly half of all outputs introduce known security flaws. A
separate peer-reviewed IEEE study established that critical vulnerabilities
increase by 37.6% after just five iterative AI refinement cycles — a phenomenon
termed "feedback loop security degradation". AI-generated code also carries
2.74x more vulnerabilities than human-authored code on a structural
basis.[^1][^2][^3][^4]

This report synthesizes empirically-documented anti-patterns specific to Claude
model outputs, classifies them into architectural, logic, and security
categories, and formalizes them into an actionable AI auditor rubric for
reviewing the NextToken codebase.

## Part I: Taxonomy of AI-Generated Code Failure Modes

### 1.1 Architectural Drift and Structural Anti-Patterns

Research covering 50 AI-generated repositories against 250 human-coded baselines
identified 10 distinct anti-patterns that occur at consistent statistical
frequencies:[^5]

| Anti-Pattern | Occurrence Rate | Description |
| --- | --- | --- |
| Excessive inline commenting | 90–100% | Comments everywhere, substituting for actual readable logic |
| Refactoring avoidance | 90–100% | Code grows linearly; AI adds features without refactoring existing structure |
| Edge case over-specification | 90–100% | Phantom checks for impossible or redundant conditions clutter core logic |
| Literal prompt fixation | 80–90% | Rigid interpretation of prompt with no architectural extrapolation |
| Bug déjà vu | 80–90% | Same logical errors recur across sessions due to lack of persistent memory |
| Shallow test coverage | 40–70% | High test count but low logical depth; tests pass but miss semantic edge cases |
| Phantom bugs | 40–70% | Guards added for non-existent edge cases, increasing noise without safety |
| Vanilla style (no architectural patterns) | 40–70% | No application of SOLID principles, design patterns, or separation of concerns |
| "It worked on my machine" syndrome | 40–70% | Environment-sensitive code lacking explicit dependency pinning or config validation |
| Return of the monolith | 40–70% | Functionality accumulates in single files/classes as context window grows |

Specific to Claude model outputs, code reviews have documented additional
structural problems:[^6]

- **Broken abstractions:** Claude creates well-defined interfaces and then
  immediately violates them by referencing concrete types throughout the
  codebase
- **Leaky layering:** Generated code appears architecturally sound on inspection
  but requires implementors to understand internals to use it safely — a
  defining trait of what Rob Zuber calls "the leaky abstraction"[^7]
- **Context-induced monolithism:** As conversation context fills up, Claude
  loses track of earlier modular decisions and begins writing new functionality
  into existing oversized files rather than creating properly scoped modules[^8]
- **Architectural drift under iteration:** Each refinement cycle introduces
  subtle design deviations from the original schema — removing constraints,
  relaxing type enforcement, or widening function scope[^9]

### 1.2 Asynchronous Logic and State Management Failures

Asynchronous code is one of the highest-risk surfaces in AI-generated
applications. Claude-generated code has been directly documented as unable to
reliably reason about race conditions, flaky polling intervals, and un-awaited
promises:[^10]

**Orphan State**

Orphan state emerges when a component or module initializes state that is never
cleaned up, never consumed, or is conditionally populated but unconditionally
referenced. Common manifestations in Claude outputs include:

- State variables declared in component initialization that are written to in
  some code paths but read without null guards in others
- Event listeners or subscriptions registered on mount with no corresponding
  teardown logic
- Asynchronous operations that mutate state after a component has unmounted (the
  "stale closure" variant)

**Unhandled Asynchronous Edge Cases**

Claude-generated code systematically produces swallowed errors in async
flows:[^11]

```javascript
// Classic Claude anti-pattern: error is caught, logged, and silently dropped
async function getOrder(orderId) {
  try {
    const user = await fetchUser(orderId);
    return await fetchOrder(user.id);
  } catch (err) {
    console.error("Error:", err); // logged but undefined returned
  }
}
// Caller receives undefined — no exception, no recovery signal
const order = await getOrder(42);
console.log(order.items); // TypeError: Cannot read properties of undefined
```

This pattern is endemic to AI-generated Node.js and React code: the catch block
logs a message and returns undefined, effectively swallowing the error while the
calling context has no indication that the operation failed. In Express.js,
swallowing errors in async route handlers means the global error middleware
never fires.[^11]

**Race Conditions and Non-Atomic Writes**

Claude Code's own `.claude.json` file has been flagged for race conditions
caused by non-atomic file writes with no locking — a concrete example of Claude
generating the same class of error it is vulnerable to producing. AI-generated
concurrent writes to shared state without mutex guards, semaphore controls, or
atomic operations represent a systematic weakness.[^12][^13]

### 1.3 Superficial and Cosmetic Error Handling

AI error handling is aesthetically correct but semantically hollow. Research
identifies the pattern as: error handling that "reveals too much, or handles too
little":[^14]

- **Over-verbose error propagation:** Stack traces, internal file paths, and
  database schema details exposed directly in HTTP responses — a finding
  confirmed across 53% of critical vulnerabilities in IDOR/authorization bug
  research[^15]
- **Catch-and-discard:** try-catch blocks that log without rethrowing, returning
  control flow without a usable fallback
- **Symmetric error messages:** Every error produces the same generic message
  regardless of domain context, making debugging impossible while maintaining a
  false appearance of robustness
- **Missing input boundary validation:** AI-generated functions accept inputs
  and use them directly without null/type/range checks at function boundaries —
  missing input validation (CWE-20) is the most common security flaw in
  LLM-generated code across all languages and models[^9]

## Part II: Security Vulnerability Taxonomy

### 2.1 Injection Vulnerabilities

AI models are trained on public codebases that include insecure patterns. The
result is that they reproduce those patterns unless explicitly constrained. The
statistical breakdown of AI-generated injection flaws:[^16][^17]

| Vulnerability | Occurrence Rate | CWE |
| --- | --- | --- |
| SQL Injection (string concatenation) | 17–28% | CWE-89 |
| XSS / Input sanitization failure | 11–19% | CWE-79 |
| Broken authentication / missing token validation | 9–15% | CWE-306 |
| Hardcoded secrets (API keys, DB credentials) | 6–12% | CWE-798 |
| Weak or outdated cryptography | 4–9% | CWE-327 |

SQL injection specifically occurs when AI uses unsafe string concatenation
rather than parameterized queries — even in codebases where parameterization is
used elsewhere, AI-generated additions commonly regress to interpolation.[^17]

### 2.2 Authentication and Authorization Failures

A DryRun Security report found AI coding agents systematically producing broken
access control, insecure authentication setups, hard-coded JWT secrets, and
missing token revocation logic. The problem is architectural rather than
syntactic: a typical Claude prompt like "hook up to a database and display user
scores" yields code that bypasses authentication and authorization
entirely.[^18][^9]

Authorization bugs represent 53% of critical vulnerabilities in production
AI-generated applications, yet are detected by traditional SAST tools at a rate
near zero. Key manifestations:[^15]

- **IDOR (Insecure Direct Object Reference):** Sequential IDs or predictable
  resource identifiers exposed directly in routes without ownership verification
- **Missing server-side authorization:** Checks applied client-side only, easily
  bypassed
- **Overly permissive CORS:** Wildcard `*` settings generated by default[^19]
- **JWT handling errors:** Secrets hardcoded rather than loaded from environment
  variables; missing expiry validation; missing algorithm pinning

### 2.3 Insecure Data Handling

AI models suggest hardcoded credentials and insecure patterns because those
patterns appear frequently in public training data. Specific insecure data
handling patterns include:[^20][^21][^16]

- **Hardcoded secrets:** API keys, database passwords, and JWT signing keys
  embedded directly in source files rather than loaded from environment
  variables
- **Cleartext transmission:** HTTP used where HTTPS is required; credentials
  transmitted in query strings or response bodies
- **Insecure cryptography:** MD5 or SHA-1 used for password hashing;
  `Math.random()` used for token generation instead of cryptographically secure
  random functions
- **Sensitive data in logs:** `console.log()` statements left in production code
  that may output PII, tokens, or internal architecture details[^19]
- **Debug modes active in production:** Development flags not gated behind
  environment checks

### 2.4 Dependency and Supply Chain Vulnerabilities

AI models suggest libraries with known CVEs patched after the model's training
cutoff, effectively re-introducing resolved vulnerabilities. Additionally:[^9]

- **Hallucinated dependencies ("slopsquatting"):** AI fabricates package names
  that don't exist; attackers can register these names in npm/PyPI with
  malicious code[^14][^9]
- **Dependency overuse:** Simple prompts yield complex dependency trees — a
  "To-do list app" prompt generated 2–5 backend dependencies depending on model
  used[^9]
- **Outdated locked versions:** AI suggests specific version pins that were
  current at training time but are now deprecated or vulnerable

## Part III: Claude-Specific Code Generation Patterns

### 3.1 The Feedback Loop Degradation Problem

The most counterintuitive finding in AI code auditing is that security worsens
as you iterate. The IEEE-ISTAS 2025 peer-reviewed study found four distinct
prompting styles produce different vulnerability trajectories:[^3][^22]

| Prompting Style | Security Trajectory |
| --- | --- |
| Efficiency-Focused (EF) | Strips validation/auth logic in pursuit of brevity |
| Feature-Focused (FF) | Adds attack surface with each feature without security review |
| Security-Focused (SF) | Best outcome but still degrades over 5+ iterations |
| Ambiguous Improvement (AI) | Worst outcome — roulette with attack surface[^22] |

This means the NextToken codebase, if iteratively developed with Claude, likely
contains latent vulnerabilities introduced by the improvement process itself —
code that was more secure at generation step 1 than at the final state.

### 3.2 Context Window Decay Effects

As Claude's context window fills, output quality degrades in specific
ways:[^23][^8]

- **Naming convention drift:** Variable and function names become inconsistent
  mid-file; Hungarian notation appears then disappears; camelCase and snake_case
  mixed
- **Pattern abandonment:** An established pattern (e.g., repository pattern,
  factory function) is used in early files and then abandoned in later files
  without deprecation
- **Self-contradiction:** Claude generates a utility function, then generates a
  near-duplicate 200 lines later because it has lost context of the earlier
  definition
- **Assumption inheritance:** Claude makes assumptions about how other code
  sections behave; as context fills, those assumptions become stale, and
  integration edges silently break

### 3.3 The "Vanilla Style" and Over-Abstraction Duality

Claude exhibits a paradoxical anti-pattern: it simultaneously generates
under-architected functional code (vanilla style, no patterns) and
over-abstracted structural code (unnecessary indirection layers) depending on
prompt framing:[^24][^5]

- **Under-architected:** No separation of concerns; business logic, data access,
  and presentation mixed in single functions; no service layer
- **Over-abstracted:** Factory-of-factory patterns, abstract base classes with
  single implementations, interfaces that wrap concrete types adding no
  isolation benefit[^6]

The key diagnostic is whether the abstraction hides complexity or merely
relocates it — Claude's abstractions frequently relocate rather than
encapsulate, creating leaky layers that force consumers to understand
implementation details.[^7]

## Part IV: NextToken AI Auditor Rubric

This rubric is structured as a strict, ordered sequence of audit passes. Each
pass is designed for an AI auditor operating on the NextToken codebase. Every
check instructs the auditor to actively hunt for structural evidence, not
surface-level compliance.

### PASS 0 — Pre-Audit Inventory (Orientation Before Review)

**Objective:** Map the codebase topology before any substantive review begins.

**Step 0.1 — Generate structural map**

Enumerate all files, modules, and directories. For each module, identify: what
it exports, what it imports, and what calls it. Flag any module that imports from
more than 5 sources (likely God Module). Flag any module imported by more than 10
consumers (critical shared dependency, highest audit priority).

**Step 0.2 — Identify AI generation markers**

Search for the following signals of AI authorship: (a) excessive inline comments
explaining trivial logic, (b) `TODO:` and `FIXME:` comments never resolved,
(c) near-duplicate functions separated by 100+ lines, (d) code blocks that use
one style convention then switch mid-file.

**Step 0.3 — Establish iteration depth estimate**

Inspect git blame or commit history. If only a small number of commits exist for
a large codebase, assume high AI generation ratio. The more AI commits without
human-authored changes, the higher the probability of feedback loop security
degradation.

### PASS 1 — Architectural Integrity Audit

**Objective:** Detect structural inconsistencies, pattern abandonment, and
responsibility leakage.

**Step 1.1 — Identify orphan modules**

For every file or module in the project, trace whether it has at least one
active, non-dead-code caller. Flag all modules with zero callers as dead
modules. Cross-reference against test files — a module tested but not called in
production logic is a candidate for dead code masquerading as live.[^25][^26]

**Step 1.2 — Identify orphan state declarations**

In every stateful component or class: list all state variables declared at
initialization. For each, trace forward: Is it written to in all code paths? Is
it read with null guards? Is it cleaned up on destroy/unmount? Any state variable
that is written conditionally but read unconditionally is an orphan state
candidate.

**Step 1.3 — Pattern consistency audit**

Identify the primary architectural pattern (repository pattern, MVC, layered
architecture). Verify it is applied consistently across ALL modules. Flag modules
that deviate — these are typically later-generated additions where Claude lost
context of the original pattern.[^27]

**Step 1.4 — Abstraction layer audit**

For every interface or abstract class: determine whether removing it and using
the concrete type directly would change any behavior. If the answer is no, the
abstraction is cosmetic and likely an AI anti-pattern. Flag all interfaces with
only one implementation that add no isolation benefit.[^6][^7]

**Step 1.5 — Dead code path detection**

Using static analysis and control flow tracing: identify all if/else branches
where the else or a specific condition is unreachable. Flag all functions whose
return value is never consumed. Identify all imported modules whose exported
symbols are never referenced. These represent the AI's tendency to generate
"phantom guards" and speculative code paths.[^5][^25]

### PASS 2 — Asynchronous Logic and State Machine Audit

**Objective:** Surface every async edge case that Claude characteristically
fails to handle.

**Step 2.1 — Async call inventory**

Map every async function, Promise, `.then()` chain, and await expression. For
each: verify a `.catch()` or try-catch block exists. Any async operation without
a handler is a critical defect.[^28][^29]

**Step 2.2 — Error propagation trace**

For every catch block in the codebase: determine what happens after the catch.
Acceptable outcomes are: rethrowing the error, returning a typed fallback value,
calling a centralized error handler, or triggering a state update that notifies
the caller. Unacceptable outcomes: logging only and returning undefined;
swallowing the error entirely. Flag every catch block that does not propagate a
meaningful signal to the caller.[^14][^11]

**Step 2.3 — Race condition surface**

Identify all locations where two or more async operations write to shared state
(React state, global variables, file system, database records). For each: verify
a locking or serialization mechanism exists. In particular check: event handlers
that can fire multiple times before a prior invocation completes; polling loops
without cancellation tokens; WebSocket message handlers that mutate state without
queuing.[^13][^10][^12]

**Step 2.4 — State lifecycle validation**

In component-based architectures: verify that every subscription, event
listener, WebSocket connection, or timer established in an initialization hook
has a corresponding teardown in the component's cleanup/unmount handler. Missing
teardowns cause memory leaks and stale state mutations in AI-generated
codebases.[^30]

**Step 2.5 — Boundary condition coverage**

For every async function that processes a collection: manually trace what happens
when the collection is empty, null, or a single item. AI-generated code
systematically misses empty-collection edge cases, zero-value numeric inputs, and
null API responses.[^31][^14]

### PASS 3 — Security Vulnerability Audit

**Objective:** Hunt for insecure data handling, injection surfaces, and
authorization failures typical of LLM output.

**Step 3.1 — Secret and credential scan**

Scan all source files for: JWT secrets, API keys, database connection strings,
OAuth client secrets, encryption keys, and passwords. Any credential value
assigned as a literal string (not loaded from `process.env` or a secrets
manager) is a critical finding. Also scan `.env.example` files — AI frequently
populates these with real values.[^20][^17][^19]

**Step 3.2 — Injection surface audit**

Identify every location where user-controlled input reaches: SQL query
construction, shell command execution, file system path resolution, HTML
rendering, or template evaluation. For each, verify parameterized queries, safe
APIs, or explicit escaping is applied. String concatenation of user input into
any query or command is a critical defect regardless of whether the input appears
"clean".[^17][^9]

**Step 3.3 — Authentication and authorization completeness**

Map every route, endpoint, and API function. For each: (a) is authentication
required and enforced server-side? (b) is authorization checked at the resource
level (i.e., does the authenticated user own this specific resource)? (c) are JWT
tokens validated for signature, expiry, issuer, and algorithm? Missing
authorization at the resource level — the IDOR pattern — is the most common
critical vulnerability class in AI-generated code, at 53% of critical
findings.[^18][^15]

**Step 3.4 — CORS and HTTP security headers**

Inspect CORS configuration: flag any wildcard (`*`) origin policy on
authenticated endpoints. Verify security headers are set: `Content-Security-Policy`,
`X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`.
AI-generated web apps routinely omit these.[^19]

**Step 3.5 — Cryptographic function audit**

Identify all hashing, encryption, and random number generation. Flag: MD5 or
SHA-1 for password hashing (use Argon2/bcrypt/PBKDF2); `Math.random()` or
`Date.now()` for token generation (use `crypto.randomBytes`); custom encryption
implementations (use established libraries); symmetric keys embedded in source
(use key management services).[^17][^19]

**Step 3.6 — Dependency hallucination check**

For every dependency listed in `package.json`, `requirements.txt`, `go.mod`, or
equivalent: verify the package exists on its respective registry, has active
maintenance, and has no critical CVEs. AI models suggest packages that don't
exist — any package that cannot be found on the official registry is a critical
supply chain risk.[^31][^14][^9]

### PASS 4 — Logic and Business Rule Integrity Audit

**Objective:** Identify semantic errors — code that is syntactically correct but
logically wrong.

**Step 4.1 — Conditional logic exhaustiveness**

For every conditional block: enumerate all possible inputs and trace which branch
each follows. Identify: conditions that are always true or always false;
conditions evaluated in wrong order causing earlier false matches; conditions
that use assignment (`=`) instead of comparison (`==`/`===`).

**Step 4.2 — Return value consistency**

For every function: verify all code paths return a value of the same type.
AI-generated functions commonly return a typed value on the success path and
undefined on the error path without documenting this in the type signature.[^11]

**Step 4.3 — Data flow integrity**

Trace the flow of each user-facing input from entry point to storage/output.
Identify every transformation applied. Verify: (a) validation occurs at the entry
boundary, not deep in the stack; (b) data is not deserialized then re-serialized
with different semantics; (c) output encoding is applied at the exit boundary.

**Step 4.4 — Transaction and atomicity audit**

Identify all multi-step operations that modify state (database writes, file
operations, external API calls in sequence). For each: verify failure of step N
triggers rollback of steps 1 through N-1. AI-generated code commonly omits
compensation logic, leaving data in partially-modified states.[^32]

**Step 4.5 — Concurrency invariant audit**

Identify all globally shared state (module-level variables, singleton instances,
caches). For each: verify no concurrent write operations can interleave to
produce inconsistent state. In non-concurrent environments: verify state is reset
between logical operations rather than carried across request boundaries.

### PASS 5 — Code Quality and Maintainability Audit

**Objective:** Detect the structural debt that makes future vulnerabilities
invisible.

**Step 5.1 — Duplication scan**

Identify code blocks of 10+ lines that appear more than once. AI-generated code
commonly duplicates logic across files when context is lost. Each duplicate is
both a maintenance risk and an indicator that one copy may receive a security fix
while the other does not.[^31]

**Step 5.2 — Complexity threshold enforcement**

Measure cyclomatic complexity for all functions. Flag any function exceeding
cyclomatic complexity 10 or cognitive complexity 15 — these are the thresholds
established for AI-generated code quality gates. High-complexity AI-generated
functions are systematically under-tested relative to their branch count.[^5]

**Step 5.3 — Test quality audit (not just coverage)**

Distinguish between: tests that verify behavior (do they assert specific
outcomes?), tests that verify presence (do they only check that functions run
without error?), and tests that were generated by AI to match AI-generated code
(circular validation). Require 80% meaningful behavioral test coverage for new
code with duplication below 3%.[^31]

**Step 5.4 — Logging security audit**

Audit all log statements. Flag: any log statement that outputs request body
content, user credentials, tokens, PII, or internal stack traces to
external-facing log sinks. AI-generated code routinely leaves debug
`console.log()` statements in production paths that leak sensitive
information.[^33][^19]

**Step 5.5 — Environment configuration validation**

Verify a startup validation routine exists that checks all required environment
variables before the application begins serving traffic. AI-generated code
commonly references `process.env.VARIABLE` throughout the codebase without
verifying the variable exists, causing silent failures or insecure fallback
behavior.[^14]

### PASS 6 — Iterative Regression Audit (AI-Specific)

**Objective:** Detect vulnerability patterns introduced by AI's own improvement
cycles.

**Step 6.1 — Before/after security regression check**

If commit history is available: identify commits where an AI assistant modified
existing security-sensitive code (auth middleware, input validation,
cryptography). For each such commit, verify the modification did not remove or
weaken a pre-existing security control. The 37.6% increase in critical
vulnerabilities across 5 iterations means each AI "improvement" is a regression
candidate.[^22][^3]

**Step 6.2 — "Security-focused" regression trap**

Even code generated with explicit security prompts degrades over iteration. For
any code block that implements security-critical logic: verify the implementation
is complete and not a surface-level approximation. Common trap: AI adds a JWT
validation check but omits algorithm verification (`alg` field), or adds SQL
parameterization to new queries but misses pre-existing raw queries in the same
file.[^18][^17]

**Step 6.3 — Context-boundary integrity check**

Identify all integration points between modules that were likely generated in
different sessions (different naming conventions, different error handling styles,
different abstraction levels). These inter-session boundaries are the
highest-probability location for silent contract violations — where one module
produces a value the other module consumes with incompatible assumptions.

## Part V: Audit Severity Classification

| Severity | Criteria | Required Action |
| --- | --- | --- |
| Critical | Hardcoded secret; auth bypass; SQL injection; IDOR; RCE surface | Block merge; immediate remediation |
| High | Swallowed async error on production path; missing CORS restriction; weak crypto | Fix before next release |
| Medium | Orphan state without null guard; missing input validation on non-critical path; dead code module | Fix within sprint |
| Low | Excessive inline comments; naming inconsistency; duplicate logic block | Fix in maintenance cycle |
| Informational | Cosmetic abstraction; speculative phantom guard; over-specified edge case | Document; refactor when convenient |

## Part VI: Toolchain Integration

The following tools map directly to audit passes and should be integrated into
the NextToken CI pipeline:[^34][^35][^14]

| Tool | Pass | Capability |
| --- | --- | --- |
| CodeQL | 3 | Static security analysis: injection, auth patterns |
| Semgrep | 1, 3 | Pattern detection: API hallucinations, deprecated patterns, custom rules |
| ESLint / Pyflakes | 1, 5 | Dead code, unused imports, unreachable branches |
| SonarQube | 1, 4, 5 | Cyclomatic complexity, duplication, maintainability index |
| Dependabot / Snyk | 3.6 | Dependency CVE tracking, hallucinated package detection |
| Gitleaks / TruffleHog | 3.1 | Hardcoded secrets and credential scanning |
| OWASP Dependency Check | 3.6 | Supply chain vulnerability mapping |
| Istanbul / coverage.py | 5.3 | Behavioral test coverage analysis |

## Conclusion

The NextToken codebase, as an AI-generated product, should be audited with the
expectation that it contains architecturally invisible vulnerabilities — code
that looks correct, passes syntax checks, and may pass basic tests, yet violates
security invariants or logical contracts. The six passes in this rubric are
sequenced to move from orientation to architecture to async logic to security to
quality to regression — each layer uncovering issues that prior passes cannot
surface alone.[^32][^9]

The most critical operational principle for the AI auditor is: do not trust
appearance of correctness. AI-generated code optimizes for linguistic and
functional coherence, not for security or logical completeness. Every try-catch
must be traced to its resolution. Every auth check must be verified at the
resource level, not just the route level. Every dependency must be confirmed to
exist. The rubric's value lies not in its checklist form, but in its insistence
that the auditor trace execution paths and data flows rather than reading for
surface-level compliance.[^22]

## References

> Reference URLs supplied from the original source document. Descriptions are the
> one-line snippets shown on the source page (truncated as they appeared there).

[^1]: [AI-Generated Code Security Risks: What Developers Must Know](https://www.veracode.com/blog/ai-generated-code-security-risks/) — AI-generated code introduces security flaws in 45% of cases.
[^2]: [AI-Generated Code Security Risks — Why Vulnerabilities Increase 2.74x and How to Prevent Them](https://www.softwareseni.com/ai-generated-code-security-risks-why-vulnerabilities-increase-2-74x-and-how-to-prevent-them/) — AI-generated code has 2.74x more vulnerabilities than code written by humans.
[^3]: [[2506.11022] Security Degradation in Iterative AI Code Generation](https://arxiv.org/abs/2506.11022) — 37.6% increase in critical vulnerabilities after just five iterations.
[^4]: [Peer-reviewed and accepted in IEEE-ISTAS 2025 (Reddit)](https://www.reddit.com/r/vibecoding/comments/1p8ohtz/peerreviewed_and_accepted_in_ieeeistas_2025/) — IEEE peer-reviewed study on "feedback loop security degradation".
[^5]: [Understanding Anti-Patterns and Quality Degradation in AI-Generated Code](https://www.softwareseni.com/understanding-anti-patterns-and-quality-degradation-in-ai-generated-code/) — anti-patterns with before/after examples.
[^6]: [Code Smells in Generated Code — some of the patterns (r/ClaudeAI)](https://www.reddit.com/r/ClaudeAI/comments/1ozhq5e/code_smells_in_generated_code_some_of_the_patterns/) — broken abstractions.
[^7]: [Rob Zuber: AI code generation is a Leaky Abstraction (LinkedIn)](https://www.linkedin.com/posts/randyshoup_brilliant-insight-from-rob-zuber-ai-code-activity-7371843679268859905-_cyl) — generated code is a leaky abstraction.
[^8]: [Why LLMs Can't Really Build Software (YouTube)](https://www.youtube.com/watch?v=VjR4JbB6lMs) — LLMs and direct problem solving vs. building software.
[^9]: [The Most Common Security Vulnerabilities in AI-Generated Code (Endor Labs)](https://www.endorlabs.com/learn/the-most-common-security-vulnerabilities-in-ai-generated-code) — injection flaws and emerging risks.
[^10]: [Hands on the wheel: live-blogging a Claude Code programming session](https://jvaneyck.wordpress.com/2025/07/24/hands-on-the-wheel-live-blogging-a-claude-code-programming-session/) — race conditions / flaky polling intervals.
[^11]: [25 Ways Claude Code Can Break Your App as a Non-Technical … (LinkedIn)](https://www.linkedin.com/posts/jeffhsipe_25-ways-claude-code-can-silently-break-activity-7452066952618876928-u9lv) — stale React state reads, race conditions.
[^12]: [[BUG] .claude.json race condition — reported 8 times since June (GitHub)](https://github.com/anthropics/claude-code/issues/28922) — non-atomic writes corrupting .claude.json.
[^13]: [I built a "Traffic Light" to prevent race conditions when running Claude Code / Agent Swarms (Reddit)](https://www.reddit.com/r/ClaudeCode/comments/1r4k1pc/i_built_a_traffic_light_to_prevent_race/).
[^14]: [Debugging AI-Generated Code: 8 Failure Patterns & Fixes (Augment Code)](https://www.augmentcode.com/guides/debugging-ai-generated-code-8-failure-patterns-and-fixes) — hallucinated APIs, security vulnerabilities.
[^15]: [Authorization Bugs Are Having Their SQL Injection Moment (ZeroPath)](https://zeropath.com/blog/idor-crisis-2025) — 94% of applications have broken access control.
[^16]: [AI-Generated Code Security: Security Risks and Opportunities (Apiiro)](https://apiiro.com/blog/ai-generated-code-security/) — insecure patterns including hardcoded API keys.
[^17]: [What the Data Really Shows | Eliahu (Eli) Assif (LinkedIn)](https://www.linkedin.com/posts/ph-d-eliahu-eli-assif-amar-97184540_security-vulnerabilities-in-ai-generated-activity-7402974186953953281-D6V1) — security vulnerabilities in AI-generated code.
[^18]: [A new report from DryRun Security examined how AI coding agents … (Instagram)](https://www.instagram.com/p/DVy0jL3jN7-/) — broken access control, hard-coded JWT secrets.
[^19]: [Secure Vibe Coding Guide | Become a Citizen Developer (CSA)](https://cloudsecurityalliance.org/blog/2025/04/09/secure-vibe-coding-guide) — regular security audits and penetration testing.
[^20]: [Managing the Risk of Hardcoded Secrets in AI-Generated Code (Cycode)](https://cycode.com/blog/managing-the-risk-of-hardcoded-secrets-in-ai-generated-code/).
[^21]: [Why AI-Generated Code Creates Hidden Security Debt (Invicti)](https://www.invicti.com/blog/web-security/why-ai-generated-code-creates-hidden-security-debt) — models reproduce insecure patterns from public codebases.
[^22]: [Iterative AI Code Generation — Exploring the Study (Symbiotic Security)](https://www.symbioticsec.ai/blog/exploring-security-degradation-iterative-ai-code-generation).
[^23]: [Anti-patterns while working with LLMs (Hacker News)](https://news.ycombinator.com/item?id=46080597) — Claude Code with the SolidWorks SDK.
[^24]: [Abstraction Is the New Literacy for Developers (DEV Community)](https://dev.to/rohit_gavali_0c2ad84fe4e0/abstraction-is-the-new-literacy-for-developers-40i8) — output quality proportional to abstraction quality.
[^25]: [Guide to Dead Code Identification and Removal 2026 (Penser)](https://pensero.ai/blog/dead-code) — no single tool finds all dead code.
[^26]: [Dead Code: A Practical Guide for Engineering Leaders (Axify)](https://axify.io/blog/dead-code) — how dead code forms and signals delivery issues.
[^27]: [Vibe coding: the good, the bad, and the mud (LinkedIn)](https://www.linkedin.com/posts/pratulg_big-ball-of-mud-is-a-casually-even-haphazardly-activity-7326597217161535489-3gWV) — guiding the LLM to adhere to architectural decisions.
[^28]: [How to Fix 'UnhandledPromiseRejectionWarning' in Node.js (OneUptime)](https://oneuptime.com/blog/post/2026-01-25-fix-unhandled-promise-rejection-warning/view).
[^29]: [How to fix: Unhandled Promise Rejection (DEV Community)](https://dev.to/failwarn/how-to-fix-unhandled-promise-rejection-3on6).
[^30]: [Handling rejected promises in React with an error boundary](http://eddiewould.com/2021/28/28/handling-rejected-promises-error-boundary-react/).
[^31]: [Code Review Checklist for AI-Generated Code (ClackyAI Blog)](https://clacky.ai/blog/code-review-checklist-ai-generated-code) — run tests, scan for vulnerabilities.
[^32]: [Understanding Security Risks in AI-Generated Code (CSA)](https://cloudsecurityalliance.org/blog/2025/07/09/understanding-security-risks-in-ai-generated-code) — 62% of AI-generated code solutions contain design flaws or known security issues.
[^33]: [The Ultimate Code Review Checklist: 7 Steps for 2025 (Zemith.com)](https://www.zemith.com/blogs/code-review-checklist).
[^34]: [We vibe coded a path tracer: Here's how we used static analysis (Datadog)](https://www.datadoghq.com/blog/delivery-guardrails-for-ai-generated-code/).
[^35]: [Security in the Vibe Code Era, Part 1: There's No Gate to Keep (GuidePoint Security)](https://www.guidepointsecurity.com/blog/security-vibe-code-part-one-no-gate-to-keep/).
