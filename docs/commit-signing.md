# Commit signing on a workstation

Run `scripts/setup-commit-signing.sh` on any box that commits to this repository.
It is idempotent, so re-running it is also how you check a box is still set up.

```bash
./scripts/setup-commit-signing.sh                 # configure, register, verify
CHECK_ONLY=1 ./scripts/setup-commit-signing.sh    # audit an existing box, change nothing
```

## Why this needs a script at all

Losing commit signing is silent in both directions. Git keeps committing, GitHub
keeps accepting the push, CI stays green, and every commit lands **Unverified**
with nothing anywhere reporting a problem. Rebuild a workstation and the setup is
gone with no symptom to notice — which is precisely why it belongs in a script
you can re-run rather than in a runbook someone remembers.

## What it configures

| Setting | Purpose |
|---|---|
| `gpg.format=ssh` | Sign with an SSH key rather than GPG — one file, no keyring, no expiry |
| `user.signingkey` | The `.pub` half of `~/.ssh/id_ed25519_signing` |
| `commit.gpgsign=true` | Sign every commit, not the ones you remember to flag |
| `gpg.ssh.allowedSignersFile` | `~/.ssh/allowed_signers` — **local verification only** |

`user.name` and `user.email` are read, never written. The script signs whatever
identity the box already commits as.

## The three ways this fails silently

Each produces a good-looking setup and an Unverified badge, and none of them
raises an error at the moment you get it wrong.

**1. The key is registered for authentication instead of signing.** GitHub keeps
two separate lists. A key added only for auth signs nothing, and the UI shows a
key present either way. The script always passes `--type signing`.

**2. The token cannot manage signing keys.** A default `gh auth login` does not
carry `admin:ssh_signing_key`, and `gh ssh-key add` fails with a bare `HTTP 404`
that reads like a missing endpoint. Adding the scope is an interactive browser
flow, so no script can do it:

```bash
gh auth refresh -h github.com -s admin:ssh_signing_key
```

The script detects this, prints that command, keeps the local setup, and exits
non-zero rather than reporting a success it did not achieve.

**3. The commit email is not verified on the account holding the key.** This is
the one that survives everything else. GitHub stamps Verified only when the
commit's author email is a *verified* email on the *signing key's* account — so a
perfect signature under an unrecognised email is Unverified, with no error at any
step. `<id>+<account>@users.noreply.github.com` is always valid for the account
that owns it, and is the fallback when a custom address is not verified.

## `gh auth login` does not set commit authorship

Worth stating because the two look like one thing and are not:

- **`gh auth login`** decides who *pushes*, who opens pull requests, and which
  account `gh ssh-key add` files a key under.
- **`git config user.email`** decides who *authored* the commit — the only field
  GitHub matches against a signing key when deciding Verified.

A box logged in as one account while committing as another email will push fine
and show Unverified forever. When several identities share a box, each needs its
own key registered on its own account; `allowed_signers` holds one line per
identity, and the script appends rather than overwrites for exactly that reason.

## Verification is by signature, never by config

The script asserts the outcome, not the settings: it makes a real signed commit
in a throwaway repository — no real branch touched, no push needed — and reads
`git log -1 --format='%G?'`. Config that reads correctly and produces an
unverifiable commit is the failure being hunted, and it is indistinguishable from
success until something actually signs.

| `%G?` | Meaning | Fix |
|---|---|---|
| `G` | Good, trusted | — |
| `N` | Not signed at all | `commit.gpgsign` did not take |
| `U` | **Signature is good**, key untrusted | No `allowed_signers` line maps this email to this key |
| `E` | Cannot be checked | `gpg.ssh.allowedSignersFile` unset or unreadable |
| `B` | Bad signature | Key does not match the signature |

`U` is the trap: nothing is wrong with the key or the signature, and GitHub will
happily show Verified while the local check looks like a failure. Read it as a
statement about `allowed_signers`, not about signing.

The GitHub half needs one push, and the badge is not the measurement — the API is:

```bash
gh api repos/<owner>/<repo>/commits/<sha> --jq .commit.verification
# {"verified": true, "reason": "valid", ...}
```

## Env

| Variable | Default | Meaning |
|---|---|---|
| `SIGNING_EMAIL` | `git config user.email` | Identity to sign as, and the email GitHub matches |
| `SIGNING_KEY` | `~/.ssh/id_ed25519_signing` | Private key path; the `.pub` beside it is registered |
| `KEY_TITLE` | `$(hostname)` | Label in GitHub's key list |
| `SKIP_REGISTER=1` | — | Configure and verify locally; do not touch GitHub |
| `CHECK_ONLY=1` | — | Audit only; create nothing, change nothing |

Exit codes: `0` configured and verified · `1` setup failed · `2` prerequisite missing.

## Two deliberate choices

**It exits non-zero on failure**, unlike `scripts/bootstrap-plugins.sh`, which
must never fail a session it is booting. This script is run by an operator
provisioning a box, where a silent partial success — key on disk, config set, no
signature GitHub will honour — is the worst outcome available.

**The key has no passphrase.** `commit.gpgsign=true` signs *every* commit,
including those from headless and scheduled agent sessions with no terminal to
prompt at. A passphrase there does not harden the box; it converts every
unattended commit into a hang. Use one, with a loaded `ssh-agent`, only where a
human is present for every commit.

An existing key is never regenerated. A new key is not an upgrade — it orphans
every signature already made with the old one while the old public key sits on
the account still looking valid.

## Floors

git **2.34** (when SSH signing landed) and OpenSSH **8.2** (when `ssh-keygen -Y
sign`, which git shells out to, landed). Below either, everything appears to
succeed and produces commits nothing can verify. The script checks both and
refuses rather than proceeding.
