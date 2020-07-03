// https://github.com/CentOS/mirrorlists-code/blob/master/frontend/ml.py
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "strings"
    "text/template"

    "github.com/go-yaml/yaml"
)

// mirrorMap is a mapping of release: repo: arch: list of urls
type mirrorMap map[string]map[string]map[string][]string

type mirrorListHandler struct {
    mirrors mirrorMap
}

type Qualifier struct {
    Arch string
    Release string
    Repo string
}


func getOne(vals url.Values, key string) string {
    v := vals[key]
    if len(v) != 1 {
        return ""
    }
    return v[0]
}

func (h *mirrorListHandler) lookupUrls(q Qualifier) ([]string, error) {
    var urls []string

    repos, ok := h.mirrors[q.Release]
    if !ok {
        repos, ok = h.mirrors["*"]
        if !ok {
            return nil, fmt.Errorf("Invalid release")
        }
    }

    archs, ok := repos[q.Repo]
    if !ok {
        archs, ok = repos["*"]
        if !ok {
            return nil, fmt.Errorf("Invalid repo")
        }
    }

    urls, ok = archs[q.Arch]
    if !ok {
        urls, ok = archs["*"]
        if !ok {
            return nil, fmt.Errorf("Invalid arch")
        }
    }

    return urls, nil
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

    var q Qualifier

    q.Arch = getOne(vals, "arch")
    if q.Arch == "" {
        http.Error(w, "arch not specified", http.StatusBadRequest)
        return
    }

    q.Repo = strings.ToLower(getOne(vals, "repo"))
    if q.Repo == "" {
        http.Error(w, "repo not specified", http.StatusBadRequest)
        return
    }

    q.Release = getOne(vals, "release")
    if q.Release == "" {
        http.Error(w, "release not specified", http.StatusBadRequest)
        return
    }

    // Look up url list
    urls, err := h.lookupUrls(q)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }


    t := template.New("url")

    for _, url := range urls {
        tp, err := t.Parse(url)
        if err != nil {
            log.Printf("Error parsing template: %v", err)
            continue
        }

        tp.Execute(w, q)
        fmt.Fprintln(w, "")
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
