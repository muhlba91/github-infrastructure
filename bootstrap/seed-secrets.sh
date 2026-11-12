#!/usr/bin/env bash
# Seeds this repository's bootstrap secrets into Google Secret Manager (see BOOTSTRAP.md).
#
# usage: bootstrap/seed-secrets.sh           create missing secrets and prompt for their values
#        bootstrap/seed-secrets.sh --check   list missing secrets (names only), exit 1 if any
#
# Values are read without echo and passed to gcloud on stdin; they are never printed or written to disk.
# Existing secrets are only given a new version after confirmation.
set -euo pipefail

# gcloud must never wait for an answer: its prompts are hidden by the redirections below
export CLOUDSDK_CORE_DISABLE_PROMPTS=1

PROJECT="public-services-402810"
PREFIX="github-infrastructure--"
REPOSITORY="github-infrastructure"
# keep in sync with the `provider = "gsm"` secrets in fnox.toml
SECRETS=(
  GITHUB_TOKEN
  GITLAB_TOKEN
  OAUTH_CLIENT_ID
  OAUTH_CLIENT_SECRET
  SCW_ACCESS_KEY
  SCW_SECRET_KEY
  AWS_ACCESS_KEY_ID
  AWS_SECRET_ACCESS_KEY
)

check_only=false
case "${1:-}" in
  --check) check_only=true ;;
  "") ;;
  *) echo "usage: $0 [--check]" >&2; exit 2 ;;
esac

exists() {
  gcloud secrets describe "$1" --project "$PROJECT" --format='value(name)' >/dev/null 2>&1
}

has_version() {
  [[ -n "$(gcloud secrets versions list "$1" --project "$PROJECT" --filter='state=ENABLED' --limit=1 --format='value(name)' 2>/dev/null)" ]]
}

missing=0
for name in "${SECRETS[@]}"; do
  id="${PREFIX}${name}"

  if $check_only; then
    if ! exists "$id" || ! has_version "$id"; then
      echo "missing: $id"
      missing=1
    fi
    continue
  fi

  if exists "$id"; then
    if has_version "$id"; then
      read -r -p "$id exists with a value. Add a new version? [y/N] " answer
      [[ "$answer" =~ ^[Yy]$ ]] || continue
    fi
  else
    gcloud secrets create "$id" --project "$PROJECT" --replication-policy=automatic \
      --labels="repository=${REPOSITORY},purpose=bootstrap" >/dev/null
    echo "created: $id"
  fi

  read -r -s -p "value for $id: " value
  echo
  if [[ -z "$value" ]]; then
    echo "empty value, skipped: $id" >&2
    continue
  fi
  printf '%s' "$value" | gcloud secrets versions add "$id" --project "$PROJECT" --data-file=- >/dev/null
  unset value
  echo "stored: $id"
done

if $check_only; then
  [[ $missing -eq 0 ]] && echo "all ${#SECRETS[@]} secrets present"
  exit $missing
fi
