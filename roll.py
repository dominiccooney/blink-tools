#!/usr/bin/python

import os
import shutil
import stat
import subprocess

# Warning: This process is destructive and just gives up on
# failure. Some amount of examining line numbers on failure is
# expected.
#
# Prerequisites:
#
# - Check out Chromium somewhere on Linux, OS X and Windows.
# - Add the experimental remote named 'wip'.
# - XML
# -- Clone git://git.gnome.org/libxml2
# -- Update config below to explain where everything is.
# - XSLT
# -- Clone git://git.gnome.org/libxslt somewhere.
# -- Update config below to explain where everything is.
# - On OS X, install these MacPorts:
#   autoconf automake libtool pkgconfig
# - On Linux:
#   sudo apt-get install libicu-dev
#   TODO(dominicc): Investigate making this use Chrome's ICU
#
# Procedure:
#
# For X = xml, xslt
# 1. On Linux, run roll.X_lgo()
# 2. On Windows, run roll.X_wgo()
# 3. On OS X, run roll.X_ogo()
# 4. On Linux, run roll.X_lgo2()
# 5. Check the try jobs pass (git cl try-results).
# 6. Update the README.chromium file with notes.
# 7. Complete code review, commit queue, etc. as normal.
#
# Troubleshooting:
#
# - Examine individual git commits to identify which step caused problems.
# - Run roll.[lwo]uhoh() to reset your local git repo and try again.

third_party_libxml_src = 'third_party/libxml/src'
third_party_libxslt = 'third_party/libxslt'

xml_configure_options = ['--without-iconv', '--with-icu', '--without-ftp',
                         '--without-http', '--without-lzma']
xslt_configure_options = ['--without-debug', '--without-mem-debug',
                          '--without-debugger', '--without-plugins']

libxml_path = 'libxml_path'
libxslt_path = 'libxslt_path'
src_path_linux = 'src_path_linux'
src_path_osx = 'src_path_osx'
src_path_windows = 'src_path_windows'
wip_ref = 'wip_ref'

# Dominic's default setup.
config = {
  # Where you have git checkout git://git.gnome.org/libxslt
  libxml_path: '/usr/local/google/work/xml/libxml2',

  # Where you have git checkout git://git.gnome.org/libxslt
  libxslt_path: '/usr/local/google/work/xml/libxslt',

  # Where you store Chromium source, up to and including src
  src_path_linux: '/usr/local/google/work/cb/src',
  src_path_osx: '/Users/dpc/ca/src',
  src_path_windows: r'C:\src\ca\src',

  # The ref of the Experimental Branch to push intermediate steps to,
  # to copy files between platforms.
  wip_ref: 'refs/wip/dominicc/autoroll-xml-thing'
}

def git(*args):
  command = ['git'] + list(args)
  # git is a batch file in depot_tools on Windows, so git needs shell=True
  subprocess.check_call(command, shell=(os.name=='nt'))

def sed_in_place(input_filename, program):
  # OS X's sed requires -e
  subprocess.check_call(['sed', '-i', '-e', program, input_filename])

def get_out_of_jail(config, which):
  os.chdir(config[which])
  git('reset', '--hard', 'origin/master')
  git('clean', '-f')

def nuke_preserving(third_party_path, files_to_preserve):
  git('rm', '-rf', third_party_path)
  shutil.rmtree(third_party_path, ignore_errors=True)
  os.mkdir(third_party_path)
  files_to_preserve_with_paths = map(
    lambda s: os.path.join(third_party_path, s),
    files_to_preserve)
  git('reset', '--', *files_to_preserve_with_paths)
  git('checkout', '--', *files_to_preserve_with_paths)

def export_to_chromium_chdir(remote_repo_path, full_third_party_path):
  os.chdir(remote_repo_path)
  git('remote', 'update', 'origin')
  commit = subprocess.check_output(['git', 'log', '-n', '1',
                                    '--pretty=format:%H', 'origin/master'])
  subprocess.check_call(('git archive origin/master | tar -x -C "%s"' %
                         full_third_party_path),
                        shell=True)
  os.chdir(full_third_party_path)
  os.remove('.gitignore')
  return commit

