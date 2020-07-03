// https://github.com/CentOS/mirrorlists-code/blob/master/frontend/ml.py
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
    "strings"
)

func getOne(vals url.Values, key string) string {
    v := vals[key]
    if len(v) != 1 {
        return ""
    }
    return v[0]
}

func handleRoot(w http.ResponseWriter, req *http.Request) {
    // The "/" pattern matches everything, so we need to check
    // that we're at the root here.
    if req.URL.Path != "/" {
        http.NotFound(w, req)
        return
    }

    if req.Method != "GET" {
        http.Error(w, "bad method", http.StatusBadRequest)
        return
    }

    //log.Printf("%s %v\n", req.Method, req.URL)
    vals := req.URL.Query()

    arch := getOne(vals, "arch")
    if arch == "" {
        http.Error(w, "arch not specified", http.StatusBadRequest)
        return
    }

    repo := strings.ToLower(getOne(vals, "repo"))
    if repo == "" {
        http.Error(w, "repo not specified", http.StatusBadRequest)
        return
    }

    release := getOne(vals, "release")
    if release == "" {
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
