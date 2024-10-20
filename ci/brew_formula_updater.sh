#!/bin/bash
# Copyright 2024 IBM Corp
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

# Validate argument count
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <base_repo_url> <new_version> <brew_tap_repo_url> <formula_file>"
    exit 1
fi

# Ensure GitHub token and GH CLI token are set
if [[ -z "$GITHUB_TOKEN" || -z "$GH_TOKEN" ]]; then
    echo "Error: Required tokens are not set."
    exit 1
fi

BASE_REPO_URL="$1"
NEW_VERSION="$2"
BREW_TAP_REPO_URL="$3"
FORMULA_FILE="$4"
FORMULA_PATH="Formula/"
RELEASES_PATH="$BASE_REPO_URL/releases/download"

echo "BASE_REPO_URL: $BASE_REPO_URL"
echo "NEW_VERSION: $NEW_VERSION"
echo "BREW_TAP_REPO_URL: $BREW_TAP_REPO_URL"
echo "FORMULA_FILE: $FORMULA_FILE"
echo "RELEASES path in base repo - "$RELEASES_PATH

URL_VALUES=(
    "$RELEASES_PATH/v$NEW_VERSION/pvsadm-darwin-amd64.tar.gz"
    "$RELEASES_PATH/v$NEW_VERSION/pvsadm-darwin-arm64.tar.gz"
    "$RELEASES_PATH/v$NEW_VERSION/pvsadm-linux-amd64.tar.gz"
)

# Compute SHA256 checksum
compute_sha256() {
    local url="$1"
    local temp_file="$(mktemp)"
    http_status=$(curl -L -w "%{http_code}" -o "$temp_file" "$url")
    
    # Check HTTP status 200
    if [ "$http_status" -ne 200 ]; then
        echo "Error: Failed to download $url (HTTP status: $http_status)"
        rm "$temp_file"
        exit 1
    fi

    # Verify the content is not HTML
    if grep -iq "<!doctype\|<html\|<head\|<body" <(head -n 10 "$temp_file"); then
        echo "Error: Downloaded content is not a valid tar. Possibly an HTML error page."
        rm "$temp_file"
        exit 1
    fi

    local sha256
    sha256=$(shasum -a 256 "$temp_file" | awk '{print $1}')
    rm "$temp_file"
    # Uncomment for testing purposes echo "SHA256 for $url: $sha256"
    echo "$sha256"
}

# Compute SHA256 checksums for the new version
SHAs=()
for url in "${URL_VALUES[@]}"; do
    SHA=$(compute_sha256 "$url")
    if [ -z "$SHA" ]; then
        echo "Error: SHA256 could not be computed for $url"
        exit 1
    fi
    SHAs+=("$SHA")
    echo $SHA
    sleep 1
done

# Clone the repository having brew formula and checkout a new branch for pushing updates
git clone "$BREW_TAP_REPO_URL" brew_tap_repo_temp
if [ $? -ne 0 ]; then
    echo "Error: failed to clone the repository"
    exit 1
fi

cd brew_tap_repo_temp || { echo "Error: Failed to navigate to brew_tap_repo_temp"; exit 1; }
BRANCH_NAME="bump_formula_v$NEW_VERSION"
git checkout -b "$BRANCH_NAME" || { echo "Error: Failed to create branch $BRANCH_NAME"; exit 1; }
cd "$FORMULA_PATH" || { echo "Error: Failed to navigate to $FORMULA_PATH"; exit 1; }

# Update the version field in formula file
if [ ! -f "$FORMULA_FILE" ]; then
    echo "Error: Formula file $FORMULA_FILE not found"
    exit 1
fi
sed -i.bak "s/version \".*\"/version \"$NEW_VERSION\"/" "$FORMULA_FILE"

# Update the individual URL and the corresponding sha hash in formula file
echo "updating individual URL and SHA to version $NEW_VERSION."
for index in "${!SHAs[@]}"; do
    SHA=${SHAs[$index]}
    echo "$SHA"  

    URL=${URL_VALUES[$index]}

    if [[ $URL == *"/pvsadm-linux-amd64.tar.gz" ]]; then
        perl -i -pe 's|('${RELEASES_PATH}'/v)(\d+\.\d+\.\d+)(/pvsadm-linux-amd64\.tar\.gz)|${1}'${NEW_VERSION}'${3}|g' $FORMULA_FILE
        perl -i -0pe 's|(url "'${RELEASES_PATH}'/v[0-9]+\.[0-9]+\.[0-9]+/pvsadm-linux-amd64\.tar\.gz"\s*sha256 ")[^"]*(")|${1}'${SHA}'${2}|g' $FORMULA_FILE

    elif [[ $URL == *"/pvsadm-darwin-arm64.tar.gz" ]]; then
        perl -i -pe 's|('${RELEASES_PATH}'/v)(\d+\.\d+\.\d+)(/pvsadm-darwin-arm64\.tar\.gz)|${1}'${NEW_VERSION}'${3}|g' $FORMULA_FILE
        perl -i -0pe 's|(url "'${RELEASES_PATH}'/v[0-9]+\.[0-9]+\.[0-9]+/pvsadm-darwin-arm64\.tar\.gz"\s*sha256 ")[^"]*(")|${1}'${SHA}'${2}|g' $FORMULA_FILE

    elif [[ $URL == *"/pvsadm-darwin-amd64.tar.gz" ]]; then
        perl -i -pe 's|('${RELEASES_PATH}'/v)(\d+\.\d+\.\d+)(/pvsadm-darwin-amd64\.tar\.gz)|${1}'${NEW_VERSION}'${3}|g' $FORMULA_FILE
        perl -i -0pe 's|(url "'${RELEASES_PATH}'/v[0-9]+\.[0-9]+\.[0-9]+/pvsadm-darwin-amd64\.tar\.gz"\s*sha256 ")[^"]*(")|${1}'${SHA}'${2}|g' $FORMULA_FILE

    else
        echo "Warning: Unrecognized URL pattern for SHA: $SHA"
    fi
done

# Check if branch already exists in formula repo
if git ls-remote --exit-code --heads origin "$BRANCH_NAME"; then
    echo "Error: The branch '$BRANCH_NAME' already exists on the remote. Please create a new branch."
    exit 1
fi

# Commit and push updated formula file to remote tap repository
git add "$FORMULA_FILE"
git commit -m "Update formula to version $NEW_VERSION"
git remote set-url origin https://x-access-token:${GITHUB_TOKEN}@${BREW_TAP_REPO_URL#https://}

if ! git push origin "$BRANCH_NAME"; then
    echo "Error: Failed to push branch $BRANCH_NAME"
    exit 1
fi

# Need to unset GITHUB_TOKEN for gh cli to use GH_TOKEN instead
unset GITHUB_TOKEN

# Create PR
if gh pr create --head "$BRANCH_NAME" \
    --title "Updates formula to version $NEW_VERSION" \
    --body "New Release version $NEW_VERSION has been created in $BASE_REPO_URL. Bumping formula version to version $NEW_VERSION."; then
    echo "Updated formula to version $NEW_VERSION, pushed changes to $BRANCH_NAME, and created a PR"
else
    echo "Error: failed to create PR"
    exit 1 
fi
