#!/usr/bin/python

import os
import shutil
import stat
import subprocess
import tempfile

# Warning: This process is destructive and just gives up on
# failure. Some amount of examining line numbers on failure is
# expected.
#
# Prerequisites:
#
# - Check out Chromium somewhere on Linux, OS X and Windows.
# - Add the experimental remote named 'wip':
#   git remote add -f wip \
#   https://chromium.googlesource.com/experimental/chromium/src
# - XML
# -- Clone git://git.gnome.org/libxml2
# -- Update config below to explain where everything is.
# - XSLT
# -- Clone git://git.gnome.org/libxslt somewhere.
# -- Update config below to explain where everything is.
# - On OS X, install these MacPorts:
#   autoconf automake libtool pkgconfig icu
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
  src_path_linux: '/usr/local/google/work/ca/src',
  src_path_osx: '/Users/dominicc/ca/src',
  src_path_windows: r'C:\src\ca\src',

  # The ref of the Experimental Branch to push intermediate steps to,
  # to copy files between platforms.
  wip_ref: 'refs/wip/dominicc/autoroll-xml-thing'
}

class WorkingDir(object):
  def __init__(self, path):
    self.prev_path = os.getcwd()
    self.path = path

  def __enter__(self):
    os.chdir(self.path)

  def __exit__(self, exc_type, exc_value, traceback):
    if exc_value:
      print('was in %s; %s before that' % (self.path, self.prev_path))
    os.chdir(self.prev_path)

def git(*args):
  command = ['git'] + list(args)
  # git is a batch file in depot_tools on Windows, so git needs shell=True
  subprocess.check_call(command, shell=(os.name=='nt'))

def sed_in_place(input_filename, program):
  # OS X's sed requires -e
  subprocess.check_call(['sed', '-i', '-e', program, input_filename])

def get_out_of_jail(config, which):
  with WorkingDir(config[which]):
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

def export_to_chromium(remote_repo_path, full_third_party_path):
  with WorkingDir(remote_repo_path):
    git('remote', 'update', 'origin')
    commit = subprocess.check_output(['git', 'log', '-n', '1',
                                      '--pretty=format:%H', 'origin/master'])
    subprocess.check_call(('git archive origin/master | tar -x -C "%s"' %
                           full_third_party_path),
                          shell=True)
  with WorkingDir(full_third_party_path):
    os.remove('.gitignore')
  return commit

def roll_libxslt_linux(config):
  with WorkingDir(config[src_path_linux]):
    files_to_preserve = ['OWNERS', 'README.chromium', 'BUILD.gn', 'libxslt.gyp']
    nuke_preserving(third_party_libxslt, files_to_preserve)

    # Update the libxslt repo and export it to the Chromium tree
    full_path_to_third_party_libxslt = os.path.join(config[src_path_linux],
                                                    third_party_libxslt)
    commit = export_to_chromium(config[libxslt_path],
                                full_path_to_third_party_libxslt)

  with WorkingDir(full_path_to_third_party_libxslt):
    # Write the commit ID into the README.chromium file
    sed_in_place('README.chromium', 's/Version: .*$/Version: %s/' % commit)

    # Ad-hoc patch for Windows
    sed_in_place('libxslt/security.c',
                 r's/GetFileAttributes\b/GetFileAttributesA/g')

    os.mkdir('linux')
    with WorkingDir('linux'):
      subprocess.check_call(['../autogen.sh'] + xslt_configure_options +
                            ['--with-libxml-src=../../libxml/linux/'])
      sed_in_place('config.h', 's/#define HAVE_CLOCK_GETTIME 1//')
      sed_in_place('config.log', 's/[a-z.0-9]+\.corp\.google\.com/REDACTED/')

      # Other platforms share this, even though it is generated on Linux.
      # Android and Windows do not have xlocale.
      sed_in_place('libxslt/xsltconfig.h',
                   '/Locale support/,/#if 1/s/#if 1/#if 0/')
      shutil.move('libxslt/xsltconfig.h', '../libxslt')

    # Back in third_party/libxslt
    # Add *everything* and push it to the cloud for configuring on OS X, Windows
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
  with WorkingDir(full_path_to_third_party_libxslt):
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
    remove_tracked_files(files_to_remove)
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
  with WorkingDir(config[src_path_windows]):
    destructive_fetch_experimental_branch(config)

    # Run the configure script.
    with WorkingDir(os.path.join(third_party_libxslt, 'win32')):
      subprocess.check_call([
        'cscript', '//E:jscript', 'configure.js', 'compiler=msvc', 'iconv=no',
        'xslt_debug=no', 'mem_debug=no', 'debugger=no', 'modules=no'
      ])

    # Add, commit and push the result.
    with WorkingDir(third_party_libxslt):
      shutil.move('config.h', 'win32')
      git('add', 'win32/config.h')
      git('commit', '-m', 'Windows')
      git('push', 'wip', 'HEAD:%s' % config[wip_ref])
      git('clean', '-f')

def roll_libxslt_osx(config):
  with WorkingDir(os.path.join(config[src_path_osx], third_party_libxslt)):
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
  print('cherry picking from %s: %s' % (commit, filename))
  command = (r"""git format-patch -1 --stdout %(commit)s | """
             r"""awk '/---.*\/%(filename)s/,/(--$|diff --git)/ {print}' | """
             r"""patch""" %
             {'commit': commit, 'filename': filename})
  print command
  subprocess.check_call(command, shell=True)

