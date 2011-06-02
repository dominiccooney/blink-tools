;; Copyright 2011 Dominic Cooney. All Rights Reserved.
;;
;; This has two useful functions for working with WebKit in emacs:
;;
;; wk-find-binding-files finds files related to JavaScript bindings
;; for the current buffer; for example, when visiting Node.h if you
;; M-x wk-find-binding-files it will search your source tree and
;; present you with a menu including Node.cpp, Node.idl, JSC and V8
;; custom bindings and JSC and V8 generated bindings (if they exist in
;; your build output.)
;;
;; wk-refresh-files is for getting your open buffers to a good state
;; after a git rebase. It looks at all of your open buffers and, if
;; they haven't been edited, updates them to the latest on-disk
;; versions; if they have been edited wk-refresh-files will prompt you
;; to revert them or not (or you can use C-g to bail out at this
;; unhappy juncture to save them and use git diff to clean up.)
;;
;; This also has some mode hooks and font tweaks for highlighting
;; tabs. To install the hooks, run:
;;
;;    (wk-setup)

(defun wk-setup ()
  "Sets up various hooks and fonts for WebKit development."
  (interactive)

  (add-hook 'c++-mode-hook 'wk-prog-mode-hook)
  (add-hook 'change-log-mode-hook 'wk-change-log-mode-hook)

  (setq auto-mode-alist
        (append (list '("\\.mm\\'" . objc-mode)
                      '("\\.h\\'" . c++-mode))
                auto-mode-alist))

  (defface wk-tab-face
    '((t (:background "Red"))) "Tab" :group 'font-lock-faces)
  (font-lock-add-keywords 'c++-mode
                          '(("\\(\t+\\)" 1 'wk-tab-face)))
  (font-lock-add-keywords 'change-log-mode
                          '(("\\(\t+\\)" 1 'wk-tab-face))))

