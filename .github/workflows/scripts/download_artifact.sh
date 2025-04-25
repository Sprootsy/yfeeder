#!/bin/bash
# The GH download-artifact@v4 does not allow to download an artifact form
# the latest run of a workflow, so I'm using this script.
set -eo pipefail

usage() {
    echo """
    ${0} OWNER REPO ARTIFACT

    Arguments:
        OWNER       Name of the repo owner.
        REPO        Name of the repository.
        ARTIFACT    Name of the artifact to download.

    Env vars:
        GITHUB_TOKEN   Github authentication token.
"""
}

OWNER="$1"
REPO="$2"
ARTIFACT="$3"

if [ -z "$GITHUB_TOKEN" ] || [ -z "$OWNER" ] || [ -z "$REPO" ] || [ -z "$ARTIFACT" ]; then
    usage
    exit 1
fi

GH_API_VERSION_HEADER="X-GitHub-Api-Version: 2022-11-28"
GH_AUTH_HEADER="Authorization: Bearer $GITHUB_TOKEN"
GH_ACCEPT_HEADER="Accept: application/vnd.github+json"

echo "Hello $OWNER! Owner of $REPO"

echo "Searching for artifacts named $ARTIFACT"
ARTIFACT_DATA=$( curl -f -H "$GH_AUTH_HEADER" -H "$GH_ACCEPT_HEADER" -H "$GH_API_VERSION_HEADER" \
    "https://api.github.com/repos/$OWNER/$REPO/actions/artifacts?name=$ARTIFACT" \
    | yq e '[ .artifacts.[] | {"download_url": .archive_download_url, "updated_at": .updated_at } ] | flatten |= sort_by(.updated_at) | .[0] | .updated_at + "," + .download_url' -
    )
if [ -z "$ARTIFACT_DATA" ]; then
    echo "No artifact found!"
    exit 2
fi

UPDATE_TIME=$(echo "$ARTIFACT_DATA" | cut -d ',' -f1)
echo "Found artifact that was updated at $UPDATE_TIME"
DOWNLOAD_URL=$(echo "$ARTIFACT_DATA" | cut -d ',' -f2)
echo "Downloading artifact from $DOWNLOAD_URL"
curl -fL -H "$GH_AUTH_HEADER" -H "$GH_ACCEPT_HEADER" -H "$GH_API_VERSION_HEADER" \
    "$DOWNLOAD_URL" -o "${ARTIFACT}.zip"
chmod 0755 "${ARTIFACT}.zip"
unzip "${ARTIFACT}.zip"
