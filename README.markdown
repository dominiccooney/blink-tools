These are a few things that make life easier for [Blink
development](http://www.chromium.org/blink)

 * [analyze.R](analyze.R) has R scripts for
   determining the likelihood that page cycler results are real
   regressions.

 * [blink-stuff.el](blink-stuff.el) has
   some Gnu Emacs hooks and functions for browsing and editing Blink
   source.

 * [completions](completions) has bash
   tab completion for layout tests names.

 * [init-chromium-env.sh](init-chromium-env.sh)
   sets up some git options for Chromium development.

 * [a11y](a11y) has an Android accessibility agent that dumps
   accessibility events, useful for debugging the accessibility
   integration of web platform stuff.

 * [serving](serving) has some Go scripts that do reverse proxies and
   set no-caching headers, useful for debugging Service Workers.

The following are of historical interest only:

 * [idl-analyze.el](idl-analyze.el) is elisp
   for parsing WebKit IDL files. It has probably completely rotted.
