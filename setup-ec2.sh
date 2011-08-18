#!/bin/bash

set -e
set -o nounset

cd ~

sudo apt-add-repository ppa:martin-james-robinson/webkitgtk

# Pull down Chromium
if [ ! -d chromium ]; then
  mkdir chromium
  pushd chromium >> /dev/null
  svn co http://src.chromium.org/svn/trunk/tools/depot_tools
  git clone http://git.chromium.org/git/chromium.git src
  gclient config http://src.chromium.org/svn/trunk/src

  # Pull down WebKit
  mkdir -p src/third_party
  git clone git://git.webkit.org/WebKit.git src/third_party/WebKit
  pushd src/third_party/WebKit >> /dev/null
  git checkout -b gclient
  popd >> /dev/null  # src/third_party/WebKit

  popd >> /dev/null  # chromium
fi

# Chromium and GTK dependencies
chromium/src/build/install-build-deps.sh

# GTK layout test dependencies
sudo apt-get install curl \
                     ruby \
                     apache2 \
                     libapache2-mod-php5 \
                     libapache2-mod-bw \
                     ttf-liberation \
                     otf-stix \
                     libgstreamer-plugins-base0.10-0 \
                     gstreamer0.10-plugins-base \
                     gstreamer0.10-plugins-good  \
                     gstreamer0.10-plugins-bad \
                     gstreamer0.10-ffmpeg

# Qt dependencies
sudo apt-get install bison flex libqt4-dev libqt4-opengl-dev libphonon-dev libicu-dev libsqlite3-dev libxext-dev libxrender-dev gperf libfontconfig1-dev libphonon-dev g++

if [ ! -d testfonts ]; then
  git clone git://gitorious.org/qtwebkit/testfonts.git
  echo WEBKIT_TESTFONTS=~/testfonts >> ~/.bashrc
fi

# Don't pull in WebKit, etc.
$EDITOR src/DEPS .gclient
