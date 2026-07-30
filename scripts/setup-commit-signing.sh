#!/usr/bin/env bash
# Configure SSH commit signing for this machine, and prove it works. Idempotent —
# safe to re-run, and re-running is the intended way to check a box is still set up.
#
# Rebuilding a workstation loses this silently: git keeps committing, GitHub keeps
# accepting, and every commit lands "Unverified" with nothing anywhere reporting a
# problem. That is the failure this script exists to make impossible to forget.
#
# THIS SCRIPT EXITS NON-ZERO ON FAILURE, unlike its sibling
# scripts/bootstrap-plugins.sh, which must never fail a session it is booting. This
# one is run by an operator provisioning a box, where a silent partial success is
# the worst outcome available: a key on disk, config set, and no signature GitHub
# will honour. Loud beats convenient here.
#
# WHY SSH AND NOT GPG: the key is one file, `gh` can register it, and there is no
# keyring, agent, or expiry to carry. GitHub treats both identically.
#
# Env:
#   SIGNING_EMAIL   the email commits are authored with, and the one GitHub matches
#                   against the signing key's account.
#                   default: `git config user.email`
#   SIGNING_KEY     private key path; the .pub beside it is the signing key.
#                   default: ~/.ssh/id_ed25519_signing
#   KEY_TITLE       label shown in GitHub's key list. default: $(hostname)
#   SKIP_REGISTER=1 configure and verify locally, do not touch GitHub
#   CHECK_ONLY=1    verify an existing setup; change nothing, create nothing
#
# Exit codes: 0 configured and verified · 1 setup failed · 2 prerequisite missing

set -euo pipefail

SIGNING_KEY="${SIGNING_KEY:-$HOME/.ssh/id_ed25519_signing}"
PUB="$SIGNING_KEY.pub"
ALLOWED="$HOME/.ssh/allowed_signers"
KEY_TITLE="${KEY_TITLE:-$(hostname)}"

log()  { printf '[commit-signing] %s\n' "$*"; }
warn() { printf '[commit-signing] WARNING: %s\n' "$*" >&2; }
# $1 is the message and $2 the exit code — "$1", never "$*", or the code is printed
# as part of the error text.
die()  { printf '[commit-signing] ERROR: %s\n' "$1" >&2; exit "${2:-1}"; }

# --- prerequisites ----------------------------------------------------------
# Both floors are exact, not defensive: SSH signing landed in git 2.34, and
# `ssh-keygen -Y sign` — which git shells out to — landed in OpenSSH 8.2. Below
# either, everything below "succeeds" and produces commits nothing can verify.
command -v git >/dev/null || die "git not on PATH" 2
command -v ssh-keygen >/dev/null || die "ssh-keygen not on PATH" 2

# `sort -V -C` asks "is this input already in version order?", so the FLOOR must be
# printed first: the check passes exactly when floor <= actual. Printed the other way
# round it rejects every version that is new enough, which is the failure it is meant
# to catch, inverted — and it looks like a correct check right up until you run it.
at_least() { printf '%s\n%s\n' "$1" "$2" | sort -V -C; }

git_ver="$(git --version | awk '{print $3}')"
at_least 2.34 "$git_ver" \
  || die "git $git_ver is too old for SSH signing; 2.34 is the floor" 2

ssh_ver="$(ssh -V 2>&1 | sed -n 's/^OpenSSH_\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p')"
if [ -n "$ssh_ver" ]; then
  at_least 8.2 "$ssh_ver" \
    || die "OpenSSH $ssh_ver has no \`ssh-keygen -Y sign\`; 8.2 is the floor" 2
fi

SIGNING_EMAIL="${SIGNING_EMAIL:-$(git config --get user.email || true)}"
[ -n "$SIGNING_EMAIL" ] \
  || die "no signing email: set SIGNING_EMAIL, or \`git config --global user.email\`" 2

