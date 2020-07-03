// https://github.com/CentOS/mirrorlists-code/blob/master/frontend/ml.py
package main

import (
    "fmt"
    "log"
    "net/http"
)

func handleRoot(w http.ResponseWriter, req *http.Request) {
    // The "/" pattern matches everything, so we need to check
    // that we're at the root here.
    if req.URL.Path != "/" {
        http.NotFound(w, req)
        return
    }

    log.Printf("%s %v\n", req.Method, req.URL)
    vals := req.URL.Query()
    log.Printf("%v\n", vals)

    arch := vals["arch"]
    if len(arch) != 1 {
        http.Error(w, "arch not specified", http.StatusBadRequest)
        return
    }

    repo := vals["repo"]
    if len(repo) != 1 {
        http.Error(w, "repo not specified", http.StatusBadRequest)
        return
    }

    release := vals["release"]
    if len(release) != 1 {
        http.Error(w, "release not specified", http.StatusBadRequest)
        return
    }

    log.Printf("  arch=%q repo=%q release=%q\n", arch, repo, release)
    fmt.Fprintln(w, "http://sjc.edge.kernel.org/centos/6.10/os/x86_64/")
}

func main() {
    http.HandleFunc("/", handleRoot)

    addr := ":8080"
    log.Println("Serving on " + addr)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