def check_copying(full_path_to_third_party_libxml_src):
  path = os.path.join(full_path_to_third_party_libxml_src, 'COPYING')
  if not os.path.exists(path):
    return
  with open(path) as f:
    s = f.read()
    if 'GNU' in s:
      raise Exception('check COPYING')

def prepare_libxml_distribution(config, temp_dir):
  '''Returns a tuple of commit hash and full path to archive.'''
  # Could push from a distribution prepared upstream instead by
  # returning the version string and a distribution tar file.
  temp_config_path = os.path.join(temp_dir, 'config')
  os.mkdir(temp_config_path)
  temp_src_path = os.path.join(temp_dir, 'src')
  os.mkdir(temp_src_path)

  commit = export_to_chromium(config[libxml_path], temp_src_path)
  with WorkingDir(temp_config_path):
    subprocess.check_call(['../src/autogen.sh'] + xml_configure_options)
    subprocess.check_call(['make', 'dist-all'])

    # Work out what it is called
    tar_file = subprocess.check_output(
        '''awk '/PACKAGE =/ {p=$3} /VERSION =/ {v=$3} '''
        '''END {printf("%s-%s.tar.gz", p, v)}' Makefile''',
        shell=True)
    return commit, os.path.abspath(tar_file)

def roll_libxml_linux(config):
  # Need to snarf this file path before changing dirs
  # TODO(dominicc): This is no longer necessary in Python 3.4.1?
  patch_file = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                            'roll-cr599427.txt')
  print(patch_file)

  full_path_to_third_party_libxml_src = os.path.join(config[src_path_linux],
                                                     third_party_libxml_src)

  with WorkingDir(config[src_path_linux]):
    # Begin pushing from git repo
    try:
      temp_dir = tempfile.mkdtemp()
      print('temporary directory: %s' % temp_dir)

      commit, tar_file = prepare_libxml_distribution(config, temp_dir)

      branch_name = 'roll-libxml-%s' % commit
      # TODO(dominicc): This is messy; at least check for word boundaries
      if branch_name in subprocess.check_output(['git', 'branch']):
        git('checkout', branch_name)
        git('reset', '--hard', 'origin/master')
      else:
        git('checkout', '-b', branch_name, 'origin/master')

      # Nuke the old libxml from orbit; this ensures only desired cruft
      # accumulates
      nuke_preserving(third_party_libxml_src, [])

      # Update the libxml repo and export it to the Chromium tree
      with WorkingDir(third_party_libxml_src):
        subprocess.check_call(
            'tar xzf %s --strip-components=1' % tar_file,
            shell=True)
    finally:
      shutil.rmtree(temp_dir)

    with WorkingDir(third_party_libxml_src):
      # Put the version number is the README file
      sed_in_place('../README.chromium', 's/Version: .*$/Version: %s/' % commit)

      # printf format specifiers
      cherry_pick_patch('d31995076e55f1aac2f935c53b585a90ece27a11', 'timsort.h')
      # crbug.com/599427
      subprocess.check_call(['patch', 'xmlstring.c', patch_file])
      # crbug.com/623378
      for f in ['xpath.c', 'xpointer.c']:
        cherry_pick_patch('b6ad54b72c7f8c422c288dd9c8756d2a15f30e53', f)
      # crbug.com/624011
      cherry_pick_patch('6eee7eee18990d52a5e0723058f0e4d186e1e278',
                        'xpointer.c')

      with WorkingDir('../linux'):
        subprocess.check_call(['../src/autogen.sh'] + xml_configure_options)
        check_copying(full_path_to_third_party_libxml_src)
        sed_in_place('config.h', 's/#define HAVE_RAND_R 1//')

      # Add *everything* and push it to the cloud for configuring on OS
      # X, Windows
      with WorkingDir('../src'):
        git('add', '*')
        git('commit', '-am', '%s libxml, linux' % commit)
        check_copying(full_path_to_third_party_libxml_src)
        git('push', '-f', 'wip', 'HEAD:%s' % config[wip_ref])

  print('Now run steps on Windows, then OS X, then back here.')

def remove_tracked_files(files_to_remove):
  files_to_remove = [f for f in files_to_remove if os.path.exists(f)]
  git('rm', '-rf', *files_to_remove)

# This continues the roll on Linux after Windows and OS X are done.
def roll_libxml_linux_2(config):
  full_path_to_third_party_libxml = os.path.join(config[src_path_linux],
                                                 third_party_libxml_src, '..')
  with WorkingDir(full_path_to_third_party_libxml):
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
      'src/include/libxml/xmlversion.h',
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
      'src/xml2-config.in',
      'src/xmlcatalog.c',
      'src/xmllint.c',
      'src/xstc',
    ]
    remove_tracked_files(files_to_remove)
    git('commit', '-m', 'Remove unused files.')
    commit_message = 'Roll libxml to %s' % commit
    git('cl', 'upload', '-t', commit_message, '-m', commit_message)
    git('cl', 'try')

def roll_libxml_windows(config):
  with WorkingDir(config[src_path_windows]):
    destructive_fetch_experimental_branch(config)
    # Run the configure script.
    with WorkingDir(os.path.join(third_party_libxml_src, 'win32')):
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
  with WorkingDir(os.path.join(config[src_path_osx], third_party_libxml_src,
                               '../mac')):
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
