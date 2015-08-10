; Copyright 2010 Dominic Cooney. All Rights Reserved.

; Writing tools. Use DD-REVISE-PARAGRAPH to start. Diadem uses
; functions in paragraph.el heavily, so consider customizing variables
; like SENTENCE-END-DOUBLE-SPACE and so on.

(require 'cl)

(defgroup dd-faces nil
  "Faces used in Diadem Mode"
  :group 'diadem
  :group 'faces)

(defface dd-primary-face
  '((t (:foreground "Pink")))
  "Face for the text primarily being scrutinized."
  :group 'dd-faces)

(defconst dd-advice-buffer "*DD-Advice*"
  "Buffer to use to display long form prompts.")

(defvar dd-current-session nil
  "State tracking what's currently being revised.")

(defstruct dd-session
  "The target of revisions. It is necessary to keep track of these"
  window
  buffer
  overlay)

(defun dd-new-session ()
  "Starts a new session.
If there is an existing session in DD-CURRENT-SESSION, deletes its overlay."
  (if dd-current-session
      (delete-overlay (dd-session-overlay dd-current-session)))
  (setq dd-current-session (make-dd-session :buffer (current-buffer)
                                            :window (selected-window))))

(defun dd-select-session-for-editing (session)
  "Selects the window and buffer of SESSION."
  (select-window (dd-session-window session))
  (set-buffer (dd-session-buffer session)))

;; Prompts.

(defstruct dd-prompt-button
  "Creates a button labeled TEXT that does ACTION.
Interpreted by DD-PROMPT-PART."
  text
  action)

(defun dd-prompt-paragraph (&rest stuff)
  "Outputs the prompts in STUFF, then formats a paragraph."
  `(paragraph ,stuff))

(defun dd-prompt (&rest parts)
  "Displays a prompt containing PARTS.
DD-PROMPT-PART defines how parts are handled."
  (select-window (or (get-buffer-window dd-advice-buffer)
                     (split-window (selected-window))) t)
  (switch-to-buffer (get-buffer-create dd-advice-buffer))
  (erase-buffer)
  (dd-prompt-part parts)
  (goto-char (point-min)))

(defun dd-prompt-part (part)
  "Interprets PART to generate prompt output."
  (cond ((stringp part) (insert part))
        ((dd-prompt-button-p part)
         (insert-text-button (dd-prompt-button-text part)
                             'action (dd-prompt-button-action part)))
        ((and (consp part) (eq 'paragraph (car part)))
         (dd-prompt-part (cdr part))
         (fill-paragraph))
        ((listp part)
         (mapcar 'dd-prompt-part part))))

;; Things from "Technical Writing and Professional Communication."

(defun dd-twpc-cite (page)
  "Generates a string citation of the Technical Writing book.
PAGE is the page to cite."
  (concat "Olsen, Leslie A., & Huckin, Thomas N. (1991) Technical "
          "Writing and Professional Communication (2nd ed.). (p. "
          (number-to-string page) ") New York, NY: McGraw-Hill."))

(defun dd-twpc-procedure ()
  "Technical Writing's \"procedure for producing more readable texts\"."
  (dd-prompt
     (make-dd-prompt-button :text "Next" :action #'dd-twpc-procedure-step-3)
     "\n\n"
"Read the paragraph. Is there an adequate topic statement clear
pattern of organization? If not, rewrite the topic statement
and/or paragraph as necessary.\n\n"
     (dd-prompt-paragraph (dd-twpc-cite 462)))
  (dd-select-session-for-editing dd-current-session))

(defun dd-twpc-procedure-step-3 (button)
  ; "Consider the first sentence."
  (with-current-buffer (dd-session-buffer dd-current-session)
    (mark-end-of-sentence 1)
    (let ((overlay (make-overlay (point) (mark))))
      (setf (dd-session-overlay dd-current-session) overlay)
      (overlay-put overlay 'face 'dd-primary-face)))
  (dd-twpc-procedure-step-4 nil))

(defun dd-twpc-procedure-step-4 (button)
  (dd-prompt
    (make-dd-prompt-button :text "Next" :action #'dd-twpc-procedure-step-5)
    "\n\n"
"Does the sentence meet the given-new criterion? (Does given
information come before new?)

* Sometimes it will not be possible to rewrite a sentence to meet
all of the desired criteria.

* This may not be relevant for the first sentence of a paragraph.\n\n"
    (dd-prompt-paragraph (dd-twpc-cite 462))))

(defun dd-twpc-procedure-step-5 (button)
  (dd-prompt
    (make-dd-prompt-button :text "Rewritten!"
                           :action #'dd-twpc-procedure-step-4)
    " "
    (make-dd-prompt-button :text "Next"
                           :action #'dd-twpc-procedure-step-6)
    "\n\n"
"Does the sentence put topical information (the subject of the
paragraph) in subject position (the subject of the sentence)?

1. Identify noun phrases in the sentence.

2. If the sentence has a passive verb, identify the hidden actor
   and consider that noun phrase too.

3. Is one of the noun phrases the subject of the paragraph? Make
   it the subject of the sentence.

4. Look at the subjects of other sentences. Are they similar? If
   not it may indicate the subject of the paragraph needs to be
   revised.

* Sometimes it will not be possible to rewrite a sentence to meet
all of the desired criteria.

* This may not be relevant for the first sentence of a paragraph.\n\n"
    (dd-prompt-paragraph (dd-twpc-cite 462))))

(defun dd-twpc-procedure-step-6 (button)
  (dd-prompt
    (make-dd-prompt-button :text "Rewritten!"
                           :action #'dd-twpc-procedure-step-4)
    " "
    (make-dd-prompt-button :text "Next"
                           :action #'dd-twpc-procedure-step-7)
    "\n\n"
"Does the sentence put light noun phrases before heavy ones?

Noun phrases vary in length, complexity and precision. Light noun
phrases are short and simple; heavy noun phrases are long and
complex. The preferred ordering is light noun phrases before
heavy noun phrases. Putting light noun phrases first packs more
of the sentence structure at the front of the sentence.

Ideally the entire structure of the sentence should fit within
about nine words. If that is not possible, at least the subject
and the verb should fit within nine words. Introductory units
don't count to this nine word limit but they should also follow
the light-before-heavy rule.

1. Identify noun phrases and count the number of words in each.

2. Restructure the sentence to put shorter phrases first. If
   necessary you can use \"it\" to shift a heavy subject noun
   phrase to the end of the sentence.

3. Count nine words into each unit. Does this include at least
   the subject and verb?

* Put a heavy noun phrase first, in violation of this rule, to
  convey topical information or emphasis.

In summary (S=Subject, V=Verb, _=Noun Phrase):

S ___ V ___________.  Preferred.
S ______ V ________.  Acceptable if the subject is not too long.
                      \"Balanced.\"
S ___________ V ___.  Avoid, except for special effects.

* Sometimes it will not be possible to rewrite a sentence to meet
all of the desired criteria.\n\n"
    (dd-prompt-paragraph (dd-twpc-cite 462))))

(defun dd-twpc-procedure-step-7 (button)
  ; Go to the next sentence.
  ; This assumes the author left point within the sentence being edited.
  (with-current-buffer (dd-session-buffer dd-current-session)
    (let ((overlay (dd-session-overlay dd-current-session)))
      (setf (point) (overlay-end overlay))
      (forward-char)
      (if (eq (point) (point-max))
          ; TODO: Do TWPC Section 26.1 here to combine sentences.
          (dd-done button)
        (progn
          (mark-end-of-sentence 1)
          (move-overlay overlay (point) (mark))
          (deactivate-mark)
          (dd-twpc-procedure-step-4 nil))))))

(defun dd-done (button)
  "Finishes an editing session."
  (dd-prompt "Thanks for revising.")
  (dd-select-session-for-editing dd-current-session)
  (delete-overlay (dd-session-overlay dd-current-session))
  (widen)
  (setq dd-current-session nil))

(defun dd-revise-paragraph ()
  "Interactively revises the current paragraph.
This resets the restriction."
  (interactive)
  (dd-new-session)
  (mark-paragraph)
  (narrow-to-region (point) (mark))
  (deactivate-mark)
  (dd-twpc-procedure))

(defvar dd-mode
  nil
  "The diadem mode variable.")
(make-variable-buffer-local 'dd-mode)

(defvar dd-after-change-timer
  nil
  "The timer responding to quiescence after editing.")
(make-variable-buffer-local 'dd-after-change-timer)

(defvar dd-after-change-timer-delay
  0.2
  "The time, in seconds, to delay before reparsing the buffer.")

(defun dd-after-change ()
  (message "Somewhere, a dog barked."))

(defun dd-after-change-immediate (begin end old-length)
  (dd-reschedule-timer)
  (message "dd: %d %d %d" begin end old-length))

(defun dd-reschedule-timer ()
  (if dd-after-change-timer
      (cancel-timer dd-after-change-timer))
  (setq
   dd-after-change-timer
   (run-with-idle-timer dd-after-change-timer-delay nil #'dd-after-change)))

(defun dd-mode (&optional arg)
  "Toggles Diadem writing support tools minor mode.
With ARG, turn Diadem on if ARG is positive."
  (interactive "P")
  (let ((old-dd-mode dd-mode))
    (setq dd-mode
          (if (null arg)
              (not dd-mode)
            (> (prefix-numeric-value arg) 0)))
    (if (not (equal dd-mode old-dd-mode))
        (progn
          (if dd-mode
              (add-hook 'after-change-functions
                        #'dd-after-change-immediate nil t)
            (remove-hook 'after-change-functions #'dd-after-change-immedate t))
          (message "Diadem %s." (if dd-mode "activated" "deactivated")))
      (message "Diadem already %s." (if dd-mode "active" "inactive")))))

(provide 'diadem)
