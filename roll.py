#!/usr/bin/python

import os
import shutil
import subprocess

# Warning: This process is destructive and just gives up on failure.
#
# Prerequisites:
#
# - Check out Chromium somewhere on Linux, OS X and Windows.
# - Add the experimental remote named 'wip'.
# - Check out git://git.gnome.org/libxslt somewhere.

third_party_libxslt = 'third_party/libxslt'

libxslt_path = 'libxslt_path'
src_path_linux = 'src_path_linux'
src_path_windows = 'src_path_windows'
wip_ref = 'wip_ref'

# Dominic's default setup.
config = {
  # Where you have git checkout git://git.gnome.org/libxslt
  libxslt_path: '/work/xml/libxslt',

  # Where you store Chromium source, up to and including src
  src_path_linux: '/work/cb/src',
  src_path_windows: r'C:\src\ca\src',

  # The ref of the Experimental Branch to push intermediate steps to,
  # to copy files between platforms.
  wip_ref: 'refs/wip/dominicc/autoroll-libxslt'
}

def git(*args):
  command = ['git'] + list(args)
  # git is a batch file in depot_tools on Windows, so git needs shell=True
  subprocess.check_call(command, shell=(os.name=='nt'))

def sed_in_place(input_filename, program):
  subprocess.check_call(['sed', '-i', program, input_filename])

def roll_libxslt_linux(config):
  files_to_preserve = ['OWNERS', 'README.chromium', 'BUILD.gn', 'libxslt.gyp']
  os.chdir(config[src_path_linux])

  # Nuke the old third_party/libxslt from orbit.
  git('rm', '-rf', third_party_libxslt)
  shutil.rmtree(third_party_libxslt)
  os.mkdir(third_party_libxslt)
  files_to_preserve_with_paths = map(
    lambda s: os.path.join(third_party_libxslt, s),
    files_to_preserve)
  git('reset', '--', *files_to_preserve_with_paths)
  git('checkout', '--', *files_to_preserve_with_paths)

  # Update the libxslt repo and export it to the Chromium tree
  os.chdir(config[libxslt_path])
  git('remote', 'update', 'origin')
  commit = subprocess.check_output(['git', 'log', '-n', '1',
                                    '--pretty=format:%H', 'origin/master'])
  full_path_to_third_party_libxslt = os.path.join(config[src_path_linux],
                                                  third_party_libxslt)
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
  # use on OS X later.
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

  # Add *everything* and push it to the cloud for configuring on OS X, Windows
  os.chdir(full_path_to_third_party_libxslt)
  git('add', '*')
  git('commit', '-m', '%s linux' % commit)
  git('push', '-f', 'wip', 'HEAD:%s' % config[wip_ref])

  print('Now run steps on Windows, then OS X.')
  # TODO: Consider hanging here and watch the repository and resume
  # the process automatically.

# TODO: Implement roll_libxslt_osx

# This continues the roll on Linux after Windows and OS X are done.
def roll_libxslt_linux_2(config):
  full_path_to_third_party_libxslt = os.path.join(config[src_path_linux],
                                                  third_party_libxslt)
  os.chdir(full_path_to_third_party_libxslt)
  git('pull', 'wip', config[wip_ref])
  commit = subprocess.check_output(['awk', '/Version:/ {print $2}',
                                    'README.chromium'])
  files_to_remove = [
    # TODO: Excluding ChangeLog and NEWS because encoding problems mean
    # bots can't patch these. Reinclude them when there is a consistent
    # encoding.
    'NEWS',
    'ChangeLog',
    # These have shebang but not executable bit; presubmit will barf on them.
    'autogen.sh',
    'linux/config.status',
    'linux/libtool',
    'linux/xslt-config',
    'xslt-config.in',
    # These are not needed.
    'autom4te.cache',
    'doc',
    'python',
    'tests',
    'xsltproc',
    'linux/doc',
    'linux/python',
    'linux/tests',
    'linux/xsltproc',
    'linux/libexslt/.deps',
    'linux/libxslt/.deps',
    'examples',
    'vms'
  ]
  git('rm', '-rf', *files_to_remove)
  git('commit', '-m', 'Remove unused files.')
  commit_message = 'Roll libxslt to %s' % commit
  git('cl', 'upload', '-t', commit_message, '-m', commit_message)
  git('cl', 'try')

def roll_libxslt_windows(config):
  # Fetch the in-progress roll from the experimental branch.
  os.chdir(config[src_path_windows])
  git('fetch', 'wip', config[wip_ref])
  git('reset', '--hard', 'FETCH_HEAD')

  # Run the configure script.
  os.chdir(os.path.join(third_party_libxslt, 'win32'))
  subprocess.check_call([
    'cscript', '//E:jscript', 'configure.js', 'compiler=msvc', 'iconv=no',
    'xslt_debug=no', 'mem_debug=no', 'debugger=no', 'modules=no'
  ])

  # Add, commit and push the result.
  os.chdir(os.path.join(config[src_path_windows], third_party_libxslt))
  shutil.move('config.h', 'win32')
  git('add', 'win32/config.h')
  git('commit', '-m', 'Windows')
  git('push', 'wip', 'HEAD:%s' % config[wip_ref])
  git('clean', '-f')

def get_out_of_jail(config, which):
  os.chdir(config[which])
  git('reset', '--hard', 'origin/master')
  git('clean', '-f')

def lgo():
  roll_libxslt_linux(config)

def lgo2():
  roll_libxslt_linux_2(config)

def luhoh():
  get_out_of_jail(config, src_path_linux)

def wgo():
  roll_libxslt_windows(config)

def wuhoh():
  get_out_of_jail(config, src_path_windows)