def roll_libxslt_linux(config):
  os.chdir(config[src_path_linux])

  files_to_preserve = ['OWNERS', 'README.chromium', 'BUILD.gn', 'libxslt.gyp']
  nuke_preserving(third_party_libxslt, files_to_preserve)

  # Update the libxslt repo and export it to the Chromium tree
  full_path_to_third_party_libxslt = os.path.join(config[src_path_linux],
                                                  third_party_libxslt)
  commit = export_to_chromium_chdir(config[libxslt_path],
                                    full_path_to_third_party_libxslt)

  # Write the commit ID into the README.chromium file
  sed_in_place('README.chromium', 's/Version: .*$/Version: %s/' % commit)

  # Ad-hoc patch for Windows
  sed_in_place('libxslt/security.c',
               r's/GetFileAttributes\b/GetFileAttributesA/g')

  # First run autogen in the root directory to generate configure for
  # use on OS X later.
#  subprocess.check_call(['./autogen.sh', '--help'])

  os.mkdir('linux')
  os.chdir('linux')
  subprocess.check_call(['../autogen.sh'] + xslt_configure_options +
                        ['--with-libxml-src=../../libxml/linux/'])
  sed_in_place('config.h', 's/#define HAVE_CLOCK_GETTIME 1//')

  # Other platforms share this, even though it is generated on Linux.
  # Android and Windows do not have xlocale.
  sed_in_place('libxslt/xsltconfig.h',
               '/Locale support/,/#if 1/s/#if 1/#if 0/')
  shutil.move('libxslt/xsltconfig.h', '../libxslt')

  # Add *everything* and push it to the cloud for configuring on OS X, Windows
  os.chdir(full_path_to_third_party_libxslt)
  git('add', '*')
  git('commit', '-m', '%s libxslt, linux' % commit)
  git('push', '-f', 'wip', 'HEAD:%s' % config[wip_ref])

  print('Now run steps on Windows, then OS X, then back here.')
  # TODO: Consider hanging here and watch the repository and resume
  # the process automatically.

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

def destructive_fetch_experimental_branch(config):
  # Fetch the in-progress roll from the experimental branch.
  git('fetch', 'wip', config[wip_ref])
  git('reset', '--hard', 'FETCH_HEAD')
  git('clean', '-f')

def roll_libxslt_windows(config):
  os.chdir(config[src_path_windows])
  destructive_fetch_experimental_branch(config)
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

def roll_libxslt_osx(config):
  os.chdir(os.path.join(config[src_path_osx], third_party_libxslt))
  destructive_fetch_experimental_branch(config)
  # Run the configure script
  subprocess.check_call(['autoreconf', '-i'])
  os.chmod('configure', os.stat('configure').st_mode | stat.S_IXUSR)
  # /linux here is not a typo; configure looks here to find xml2-config
  subprocess.check_call(['./configure'] + xslt_configure_options +
                        ['--with-libxml-src=../libxml/linux/'])
  # Clean up and emplace the file
  sed_in_place('config.h', 's/#define HAVE_CLOCK_GETTIME 1//')
  os.mkdir('mac')
  shutil.move('config.h', 'mac')
  # Commit and upload the result
  git('add', 'mac/config.h')
  git('commit', '-m', 'OS X')
  git('push', 'wip', 'HEAD:%s' % config[wip_ref])
  git('clean', '-f')

# Does something like cherry-pick, but against edited files.
def cherry_pick_patch(commit, filename):
  command = (r"""git format-patch -1 --stdout %(commit)s | """
             r"""awk '/---.*\/%(filename)s/,/--$/ {print}' | patch""" %
             {'commit': commit, 'filename': filename})
  subprocess.check_call(command, shell=True)

def roll_libxml_linux(config):
  # Need to snarf this file path before changing dirs
  # TODO(dominicc): This is no longer necessary in Python 3.4.1?
  patch_file = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                            'roll-cr599427.txt')

  os.chdir(config[src_path_linux])

  # Nuke the old libxml from orbit; this ensures only desired cruft accumulates
  nuke_preserving(third_party_libxml_src, [])
  # Update the libxml repo and export it to the Chromium tree
  full_path_to_third_party_libxml_src = os.path.join(config[src_path_linux],
                                                     third_party_libxml_src)
  commit = export_to_chromium_chdir(config[libxml_path],
                                    full_path_to_third_party_libxml_src)
  # Put the version number is the README file
  sed_in_place('../README.chromium', 's/Version: .*$/Version: %s/' % commit)

  # printf format specifiers
  cherry_pick_patch('d31995076e55f1aac2f935c53b585a90ece27a11', 'timsort.h')
  # crbug.com/595262
  cherry_pick_patch('27a27edb43e42023db7d154dcd9f66a500296cc1', 'parser.c')
  # crbug.com/599427
  subprocess.check_call(['patch', 'xmlstring.c', patch_file])
  # crbug.com/602280
  cherry_pick_patch('bc5dfe3dbb61e497438849dbe909520128f5bbac', 'uri.c')

  os.chdir('../linux')
  subprocess.check_call(['../src/autogen.sh'] + xml_configure_options)
  sed_in_place('config.h', 's/#define HAVE_RAND_R 1//')

  # Add *everything* and push it to the cloud for configuring on OS X, Windows
  os.chdir('../src')
  git('add', '*')
  git('commit', '-am', '%s libxml, linux' % commit)
  git('push', '-f', 'wip', 'HEAD:%s' % config[wip_ref])

  print('Now run steps on Windows, then OS X, then back here.')

