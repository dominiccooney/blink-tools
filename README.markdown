These are a few things that make life easier for [WebKit
development](http://www.webkit.org/coding/contributing.html):

 * [webkit-stuff.el](webkit-stuff/blob/master/webkit-stuff.el) has
   some Gnu Emacs hooks and functions for browsing and editing WebKit
   source.
 * [wk-gdb](webkit-stuff/blob/master/wk-gdb) is a gdb script with
   functions for spelunking data structures. I would rather [script
   gdb in
   Python,](http://sourceware.org/gdb/current/onlinedocs/gdb/Python.html)
   but WebKit development uses toolchains with ancient versions of gdb
   that don't support Python scripting.
 * [init-env.sh](webkit-stuff/blob/master/init-env.sh) configures git
   with WebKit's ChangeLog creating and merging tools. Edit the
   `REAL_NAME` and `EMAIL` variables in the script before running it.