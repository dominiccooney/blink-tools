These are a few things that make life easier for [Blink
development](http://www.chromium.org/blink)

 * [analyze.R](webkit-tools/blob/master/analyze.R) has R scripts for
   determining the likelihood that page cycler results are real
   regressions.

 * [blink-stuff.el](webkit-tools/blob/master/blink-stuff.el) has
   some Gnu Emacs hooks and functions for browsing and editing Blink
   source.

 * [completions](webkit-tools/blob/master/completions) has bash
   tab completion for layout tests names.

 * [init-chromium-env.sh](webkit-tools/blob/master/init-chromium-env.sh)
   sets up some git options for Chromium development.

 * [html5rocks](webkit-tools/blob/master/html5rocks) has a Chrome
   extension for turning off [html5rocks.com's](http://html5rocks.com)
   obnoxious dotty backgrounds.

The following are of historical interest only:

 * [idl-analyze.el](webkit-tools/blob/master/idl-analyze.el) is elisp
   for parsing WebKit IDL files. It has probably completely rotted.

 * [workflow](webkit-tools/blob/master/workflow) has some Go code for
   downloading data from Buganizer and computing some statistics about
   patch acceptance rates.
