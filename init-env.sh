#!/bin/bash

# This configures a WebKit git enlistment with WebKit's ChangeLog
# preparing and merging tools and user metadata.

REAL_NAME='Dominic Cooney'
EMAIL='dominicc@chromium.org'

if [ ! -f Tools/Scripts/resolve-ChangeLogs ]; then
  echo 'run in a WebKit enlistment'
  exit 1
fi

if [ ! -d .git ]; then
  echo 'this is designed for doing WebKit development with git'
  echo 'see <https://trac.webkit.org/wiki/UsingGitWithWebKit>'
  exit 1
fi

git config merge.changelog.driver "perl $(pwd)/Tools/Scripts/resolve-ChangeLogs --merge-driver %O %A %B"
git config core.editor "perl $(pwd)/Tools/Scripts/commit-log-editor --regenerate-log"

git config user.name "${REAL_NAME}"
git config user.email "${EMAIL}"
git config bugzilla.username "${EMAIL}"
git config color.status auto
git config color.diff auto
git config color.branch auto