(defun wk-prog-mode-hook ()
  (if (and (string-match "WebKit" (buffer-file-name)))
      (progn
       (set-variable 'indent-tabs-mode nil)
       (set-variable 'c-basic-offset 4))))

(defun wk-change-log-mode-hook ()
  (if (string-match "WebKit" (buffer-file-name))
      (let ()
        (set-variable 'indent-tabs-mode nil)
        (flyspell-mode))))

(defun wk-characterize-path (file-name)
  "Characterizes a file path as JS for JSC, V8, or WC for WebCore."
  (catch 'return
    (let ((patterns '(("/Source/WebCore/bindings/js/" . "JS")
                      ("/WebKitBuild/\\(Debug\\|Release\\)/DerivedSources/WebCore/" . "JS")
                      ("/Source/WebCore/bindings/v8/" . "V8")
                      ("/Source/WebKit/chromium/" . "V8")
                      ("/Source/WebCore/" . "WC"))))
      (dolist (pattern patterns)
        (when (string-match (car pattern) file-name)
          (throw 'return (cdr pattern)))))
    "??"))

(defun wk-binding-alternatives (root base-name)
  "Gets file paths for a binding BASE-NAME in WebKit tree at ROOT."
  (let ((patterns (list
                   ;; WebCore .h, .cpp and .idl
                   (concat root "/Source/WebCore/*/" base-name ".*")
                   (concat root "/Source/WebCore/*/*/" base-name ".*")
                   ;; JSC generated bindings
                   (concat root "/WebKitBuild/*/DerivedSources/WebCore/JS"
                           base-name ".*")
                   ;; JSC custom code
                   (concat root "/Source/WebCore/bindings/js/JSCustom"
                           base-name ".*")
                   (concat root "/Source/WebCore/bindings/js/JS" base-name
                           "Custom.*")
                   ;; V8 generated bindings
                   (concat root "/Source/WebKit/chromium/xcodebuild"
                           "/DerivedSources/*/webcore/bindings/V8"
                           base-name ".cpp")
                   (concat root "/Source/WebKit/chromium/xcodebuild"
                           "/DerivedSources/*/webkit/bindings/V8"
                           base-name ".h")
                   ;; V8 custom code
                   (concat root "/Source/WebCore/bindings/v8/custom/V8Custom"
                           base-name ".*")
                   (concat root "/Source/WebCore/bindings/v8/custom/V8"
                           base-name "Custom.*")))
        (result))
    (dolist (pattern patterns)
      (message "Considering %s" pattern)
      (setq result (append (file-expand-wildcards pattern) result)))
    (sort result #'string<)))

(defconst wk-file-list-buffer "*WebKit files*"
  "Buffer to use for display lists of WebKit files.")

(defun wk-display-file-list (root files)
  "Display a list of FILES shortening the names by trimming off ROOT."
  (catch 'return
    (unless files
      (message "no files to display")
      (throw 'return nil))
    (select-window (or (get-buffer-window wk-file-list-buffer)
                       (split-window (selected-window))))
    (shrink-window (- (window-height) (+ 2 (length files))))
    (switch-to-buffer (get-buffer-create wk-file-list-buffer))
    (erase-buffer)
    (dolist (file files)
      (let ((short-file-path (substring file (length root))))
        (insert (wk-characterize-path file) " ")
        (insert-text-button short-file-path
                            'action #'wk-display-file-list-action
                            'button file)
        (insert "\n")))
    (goto-char (point-min))
    (forward-char 3)  ;; skip past the characterization to the path
    ))

(defun wk-display-file-list-action (button)
  "Visits a file when it is activated in the file list."
  (let ((path (get-text-property button 'button)))
    (delete-window)
    (find-file path)))

(defun wk-root-base-name-of-file (file)
  "Gets the base name of FILE or nil if plausibly a file with a JS binding.
Returns a pair of `(ROOT . BASE-NAME)' where ROOT is the WebKit folder."
  (catch 'return
    (let ((patterns '(
            ;; JSC derived sources
            ("\\(.*\\)/WebKitBuild/\\(Debug\\|Release\\)/DerivedSources/WebCore/JS\\(.*\\)\\..*$"
             . ((root . 1) (base-name . 3)))
            ;; Chromium derived sources
            ("\\(.*\\)/Source/WebKit/chromium/xcodebuild/DerivedSources/\\(Debug\\|Release\\)/webcore/bindings/V8\\(.*\\)\\.cpp$"
             . ((root . 1) (base-name . 3)))
            ("\\(.*\\)/Source/WebKit/chromium/xcodebuild/DerivedSources/\\(Debug\\|Release\\)/webkit/bindings/V8\\(.*\\)\\.h$"
             . ((root . 1) (base-name . 3)))
            ;; JSC custom bindings
            ("\\(.*\\)/Source/WebCore/bindings/js/JSCustom\\(.*\\)\\..*"
             . ((root . 1) (base-name . 2)))
            ("\\(.*\\)/Source/WebCore/bindings/js/JS\\(.*\\)Custom\\..*"
             . ((root . 1) (base-name . 2)))
            ;; V8 custom bindings
            ("\\(.*\\)/Source/WebCore/bindings/v8/custom/V8Custom\\(.*\\)\\..*"
             . ((root . 1) (base-name . 2)))
            ("\\(.*\\)/Source/WebCore/bindings/v8/custom/V8\\(.*\\)Custom\\..*"
             . ((root . 1) (base-name . 2)))
            ;; WebCore types and IDLs
            ("\\(.*\\)/Source/WebCore/\\([a-z/]*\\)/\\([A-Z].*\\)\\.\\(.*\\)$"
             . ((root . 1) (base-name . 3))))))
      (dolist (pattern patterns)
        (when (string-match (car pattern) file)
          (throw 'return (cons (match-string (cdr (assoc 'root (cdr pattern))) file)
                               (match-string (cdr (assoc 'base-name (cdr pattern))) file))))))))

(defun wk-find-binding-files ()
  "Finds JavaScript binding files related to the current buffer."
  (interactive)
  (let ((root-base-name (wk-root-base-name-of-file (buffer-file-name))))
    (if root-base-name
        (wk-display-file-list (car root-base-name)
                              (wk-binding-alternatives (car root-base-name) (cdr root-base-name)))
      (message "JavaScript binding files not found"))))

(defun wk-refresh-files ()
  "Refreshes buffers, useful after a git rebase."
  (interactive)
  (save-excursion
    (dolist (buffer (buffer-list))
      (set-buffer buffer)
      (let ((saved-point (point))
            (file-name (buffer-file-name)))
        (when (not (verify-visited-file-modtime (current-buffer)))
          (if (buffer-modified-p)
              (switch-to-buffer buffer))
          (revert-buffer t (not (buffer-modified-p)) t)
          (goto-char saved-point))))))

;; (setq edebug-all-defs nil)

(provide 'webkit-stuff)

;; TODO: don't hard-code WebKitBuildDir
;; TODO: don't assume WebKit build dir is under source tree root
;; TODO: characterize results as [D]ebug or [R]elease
;; TODO: color the results buffer instead of using ugly underlined buttons
;; TODO: q should close the results buffer
