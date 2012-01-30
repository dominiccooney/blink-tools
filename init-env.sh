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

if [ -z "$(grep Chromium ~/.subversion/config)" ]; then
  mkdir -p ~/.subversion
  echo '# Chromium auto-props
[miscellany]
enable-auto-props = yes

[auto-props]
*.jpg = svn:mime-type=image/jpeg
*.pdf = svn:mime-type=application/pdf
*.png = svn:mime-type=image/png
*.webp = svn:mime-type=image/webp' >> ~/.subversion/config
else
  echo 'Not updating ~/.subversion/config properties'
fi

git config merge.changelog.driver "perl $(pwd)/Tools/Scripts/resolve-ChangeLogs --merge-driver %O %A %B"
git config core.editor "perl $(pwd)/Tools/Scripts/commit-log-editor --regenerate-log"

git config user.name "${REAL_NAME}"
git config user.email "${EMAIL}"
git config core.autocrlf false
git config core.filemode false
git config bugzilla.username "${EMAIL}"
git config color.status auto
git config color.diff auto
git config color.branch auto
