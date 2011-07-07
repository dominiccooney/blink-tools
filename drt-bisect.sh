#!/bin/bash

# Does a clean release build of WebKit and runs DumpRenderTree,
# grepping the output for PASS. If PASS is output, then the revision
# is "good." If not, the revision is "bad". If the build fails or
# DumpRenderTree fails to run, the revision is skipped.
#
# Usage:
#
# git bisect start known_bad known_good
# EXPORT TEST=/full/path/to/test
# git bisect run ~/webkit-tools/drt-bisect.sh

if [[ -z "$TEST" || ! -a "$TEST" ]]; then
  echo '$TEST should point to a file to run in DRT'
  exit 128
fi

Tools/Scripts/build-webkit --release --clean || exit 125
Tools/Scripts/build-webkit --release || exit 125
Tools/Scripts/build-dumprendertree --release || exit 125
out=$(~/bin/Release/DumpRenderTree $TEST 2>&1)
echo $out
if [[ "$out" =~ "PASS" ]]; then
  echo 'good'
  exit 0
fi

echo 'bad'
exit 1
