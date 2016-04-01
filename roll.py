#!/usr/bin/python

import os
import shutil
import subprocess

# Warning: This process is destructive and just gives up on failure.
#
# Prerequisites:
#
# Check out Chromium somewhere on Linux, Mac and Windows.
# Check out git://git.gnome.org/libxslt somewhere.

# Where you have git checkout git://git.gnome.org/libxslt
libxslt_path = 'libxslt_path'

# Where you store Chromium source on Linux, up to and including src
linux_src_path = 'src_path'

# The ref of the Experimental Branch to push intermediate steps to,
# to copy files between platforms.
wip_ref = 'wip_ref'

# Dominic's default setup.
config = {
  libxslt_path: '/work/xml/libxslt',
  linux_src_path: '/work/cb/src',
  wip_ref: 'refs/wip/dominicc/autoroll-libxslt'
}

def sed_in_place(input_filename, program):
  subprocess.check_call(['sed', '-i', program, input_filename])

def roll_libxslt(config):
  third_party_libxslt = 'third_party/libxslt'
  files_to_preserve = ['OWNERS', 'README.chromium', 'BUILD.gn']

  full_path_to_third_party_libxslt = os.path.join(config[linux_src_path],
                                                  third_party_libxslt)

  os.chdir(config[linux_src_path])

  # Nuke the old third_party/libxslt from orbit.
  subprocess.check_call(['git', 'rm', '-rf', third_party_libxslt])
  shutil.rmtree(third_party_libxslt)
  os.mkdir(third_party_libxslt)
  files_to_preserve_with_paths = map(
    lambda s: os.path.join(third_party_libxslt, s),
    files_to_preserve)
  subprocess.check_call(['git', 'reset', '--'] + files_to_preserve_with_paths)
  subprocess.check_call(['git', 'checkout', '--'] +
                        files_to_preserve_with_paths)

  # Update the libxslt repo and export it to the Chromium tree
  os.chdir(config[libxslt_path])
  subprocess.check_call(['git', 'remote', 'update', 'origin'])
  commit = subprocess.check_output(['git', 'log', '-n', '1',
                                    '--pretty=format:%H', 'origin/master'])
  subprocess.check_call(('git archive origin/master | tar -x -C "%s"' %
                         full_path_to_third_party_libxslt),
                        shell=True)
  os.chdir(full_path_to_third_party_libxslt)
  os.remove('.gitignore')

  # Write the commit ID into the README.chromium file
  sed_in_place('README.chromium', 's/Version: .*$/Version: %s/' % commit)

  # Ad-hoc patch for Windows
  sed_in_place('libxslt/security.c',
               r's/GetFileAttributes\b/GetFileAttributesA/g')

  # First run autogen in the root directory to generate configure for
  # use on Mac later.
  subprocess.check_call(['./autogen.sh'])
  subprocess.check_call(['make', 'distclean'])

  os.mkdir('linux')
  os.chdir('linux')
  subprocess.check_call(['../autogen.sh', '--without-debug',
                         '--without-mem-debug', '--without-debugger',
                         '--without-plugins',
                         '--with-libxml-src=../../libxml/linux/'])
  sed_in_place('config.h', 's/#define HAVE_CLOCK_GETTIME 1//')
  # Other platforms share this, even though it is generated on Linux.
  shutil.move('libxslt/xsltconfig.h', '../libxslt')

  # Add *everything* and push it to the cloud for configuring on Mac, Windows
  os.chdir(full_path_to_third_party_libxslt)
  subprocess.check_call(['git', 'add', '*'])
  subprocess.check_call(['git', 'commit', '-m', '%s linux' % commit])
  subprocess.check_call(['git', 'push', '-f', 'wip',
                         'HEAD:%s' % config[wip_ref]])

  print('Now run steps on Windows.')
  # TODO: pull files back from Windows
  # TODO: shutil.rmtree unwanted files from libxslt


def get_out_of_jail(config):
  os.chdir(config[linux_src_path])
  subprocess.check_call(['git', 'reset', '--hard', 'origin/master'])

def go():
  roll_libxslt(config)

def uhoh():
  get_out_of_jail(config)
