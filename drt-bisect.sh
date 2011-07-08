#!/bin/bash

# Does a clean release build of WebKit and runs DumpRenderTree,
# checking the output to see if the revision is good or not.
#
# If the build fails, the revision is skipped.
#
# If the build is OK, drt-bisect scans the output for FAIL; if it
# finds FAIL it marks the revision as "bad". If there is no FAIL and
# the output contains PASS then it marks the revision as
# "good". Finally, if the output contains neither PASS nor FAIL then
# it marks the revision as bad.
#
# Usage:
#
# git bisect start known_bad known_good
# git bisect run ~/webkit-tools/drt-bisect.sh /full/path/to/test1 ...

# TODO: make it work on Linux, Windows; other ports
# TODO: make it find build output intelligently

tests=$*

# Must specify some tests
if [ -z "$tests" ]; then
  echo "usage: $0 /full/path/to/test1 /full/path/to/test2 ..."
  exit 128
fi

# Check the tests exist up-front, otherwise git bisect wastes a lot of
# time building for nothing.
for test in $tests
do
  if [[ ! -a "$test" ]]; then
    echo "not found: $test"
    exit 128
  fi
done

# The location of build-webkit depends on the era of the revision
script_dir=
script_dir_candidates="Tools/Scripts WebKitTools/Scripts"
for script_dir_candidate in $script_dir_candidates
do
  if [ -a "${script_dir_candidate}/build-webkit" ]; then
    script_dir=$script_dir_candidate
  fi
done

if [ -z "${script_dir}" ]; then
  echo 'could not find build-webkit and friends'
  echo '(are you running from root of WebKit tree?)'
  exit 128
fi

# Guess where built products go
if [[ "$(pwd)" =~ '/Volumes/' ]]; then
  build_dir=WebKitBuild
else
  build_dir=~/bin
fi

${script_dir}/build-webkit --release --clean || exit 125
rm -rf "${build_dir}/*" || exit 125
${script_dir}/build-webkit --release || exit 125
${script_dir}/build-dumprendertree --release || exit 125

for test in $tests
do
  out=$("${build_dir}/Release/DumpRenderTree" "$test" 2>&1)
  echo $out
  if [[ "$out" =~ "FAIL" || ! "$out" =~ "PASS" ]]; then
    echo 'bad'
    exit 1
  fi
done

echo 'good'
exit 0