# --- check-only -------------------------------------------------------------
if [ "${CHECK_ONLY:-}" = "1" ]; then
  # if/else rather than `A && B || C` (shellcheck SC2015): in the && || form, C also
  # runs when A succeeded and B failed, so a failing log line would be reported as a
  # configuration defect. An audit that invents failures is worse than none.
  fail=0
  for kv in "gpg.format=ssh" "commit.gpgsign=true"; do
    k="${kv%%=*}"; want="${kv#*=}"; got="$(git config --get "$k" || true)"
    if [ "$got" = "$want" ]; then
      log "$k = $got"
    else
      warn "$k = ${got:-unset}, want $want"; fail=1
    fi
  done
  if [ -f "$PUB" ]; then log "key present: $PUB"; else warn "no key at $PUB"; fail=1; fi
  if [ -f "$ALLOWED" ]; then log "allowed signers: $ALLOWED"; else warn "no $ALLOWED"; fail=1; fi
  [ "$fail" -eq 0 ] || die "not fully configured — re-run without CHECK_ONLY"
  log "configured; verifying by signing a commit"
else
  # --- key ------------------------------------------------------------------
  # Never regenerate over an existing key. A new key is not an upgrade: it orphans
  # every signature already made with the old one, and the old public key stays on
  # the account looking valid. Reuse is the safe default; deleting is deliberate.
  mkdir -p "$HOME/.ssh" && chmod 700 "$HOME/.ssh"
  if [ -f "$SIGNING_KEY" ]; then
    log "reusing existing key $SIGNING_KEY"
  else
    # No passphrase, deliberately. commit.gpgsign=true signs EVERY commit, including
    # those made by headless and scheduled agent sessions that have no terminal to
    # prompt at — a passphrase there does not harden the box, it just converts every
    # unattended commit into a hang. Use a passphrase plus a loaded ssh-agent only
    # where a human is present for every commit.
    log "creating $SIGNING_KEY (ed25519, no passphrase — see the comment above)"
    ssh-keygen -t ed25519 -f "$SIGNING_KEY" -N "" -C "$SIGNING_EMAIL signing ($KEY_TITLE)" -q
  fi
  [ -f "$PUB" ] || die "private key exists but $PUB does not; cannot continue"

  # --- git config -----------------------------------------------------------
  git config --global gpg.format ssh
  git config --global user.signingkey "$PUB"
  git config --global commit.gpgsign true

  # allowed_signers is what makes `git log --format=%G?` answer G instead of E.
  # Without it git reports "needs to be configured" on a PERFECTLY GOOD signature,
  # which reads as a signing failure and sends you debugging the wrong thing.
  # GitHub never consults this file — it is purely local verification.
  key_material="$(awk '{print $1, $2}' "$PUB")"
  touch "$ALLOWED" && chmod 600 "$ALLOWED"
  if grep -qF "$SIGNING_EMAIL $key_material" "$ALLOWED" 2>/dev/null; then
    log "allowed_signers already lists $SIGNING_EMAIL"
  else
    # Append, never overwrite: a box may sign as more than one identity, and each
    # needs its own line. Clobbering silently disarms local verification for the rest.
    printf '%s %s\n' "$SIGNING_EMAIL" "$key_material" >> "$ALLOWED"
    log "added $SIGNING_EMAIL to $ALLOWED"
  fi

  log "signing as $SIGNING_EMAIL with $(ssh-keygen -lf "$PUB" | awk '{print $2}')"
fi

git config --global gpg.ssh.allowedSignersFile "$ALLOWED"

# --- register with GitHub ---------------------------------------------------
# A SIGNING key is a separate entry from an AUTHENTICATION key. Adding the same key
# for auth only — the usual mistake — leaves signatures valid locally and Unverified
# on GitHub, with no error at any step to say why.
register() {
  [ "${SKIP_REGISTER:-}" = "1" ] && { log "SKIP_REGISTER=1 — not registering with GitHub"; return 0; }

  if ! command -v gh >/dev/null; then
    warn "gh not on PATH — register the key by hand:"
    warn "  https://github.com/settings/ssh/new  (Key type: Signing Key)"
    warn "  $(cat "$PUB")"
    return 0
  fi

  # The scope is the failure this whole block exists to report. A `gh` token from a
  # default `gh auth login` does NOT carry admin:ssh_signing_key, so `ssh-key add`
  # fails with a bare 404 that reads like a missing endpoint. Adding the scope is an
  # interactive browser flow, so a script cannot do it and must not pretend to.
  if ! gh api user/ssh_signing_keys >/dev/null 2>&1; then
    warn "the gh token cannot manage signing keys (needs admin:ssh_signing_key)."
    warn "run this yourself — it opens a browser, so no script can do it for you:"
    warn "  gh auth refresh -h github.com -s admin:ssh_signing_key"
    warn "then re-run this script. Local signing is configured; GitHub will show"
    warn "Unverified until the key is registered."
    return 1
  fi

  if gh api user/ssh_signing_keys --jq '.[].key' 2>/dev/null | grep -qF "$key_material_global"; then
    log "key already registered on $(gh api user --jq .login 2>/dev/null) as a signing key"
    return 0
  fi

  log "registering the signing key on GitHub"
  gh ssh-key add "$PUB" --type signing --title "$KEY_TITLE" \
    || { warn "could not register the key"; return 1; }
}

