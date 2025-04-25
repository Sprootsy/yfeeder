#!/bin/bash
# The GH download-artifact@v4 does not allow to download an artifact form
# the latest run of a workflow, so I'm using this script.

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

if [ -z "$GITHUB_TOKEN" ] || [ -z "$OWNER" ] || [ -z "$REPO" ] || [ -z "$WORKfLOW_ID"]; then
    usage()
    exit 1
fi

GH_API_VERSION_HEADER="X-GitHub-Api-Version: 2022-11-28"
GH_AUTH_HEADER="Authorization: Bearer $GITHUB_TOKEN"
GH_ACCEPT_HEADER="Accept: application/vnd.github+json"
GH_HEADERS="-H $GH_API_VERSION_HEADER -H $GH_AUTH_HEADER -H $GH_ACCEPT_HEADER"
CURLCMD="curl -sL $GH_HEADERS"

echo "Hello $OWNER! Owner of $REPO"
echo "Searching for id of $WORKFLOW_NAME"

echo "Searching for artifacts in $RUN_ID"
ARTIFACT_DATA=$( $CURLCMD "https://api.github.com/repos/$OWNER/$REPO/actions/artifacts?name=$ARTIFACT" \
    | yq e '[ .artifacts.[] | {"download_url": .archive_download_url, "updated_at": .updated_at } ] | flatten |= sort_by(.updated_at) | .[0] | .updated_at + "," + .download_url' )
if [ -z "$ARTIFACT_DATA" ]; then
    echo "No artifact found!"
    exit 2
fi

UPDATE_TIME=$(echo "$ARTIFACT_DATA" | cut -d ',' -f2)
echo "Found artifact that was updated at $UPDATE_TIME"
DOWNLOAD_URL=$(echo "$ARTIFACT_DATA" | cut -d ',' -f1)
echo "Downloading artifact from $DOWNLOAD_URL"
$CURLCMD "$DOWNLOAD_URL" -o "${ARTIFACT}.zip"
unzip "${ARTIFACT}.zip"