# This continues the roll on Linux after Windows and OS X are done.
def roll_libxml_linux_2(config):
  full_path_to_third_party_libxml = os.path.join(config[src_path_linux],
                                                 third_party_libxml_src, '..')
  os.chdir(full_path_to_third_party_libxml)
  git('pull', 'wip', config[wip_ref])
  commit = subprocess.check_output(['awk', '/Version:/ {print $2}',
                                    'README.chromium'])
  files_to_remove = [
    'src/HACKING',
    'src/INSTALL.libxml2',
    'src/MAINTAINERS',
    'src/Makefile.win',
    'src/README.cvs-commits',
    'src/VxWorks',
    'src/autogen.sh',
    'src/autom4te.cache',
    'src/build_glob.py',
    'src/chvalid.def',
    'src/doc',
    'src/example',
    'src/genChRanges.py',
    'src/global.data',
    'src/include/libxml/xmlwin32version.h',
    'src/include/libxml/xmlwin32version.h.in',
    'src/libxml2.doap',
    'src/macos/libxml2.mcp.xml.sit.hqx',
    'src/optim',
    'src/os400',
    'src/python',
    'src/result',
    'src/rngparser.c',
    'src/test',
    'src/testOOM.c',
    'src/testOOMlib.c',
    'src/testOOMlib.h',
    'src/vms',
    'src/win32/VC10/config.h',
    'src/win32/wince',
    'src/xmlcatalog.c',
    'src/xmllint.c',
    'src/xstc',
  ]
  git('rm', '-rf', *files_to_remove)
  git('commit', '-m', 'Remove unused files.')
  commit_message = 'Roll libxml to %s' % commit
  git('cl', 'upload', '-t', commit_message, '-m', commit_message)
  git('cl', 'try')

def roll_libxml_windows(config):
  os.chdir(config[src_path_windows])
  destructive_fetch_experimental_branch(config)
  # Run the configure script.
  os.chdir(os.path.join(third_party_libxml_src, 'win32'))
  subprocess.check_call([
    'cscript', '//E:jscript', 'configure.js', 'compiler=msvc', 'iconv=no',
    'icu=yes', 'ftp=no', 'http=no'
  ])

  # Add, commit and push the result.
  shutil.move('VC10/config.h', '../../win32/config.h')
  git('add', '../../win32/config.h')
  shutil.move('../include/libxml/xmlversion.h', '../../win32/xmlversion.h')
  git('add', '../../win32/xmlversion.h')
  git('commit', '-m', 'Windows')
  git('push', 'wip', 'HEAD:%s' % config[wip_ref])
  git('clean', '-f')

def roll_libxml_osx(config):
  os.chdir(os.path.join(config[src_path_osx], third_party_libxml_src,
                        '../mac'))
  destructive_fetch_experimental_branch(config)
  subprocess.check_call(['autoreconf', '-i', '../src'])
  subprocess.check_call(['../src/configure'] + xml_configure_options)
  sed_in_place('config.h', 's/#define HAVE_RAND_R 1//')
  git('commit', '-am', 'libxml, mac')
  git('push', 'wip', 'HEAD:%s' % config[wip_ref])

def luhoh():
  get_out_of_jail(config, src_path_linux)

def xml_lgo():
  roll_libxml_linux(config)

def xml_lgo2():
  roll_libxml_linux_2(config)

def xslt_lgo():
  roll_libxslt_linux(config)

def xslt_lgo2():
  roll_libxslt_linux_2(config)

def ouhoh():
  get_out_of_jail(config, src_path_osx)

def xml_ogo():
  roll_libxml_osx(config)

def xslt_ogo():
  roll_libxslt_osx(config)

def wuhoh():
  get_out_of_jail(config, src_path_windows)

def xml_wgo():
  roll_libxml_windows(config)

def xslt_wgo():
  roll_libxslt_windows(config)
