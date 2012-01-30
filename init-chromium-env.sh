#!/bin/bash

# This configures a WebKit git enlistment with WebKit's ChangeLog
# preparing and merging tools and user metadata.

REAL_NAME='Dominic Cooney'
EMAIL='dominicc@chromium.org'

if [ ! -f DEPS ]; then
  echo 'run in a Chromium enlistment'
  exit 1
fi

if [ ! -d .git ]; then
  echo 'this is designed for doing Chromium development with git'
  echo 'see <http://code.google.com/p/chromium/wiki/UsingNewGit>'
  exit 1
fi

git config user.name "${REAL_NAME}"
git config user.email "${EMAIL}"
git config core.autocrlf false
git config core.filemode false
git config color.status auto
git config color.diff auto
git config color.branch auto
