#!/bin/bash
#
set -eo pipefail

usage() {
    echo """
    ${0} OWNER REPO SOURCE TARGET

    Arguments:
        OWNER       Name of the repo owner.
        REPO        Name of the repository.
        SOURCE      Name of the source branch.
        TARGET      Name of the target branch.

    Env vars:
        GITHUB_TOKEN   Github authentication token.
"""
}

OWNER="$1"
REPO="$2"
SOURCE="$3"
TARGET="$4"

if [ -z "$GITHUB_TOKEN" ] || [ -z "$OWNER" ] || [ -z "$REPO" ] || [ -z "$SOURCE" ] || [ -z "$TARGET" ]; then
    usage
    exit 1
fi

GH_API_VERSION_HEADER="X-GitHub-Api-Version: 2022-11-28"
GH_AUTH_HEADER="Authorization: Bearer $GITHUB_TOKEN"
GH_ACCEPT_HEADER="Accept: application/vnd.github+json"

echo "Hello ${OWNER}! Owner of $REPO"

PR_BODY='{"title":"bot: daily news update","head":"'"$SOURCE"'","base":"'"$TARGET"'"}'
echo "Creating PR: $PR_BODY"


MERGE_URL=$( curl -f -X POST -H "$GH_API_VERSION" -H "$GH_AUTH_HEADER" -H "$GH_ACCEPT_HEADER" \
             -d "$PR_BODY" \
             "https://api.github.com/repos/$OWNER/$REPO/pulls" \
             | tee | yq e '.url' - )
if [ -z "$MERGE_URL" ]; then
    echo "Cannot extract PR url from reply"
    exit 2
fi
echo "Merging PR: $PR_URL"
MERGE_BODY='{"title":"bot: daily news update","commit_message": "bot: daily news update", "merge_method": "squash" }'
if curl -f -XPUT -H "$GH_API_VERSION" -H "$GH_AUTH_HEADER" -H "$GH_ACCEPT_HEADER" \
    -d "$MERGE_BODY" \
    "$MERGE_URL/merge" ; then
    echo "PR Merged"
else
    echo "Cannot merge PR!"
    exit 10
fi


