#!/usr/bin/python

import os
import subprocess

third_party_libxslt = 'third_party/libxslt'
files_to_preserve = ['OWNERS', 'README.chromium']

def get_out_of_jail(src_path):
  os.chdir(src_path)
  subprocess.check_call(['git', 'reset', '--hard', 'HEAD'])

def roll_libxslt(src_path):
  os.chdir(src_path)
  files_to_preserve_with_paths = map(
    lambda s: os.path.join(third_party_libxslt, s),
    files_to_preserve)
  subprocess.check_call(['git', 'rm', '-rf', third_party_libxslt])
  subprocess.check_call(['rm', '-rf', third_party_libxslt])
  os.mkdir(third_party_libxslt)
  subprocess.check_call(['git', 'reset', '--'] + files_to_preserve_with_paths)
  subprocess.check_call(['git', 'checkout', '--'] +
                        files_to_preserve_with_paths)
