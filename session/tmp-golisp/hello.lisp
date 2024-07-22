; run with: `rlwrap go run *.go hello.lisp` or `rlwrap go run *.go hello.lisp World`
; — or load this in a REPL session started with `rlwrap go run *.go` by calling `(loadFile "hello.lisp")`

(def greet
    (fn (name)
        (def repeat (not (bool name)))
        (set name (or name (readLine "Name: ")))
        (if (= "" name)
            (set repeat :false)
            (println "Hello, " name "!"))
        (if repeat (greet :nil))))

(if (isEmpty osArgs)
    (greet :nil)
    (map greet osArgs))
