;;; Analyzes Web IDL files

(require 'guruguru)

;; Paths

(defvar default-idl-root
  "~/es-operating-system-read-only/esidl/dom")

;; For parsing WebIDL

(setq max-lisp-eval-depth 60000)
(setq max-specpdl-size 120000)

;;;; General utilities.

(defun find-files (glob)
  "Finds files that match GLOB relative to the current directory."
  (split-string (shell-command-to-string (format "find . -name '%s'" glob))))

(defmacro in-dir (dir &rest body)
  "Switches the working directory to DIR and evaluates BODY."
  `(let ((old-dir default-directory))
     (cd ,dir)
     (let ((result (progn ,@body)))
       (cd old-dir)
       result)))

(defmacro process-file (file &rest body)
  "Pulls the contents of FILE into a temporary buffer and runs BODY."
  `(with-temp-buffer
     (insert-file-contents ,file nil nil nil t)
     (goto-char (point-min))
     (let ((result (progn ,@body)))
       (kill-buffer)
       result)))

(defun find-idl-files (&optional root)
  "Finds the *.idl files under ROOT."
  (setq root (or root default-idl-root))
  (in-dir root (find-files "*.idl")))

;;; WebIDL parsing

; This design is copied from js2-mode's visitor implementation.
(defun wk-visit-ast (node callback)
  "Visit every node in NODE with visitor CALLBACK.

CALLBACK is a function that takes two arguments: (NODE END-P). It
is called twice, once to visit the node, and again after all the
node's children have been processed. The END-P argument is nil on
the first call and non-nil on the second call. The first call to
CALLBACK can opt not to visit the node's children by returning
nil. CALLBACK always gets the second call with non-nil END-P,
however."
  (cond
   ((null node)
    ;; no node to visit
    nil)
   ((listp node)
    ;; this is a list of nodes, visit them in order
    (dolist (n node) (wk-visit-ast n callback)))
   (t
    (let* ((node-kind (aref node 0))
           (vfunc (get node-kind 'wk-visitor)))
      ;; start the visit
      (when (funcall callback node nil)
        (if vfunc
            ;; visit child nodes
            (funcall vfunc node callback)
          (error "%s does not define visitor traversal function" node-kind)))
      ;; end the visit
      (funcall callback node t)))))

(defstruct (wk-node
            (:constructor nil))  ; abstract
  "Base WebIDL node type"
  (line (line-number-at-pos gg-start))
  (pos gg-start)
  (end gg-end))


(defstruct (wk-ns-name
            (:include wk-node))
  "A namespaced name, like dom::Node."
  module-name
  local-name
  )


(defstruct (wk-ty-optional
            (:include wk-node))
  "An optional type, like DOMString?"
  underlying-type)


(defstruct (wk-ty-array
            (:include wk-node))
  "An array type, like DOMString[]."
  underlying-type)


(defstruct (wk-module
            (:include wk-node))
  "WebIDL module declaration."
  name
  declarations
  )

(put 'cl-struct-wk-module 'wk-visitor 'wk-visit-module)

(defun wk-visit-module (node callback)
  "Visit the declarations of a `wk-module'."
  (wk-visit-ast (wk-module-declarations node) callback))


(defstruct (wk-interface
            (:include wk-node))
  "WebIDL interface declaration."
  name
  extends
  extended-attributes
  members
  )

(put 'cl-struct-wk-interface 'wk-visitor 'wk-visit-interface)

(defun wk-visit-interface (node callback)
  "Visit the extended attributes and members of a `wk-interface'."
  (wk-visit-ast (wk-interface-extended-attributes node) callback)
  (wk-visit-ast (wk-interface-members node) callback))


(defstruct (wk-typedef
            (:include wk-node))
  "WebIDL typedef declaration."
  type
  new-name
  )

(put 'cl-struct-wk-typedef 'wk-visitor 'wk-visit-typedef)

(defun wk-visit-typedef (node callback)
  nil)


(defstruct (wk-implements
            (:include wk-node))
  "WebIDL implements declaration."
  implementer
  implemented)

(put 'cl-struct-wk-implements 'wk-visitor 'wk-visit-implements)

(defun wk-visit-implements (node callback)
  nil)


(defstruct (wk-extended-attribute 
            (:include wk-node))
  "WebIDL extended attribute with an optional value or formal parameter list.

Extended attributes can be attached to interfaces, attributes,
methods and formal parameters.

Formal parameter lists are used with the Constructor attribute."
  name
  value
  formals)

(put 'cl-struct-wk-extended-attribute 'wk-visitor 'wk-visit-extended-attribute)

(defun wk-visit-extended-attribute (node callback)
  nil)


(defstruct (wk-member-attribute
            (:include wk-node))
  "WebIDL attribute member."
  spec-attributes
  extended-attributes
  type
  name
  getter-raises
  setter-raises
  )

(put 'cl-struct-wk-member-attribute 'wk-visitor 'wk-visit-member-attribute)

(defun wk-visit-member-attribute (node callback)
  "Visit the extended attributes of a `wk-member-attribute'."
  (wk-visit-ast (wk-member-attribute-extended-attributes node) callback))


(defstruct (wk-member-method
            (:include wk-node))
  "WebIDL method member."
  spec-attributes
  extended-attributes
  type
  name
  formals
  raises
  )

(put 'cl-struct-wk-member-method 'wk-visitor 'wk-visit-member-method)

(defun wk-visit-member-method (node callback)
  "Visits the extended attributes of a `wk-member-method'."
  (wk-visit-ast (wk-member-method-extended-attributes node) callback))


(defstruct (wk-formal
            (:include wk-node))
  "A formal parameter of a method."
  direction   ; 'in
  optional    ; 'required or 'optional
  variadicity ; 'once or 'variadic
  extended-attributes
  type
  name
  )

(put 'cl-struct-wk-formal 'wk-visitor 'wk-visit-formal)

(defun wk-visit-formal (node callback)
  "Visits the extended attributes of a `wk-formal'."
  (wk-visit-ast (wk-formal-extended-attributes node) callback))


(defstruct (wk-member-field
            (:include wk-node))
  "WebIDL constant field member."
  type
  name
  value ; string--unparsed, this may be an decimal or 0x hex literal
  )

(put 'cl-struct-wk-member-field 'wk-visitor 'wk-visit-member-field)

(defun wk-visit-member-field (node callback)
  ;; for whatever reason, fields don't seem to use extended attributes
  nil)


(setq webidl-grammar
  (gg-grammar
   (start
    _ ms := module-list _ eof -> ms)

   (_  ; optional whitespace
      whitespace | )
   (whitespace
      whitespace whitespace-element
    | whitespace-element)
   (whitespace-element
      "/\\*\\([^*]\\|\\*[^/]\\)*\\*/" ; multiline comment
    | "//[^\n]*"                      ; single line comment
    | "[ \t\n]+")

   (identifier "[a-zA-Z_][a-zA-Z0-9_]*")
   (ident-char "[a-zA-Z0-9_]")

   (module-list
      m := module _ ms := module-list -> (cons m ms)
    | -> nil)

   (module
      "module" _ name := identifier _ ?{ _ decls := declarations _ "};"
      -> (make-wk-module :name name :declarations decls))

   (declarations
      d := interface _ ds := declarations -> (cons d ds)
    | d := typedef _ ds := declarations -> (cons d ds)
    | d := implements _ ds := declarations -> (cons d ds)
    | -> nil)

   (implements
      implementor := identifier _ "implements" _ implemented := identifier _ ?;
      -> (make-wk-implements :implementer name :implemented implemented))

   (typedef
      "typedef" _ ty := type  _ new-name := identifier ?;
      -> (make-wk-typedef :type ty :new-name new-name))

   (ns-name
      module-name := identifier "::" local-name := identifier
      -> (make-wk-ns-name :module-name module-name :local-name local-name))

   (interface
      extended-attributes := extended-attribute-list-opt _ "interface" _
      name := identifier _ extends := extends-opt _ ?{ _ ms := member-list _
      ?} _ ?;
      -> (make-wk-interface :name name :extends extends
                            :extended-attributes extended-attributes
                            :members ms))

   (extended-attribute-list-opt
      ?[ _ attributes := extended-attribute-list-elements _ ?] -> attributes
    | -> nil)

   ; TODO: modify this to find out which attribute lists are missing commas
   (extended-attribute-list-elements
      a := extended-attribute _ ",?" _
      as := extended-attribute-list-elements -> (cons a as)
    | a := extended-attribute _ ",?" -> (list a))

   (extended-attribute
      i := identifier _ ?= _ v := extended-attribute-value
      -> (make-wk-extended-attribute :name i :value v :formals nil)
    | i := identifier _ ?( _ formals := argument-list-opt _ ?)
      -> (make-wk-extended-attribute :name i :value nil :formals formals)
    | i := identifier
      -> (make-wk-extended-attribute :name i :value nil :formals nil))

   (extended-attribute-value
      "[^] ,]+")

   (extends-opt
      ?: _ types := type-list -> types
    | -> nil)

   (member-list
      m := member _ ms := member-list -> (cons m ms)
    | -> nil)

   (member
      member-attribute
    | member-method
    | member-field)

   (member-attribute
      extended-attributes := extended-attribute-list-opt _
      attrs := spec-attributes _
      "attribute" _
      type := type _
      name := identifier _
      getter-raises := getter-raises-opt _
      setter-raises := setter-raises-opt _ ?;
      -> (make-wk-member-attribute :spec-attributes attrs
                                   :extended-attributes extended-attributes
                                   :type type :name name
                                   :getter-raises getter-raises
                                   :setter-raises setter-raises))

   ;; FIXME: model optional array types
   (type
      ty := type-0 ?[ ?] -> (make-wk-ty-array :underlying-type ty)
    | ty := type-0 ?? -> (make-wk-ty-optional :underlying-type ty)
    | type-0)

   (type-0
      type-1
    | ns-name
    | identifier)

   (type-1
      "void" ~ ident-char -> 'void
    | "any" ~ ident-char -> 'any
    | "boolean" ~ ident-char -> 'boolean
    | "octet" ~ ident-char -> 'octet
    | "short" ~ ident-char -> 'short
    | "unsigned short" ~ ident-char -> 'unsigned-short
    ;; put (unsigned) long long first as they are prefixes of (unsigned) long
    | "long long" ~ ident-char -> 'long-long 
    | "unsigned long long" ~ ident-char -> 'unsigned-long-long
    | "long" ~ ident-char -> 'long
    | "unsigned long" ~ ident-char -> 'unsigned-long
    | "float" ~ ident-char -> 'float
    | "double" ~ ident-char -> 'double
    | "DOMString" ~ ident-char -> 'string
    | "object" ~ ident-char -> 'object

    ;; http://lists.w3.org/Archives/Public/public-script-coord/2010JanMar/0000.html
    | "byte" ~ ident-char -> 'byte
    | "unsigned byte" ~ ident-char -> 'unsigned-byte
    )

   (spec-attributes
      "readonly" whitespace xs := spec-attributes
      -> (cons 'readonly xs)
    | "stringifier" whitespace xs := spec-attributes
      -> (cons 'stringifier xs)
    | "getter" whitespace xs := spec-attributes
      -> (cons 'getter xs)
    | "setter" whitespace xs := spec-attributes
      -> (cons 'setter xs)
    | "creator" whitespace xs := spec-attributes
      -> (cons 'creator xs)
    | "deleter" whitespace xs := spec-attributes
      -> (cons 'deleter xs)
    | -> nil)

   (getter-raises-opt
      "getter" _ raises
    | -> nil)

   (setter-raises-opt
      "setter" _ raises
    | -> nil)

   (member-field
      "const" _ type := type _ name := identifier _ ?= _ value := const-value _
      ?;
      -> (make-wk-member-field :type type :name name :value value))

   (const-value
      "[^;]+")  ; TODO: parse numeric constants

   ;; Method identifiers are optional for getter, setter methods
   (member-method
      attrs := spec-attributes _
      extended-attributes := extended-attribute-list-opt _ type := type _
      name := identifier-opt _ ?( _ formals := argument-list-opt _ ?) _
      raises := raises-opt _ ?;
      -> (make-wk-member-method :spec-attributes attrs
                                :extended-attributes extended-attributes
                                :type type :name name :formals formals
                                :raises raises))

   (identifier-opt
      identifier
    | -> nil)

   (argument-list-opt
      argument-list
    | -> nil)

   (argument-list
      a := argument _ ?, _ as := argument-list -> (cons a as)
    | a := argument -> (list a))

   (argument
      direction := direction _
      optional := optional _
      extended-attributes := extended-attribute-list-opt _
      type := type _
      variadic := variadic _
      name := identifier
      -> (make-wk-formal :direction direction :optional optional
                         :variadicity variadic
                         :extended-attributes extended-attributes :type type
                         :name name))

   (direction
      "in" -> 'in
    | -> 'in)  ;; optional, but always 'in'

   (optional
      "optional" -> 'optional
    | -> 'required)

   (variadic
      ?. ?. ?. -> 'variadic
    | -> 'once)

   (raises
      "raises" _ ?( _ types := type-list _ ?) -> types)

   (raises-opt
      raises
    | -> nil)

   (type-list
      type := type _ ?, _ types := type-list -> (cons type types)
    | type := type -> (list type))
))

(defun test-webidl-parser (&optional files debug)
  "Tests the IDL grammar by parsing FILES in `default-idl-root'.

If FILES is nil, tries to parse all of the IDL files in
`default-idl-root'. If DEBUG, then debug messages are printed."
  (in-dir default-idl-root
   (let* ((test-idl-files (or files (find-idl-files)))
          (ok 0)
          (total (length test-idl-files))
          result)
     (dolist (filename test-idl-files)
       (save-excursion
         (find-file (format "%s/%s" default-idl-root filename))
         (goto-char (point-min))
         (setq result (gg-parse webidl-grammar debug))
         (if result
             (incf ok)
           (message "FAIL %s" filename))
         (kill-buffer)))
     (message "Parsed %d of %d (%0.0f%%)" ok total
              (* 100 (/ ok (float total)))))))

;;; For indexing IDL files

;; (defun id-index-webidl-files (output-dir)
;;   "Indexes all of the member names in WebIDL."
;;   (in-dir default-webkit-root
;;    (let* ((webkit-idl-files (find-webkit-idl-files))
;;           (ok 0)
;;           (total (length webkit-idl-files))
;;           (index (make-hash-table :test 'equal))
;;           result)
;;      (dolist (filename webkit-idl-files)
;;        (process-file (format "%s/%s" default-webkit-root filename)
;;         ;; parse the file
;;         (setq result (gg-parse webkit-idl-grammar))
;;         (if result
;;             (progn
;;               (incf ok)
;;               ;; add the contents of the file to the index
;;               (id-index-members-in-module index filename (cdr result))
;;               ;; convert the file to HTML and save it
;;               (id-format-idl-file filename)
;;               (set-visited-file-name (format "%s/%s" output-dir
;;                                              (id-encode-filename filename)))
;;               (save-buffer))
;;           (message "FAIL %s" filename))))
;;      ;; write the index
;;      (save-excursion
;;        (find-file (format "%s/index.html" output-dir))
;;        (id-format-index-page index)
;;        (save-buffer))
;;      ;; output statistics
;;      (message "Parsed %d of %d (%0.0f%%)" ok total
;;               (* 100 (/ ok (float total))))
;;      index)))

;; (defun id-index-members-in-module (hash filename module)
;;   "Indexes members in MODULE from file FILENAME.

;; HASH is a hashtable mapping member names to lists of (FILENAME
;; LINE . INTERFACE-NAME) occurrences."
;;   (let (module-name interface-name)
;;     (wk-visit-ast module 'id-index-members-callback)))

;; (defun id-index-members-callback (node end-p)
;;   "Internal.
;; Visitor callback for `id-index-members-in-module'."
;;   (unless end-p
;;     (case (aref node 0)
;;       ;; record the name of the module, for naming results later
;;       ('cl-struct-wk-module
;;        (setq module-name (wk-module-name node)))

;;       ;; record the name of the interface, for naming results later
;;       ('cl-struct-wk-interface
;;        (setq interface-name
;;              (format "%s::%s" module-name (wk-interface-name node))))

;;       (t
;;        (let (name-line-attrs occurrence)
;;          ;; extract the member name and the line it is defined on
;;          ;; TODO: this could be simplified by adding an abstract wk-member
;;          (setq name-line-attrs
;;                (cond
;;                 ((wk-member-field-p node)
;;                  ;; wk-member-field doesn't have extended attributes
;;                  (list (wk-member-field-name node) (wk-node-line node) nil))
;;                 ((wk-member-attribute-p node)
;;                  (list (wk-member-attribute-name node) (wk-node-line node)
;;                        (wk-member-attribute-extended-attributes node)))
;;                 ((wk-member-method-p node)
;;                  (list (wk-member-method-name node) (wk-node-line node)
;;                        (wk-member-method-extended-attributes node)))
;;                 (t
;;                  nil)))
;;          ;; add the result, if any, to the index table
;;          (when name-line-attrs
;;            (destructuring-bind (name line attrs) name-line-attrs
;;              (setq occurrence
;;                    (list filename line interface-name attrs))
;;              (puthash name
;;                       (cons occurrence (gethash name hash))
;;                       hash)))))))
;;     ;; keep walking
;;     t)
