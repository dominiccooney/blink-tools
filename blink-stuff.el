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
  "Sets up various hooks and fonts for Blink development."
  (interactive)

  (add-hook 'c++-mode-hook 'wk-prog-mode-hook)
  (add-hook 'change-log-mode-hook 'wk-change-log-mode-hook)

  (setq auto-mode-alist
        (append (list '("\\.mm\\'" . objc-mode)
                      '("\\.h\\'" . c++-mode))
                auto-mode-alist))

  (defface wk-tab-face
    '((t (:background "Red"))) "Tab" :group 'font-lock-faces)

  (defface wk-generated-file-face
    '((t (:foreground "Black"
          :background "color-149"))) "Generated file" :group 'font-lock-faces)

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
        (set-variable 'indent-tabs-mode nil))))

(defun wk-characterize-path (file-name)
  "Characterizes a file path as V8, or co for core."
  (catch 'return
    (let ((patterns '(("/blink/bindings/" . "V8")
                      ("/bindings/v8/" . "V8")
                      ("/WebKit/chromium/" . "V8")
                      ("/Source/core/" . "co"))))
      (dolist (pattern patterns)
        (when (string-match (car pattern) file-name)
          (throw 'return (cdr pattern)))))
    "??"))

(defun wk-is-generated-file (file-name)
  "Returns non-NIL if the specified file is generated."
  (catch 'return
    (let ((patterns '(("/blink/bindings/" . 't)
                      ("/bindings/v8/custom/" . nil)
                      ("/WebKit/chromium/" . nil)
                      ("/Source/core/" . nil))))
      (let ((case-fold-search nil))
        (dolist (pattern patterns)
          (when (string-match (car pattern) file-name)
            (throw 'return (cdr pattern))))))
    nil))


(defun wk-binding-alternatives (root base-name)
  "Gets file paths for a binding BASE-NAME in Blink tree at ROOT."
  (let ((patterns (list
                   ;; core .h, .cpp and .idl
                   (concat root "/Source/core/*/" base-name ".*")
                   (concat root "/Source/core/*/*/" base-name ".*")
                   ;; V8 generated bindings, Chromium, ninja build
                   (concat root "/../../out/*/gen/blink/bindings/V8"
                           base-name ".cpp")
                   (concat root "/../../out/*/gen/blink/bindings/V8"
                           base-name ".h")
                   ;; V8 custom code
                   (concat root "/Source/core/bindings/v8/custom/V8Custom"
                           base-name ".*")
                   (concat root "/Source/core/bindings/v8/custom/V8"
                           base-name "Custom.*")))
        (result))
    (dolist (pattern patterns)
      (message "Considering %s" pattern)
      (setq result (append (file-expand-wildcards pattern) result)))
    (sort result #'string<)))

(defconst wk-file-list-buffer "*Blink files*"
  "Buffer to use for display lists of Blink files.")

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
    (setq files
          (sort files (lambda (p1 p2)
                        (or (string< (wk-characterize-path p1)
                                     (wk-characterize-path p2))
                            (string< (substring p1 (length root))
                                     (substring p2 (length root)))))))
    (dolist (file files)
      (let ((short-file-path (substring file (length root)))
            (start (point)))
        (insert (wk-characterize-path file))
        (if (wk-is-generated-file file)
            (add-text-properties
             start (point)
             '(font-lock-face wk-generated-file-face)))
        (insert " ")
        (insert-text-button short-file-path
                            'action #'wk-display-file-list-action
                            'button file)
        (insert "\n")))
    (goto-char (point-min))
    (forward-char 3)  ;; skip past the characterization to the path
    (font-lock-mode 't)
    ))

(defun wk-display-file-list-action (button)
  "Visits a file when it is activated in the file list."
  (let ((path (get-text-property button 'button)))
    (delete-window)
    (find-file path)))

(defun wk-root-base-name-of-file (file)
  "Gets the base name of FILE or nil if plausibly a file with a JS binding.
Returns a pair of `(ROOT . BASE-NAME)' where ROOT is the Blink folder."
  (catch 'return
    (let ((patterns '(
            ;; Chromium, ninja build
            ("\\(.*\\)/out/\\(Debug\\|Release\\)/blink/bindings/V8\\(.*\\)\\.cpp$"
             . ((root . (1 "/third_party/WebKit")) (base-name . 3)))
            ("\\(.*\\)/out/\\(Debug\\|Release\\)/blink/bindings/V8\\(.*\\)\\.h$"
             . ((root . (1 "/third_party/WebKit")) (base-name . 3)))
            ;; V8 custom bindings
            ("\\(.*\\)/Source/core/bindings/v8/custom/V8Custom\\(.*\\)\\..*"
             . ((root . 1) (base-name . 2)))
            ("\\(.*\\)/Source/core/bindings/v8/custom/V8\\(.*\\)Custom\\..*"
             . ((root . 1) (base-name . 2)))
            ;; WebCore types and IDLs
            ("\\(.*\\)/Source/core/\\([a-z/]*\\)/\\([A-Z].*\\)\\.\\(.*\\)$"
             . ((root . 1) (base-name . 3))))))
      (let* ((mk-path
              (lambda (pat)
                (cond
                 ((listp pat)
                  (apply #'concat (mapcar mk-path pat)))
                 ((numberp pat)
                  (match-string pat file))
                 ((stringp pat)
                  pat)
                 (t (error "invalid pattern: %s" pat))))))
        (dolist (pattern patterns)
          (when (string-match (car pattern) file)
            (throw 'return (cons (funcall mk-path (cdr (assoc 'root (cdr pattern))))
                                 (funcall mk-path (cdr (assoc 'base-name (cdr pattern))))))))))))

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
          (if (not (file-exists-p file-name))
              (kill-buffer buffer)
            (progn
              (if (buffer-modified-p)
                  (switch-to-buffer buffer))
              (revert-buffer t (not (buffer-modified-p)) t)
              (goto-char saved-point))))))))

;; (setq edebug-all-defs nil)

(provide 'webkit-stuff)

;; TODO: characterize results as [D]ebug or [R]elease
;; TODO: color the results buffer instead of using ugly underlined buttons
;; TODO: q should close the results buffer