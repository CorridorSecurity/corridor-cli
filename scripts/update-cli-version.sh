#!/usr/bin/env bash
set -euo pipefail

CLI_VERSION_FILE="cli/VERSION"
FRONTEND_VERSION_FILE="frontend/public/cli/VERSION"

# ── 1. Pull latest main ─────────────────────────────────────────────
echo "Fetching and checking out latest main…"
git fetch origin main
git checkout main
git pull origin main

# ── 2. Determine the new version ────────────────────────────────────
if [ -f "$CLI_VERSION_FILE" ]; then
  CURRENT_VERSION=$(cat "$CLI_VERSION_FILE" | tr -d '[:space:]')
else
  CURRENT_VERSION="0.0.0"
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"
PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "Current version: ${CURRENT_VERSION}"
echo "New version:     ${NEW_VERSION}"

# ── 3. Create a release branch and bump version files ────────────────
BRANCH="cli-${NEW_VERSION}"
echo "Creating branch ${BRANCH}…"
git checkout -b "$BRANCH"

mkdir -p "$(dirname "$CLI_VERSION_FILE")"
mkdir -p "$(dirname "$FRONTEND_VERSION_FILE")"

echo "$NEW_VERSION" > "$CLI_VERSION_FILE"
echo "$NEW_VERSION" > "$FRONTEND_VERSION_FILE"

git add "$CLI_VERSION_FILE" "$FRONTEND_VERSION_FILE"
git commit -m "Bump CLI version to ${NEW_VERSION}"
git push -u origin "$BRANCH"

# ── 4. Open a draft PR ──────────────────────────────────────────────
PR_TITLE="Bump CLI version to ${NEW_VERSION}"
PR_BODY="Bump CLI version from ${CURRENT_VERSION} to ${NEW_VERSION}.

### Changes
- \`cli/VERSION\` → \`${NEW_VERSION}\`
- \`frontend/public/cli/VERSION\` → \`${NEW_VERSION}\`"

PR_URL=$(gh pr create \
  --title "$PR_TITLE" \
  --body "$PR_BODY" \
  --base main \
  --head "$BRANCH" \
  --draft)

echo ""
echo "✅  PR created: ${PR_URL}"

# ── 5. Print next steps ─────────────────────────────────────────────
cat <<EOF

──────────────────────────────────────────────────────
Next steps
──────────────────────────────────────────────────────

1. The **CLI Installer Test / test_cli_install** check will initially fail
   because it tests against the latest published release, which doesn't
   have version ${NEW_VERSION} yet.

2. Go to **Actions → Deploy CLI** in GitHub and click **Run workflow**
   on the \`${BRANCH}\` branch. The workflow will automatically:
     • Build hooks binaries and sign macOS hooks with Apple Developer certificate
     • Build CLI binaries for all platforms (linux/darwin/windows, amd64/arm64)
     • Sign and notarize macOS CLI binaries via goreleaser's built-in quill support
     • Publish a GitHub release to CorridorSecurity/corridor-cli with all
       archives and checksums

3. After the deploy completes, re-run the **CLI Installer Test /
   test_cli_install** check on the PR — it should now pass.

4. Merge the PR to \`main\`. Once frontend is deployed to prod, the
   latest version of the CLI will be automatically installed via the
   install.sh script.

──────────────────────────────────────────────────────
EOF
