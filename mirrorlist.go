// https://github.com/CentOS/mirrorlists-code/blob/master/frontend/ml.py
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "strings"

    "github.com/go-yaml/yaml"
)

// mirrorMap is a mapping of release: repo: arch: list of urls
type mirrorMap map[string]map[string]map[string][]string

type mirrorListHandler struct {
    mirrors mirrorMap
}

func getOne(vals url.Values, key string) string {
    v := vals[key]
    if len(v) != 1 {
        return ""
    }
    return v[0]
}

func (h *mirrorListHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    log.Printf("%s %s %v\n", req.RemoteAddr, req.Method, req.URL)

    if req.Method != "GET" {
        http.Error(w, "bad method", http.StatusBadRequest)
        return
    }

    // The "/" pattern matches everything, so we need to check
    // that we're at the root here.
    if req.URL.Path != "/" {
        http.NotFound(w, req)
        return
    }

    // Handle the query string
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

    // Look up url list
    urls, ok := h.mirrors[release][repo][arch]
    if !ok {
        http.Error(w, "Invalid release/repo/arch combination", http.StatusNotFound)
        return
    }

    for _, url := range urls {
        fmt.Fprintln(w, url)
    }
}


func loadConfig(path string) (mirrorMap, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("Error opening config: %w", err)
    }

    d := yaml.NewDecoder(f)

    v := make(mirrorMap)
    err = d.Decode(v)
    if err != nil {
        return nil, fmt.Errorf("Error reading config: %w", err)
    }

    return v, nil
}

func main() {
    var err error
    var handler mirrorListHandler

    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: mirrorlist configfile\n")
        os.Exit(1)
    }
    configPath := os.Args[1]

    handler.mirrors, err = loadConfig(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    http.Handle("/", &handler)

    addr := ":8080"
    log.Println("Serving on " + addr)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