key_material_global="$(awk '{print $1, $2}' "$PUB")"
registered=0
register || registered=1

# --- verify by measurement --------------------------------------------------
# Assert the signature, never the config. Config that reads correctly and produces
# an unverifiable commit is exactly the state this script exists to catch, and it is
# indistinguishable from success until something actually signs.
#
# A throwaway repository, so verification never touches a real branch and never
# needs a push.
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

git -C "$tmp" init -q
git -C "$tmp" config user.name "$(git config --get user.name || echo 'signing check')"
git -C "$tmp" config user.email "$SIGNING_EMAIL"
git -C "$tmp" commit -q --allow-empty -m "signing verification (throwaway)"

status="$(git -C "$tmp" log -1 --format='%G?')"
if [ "$status" != "G" ]; then
  git -C "$tmp" log -1 --format='%G? %GK' >&2 || true
  # These four are distinct diagnoses, not shades of "broken". U in particular is the
  # one that looks like a signing failure and is not: the signature is GOOD, and git
  # merely has no line in allowed_signers tying this email to this key. Collapsing it
  # into a generic error sends you regenerating a key that was never the problem.
  case "$status" in
    N) die "commits are NOT signed (%G? = N) — commit.gpgsign did not take" ;;
    U) die "signature is GOOD but the key is untrusted (%G? = U) — no allowed_signers line maps $SIGNING_EMAIL to this key" ;;
    E) die "signature cannot be checked (%G? = E) — gpg.ssh.allowedSignersFile is unset or unreadable" ;;
    B) die "BAD signature (%G? = B) — the key does not match the signature" ;;
    *) die "unexpected signature status (%G? = $status)" ;;
  esac
fi
log "local: %G? = G, $(git -C "$tmp" log -1 --format='%GK')"

# --- report -----------------------------------------------------------------
if [ "$registered" -ne 0 ]; then
  warn "local signing works; the key is NOT registered on GitHub (see above)."
  warn "commits will sign locally and still show Unverified until it is."
  exit 1
fi

if [ "${SKIP_REGISTER:-}" != "1" ] && command -v gh >/dev/null; then
  account="$(gh api user --jq .login 2>/dev/null || echo '?')"
  log "GitHub: registered on $account as a signing key"
  # The last thing that can still be wrong, and the only one left that is silent:
  # GitHub stamps Verified only when the commit email is a VERIFIED email on the
  # account holding the key. A good signature under an unrecognised email is
  # Unverified with no error anywhere.
  if gh api user/emails >/dev/null 2>&1; then
    if gh api user/emails --jq '.[] | select(.verified) | .email' | grep -qxF "$SIGNING_EMAIL"; then
      log "GitHub: $SIGNING_EMAIL is verified on $account"
    else
      warn "$SIGNING_EMAIL is NOT a verified email on $account — commits will sign"
      warn "cleanly and still show Unverified. Add and verify it in GitHub's email"
      warn "settings, or author as <id>+<account>@users.noreply.github.com, which is"
      warn "always valid for the account that owns it."
      exit 1
    fi
  else
    log "(cannot confirm $SIGNING_EMAIL is verified on $account — needs the 'user' scope)"
    log " gh auth refresh -h github.com -s user"
  fi
fi

log "done. Confirm end to end on the next commit you push:"
log "  gh api repos/<owner>/<repo>/commits/<sha> --jq .commit.verification"
