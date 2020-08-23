// https://github.com/CentOS/mirrorlists-code/blob/master/frontend/ml.py
// https://github.com/fedora-infra/mirrormanager2/blob/0.14/mirrorlist/mirrorlist_server.py
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

type config struct {
    Listen string
    Mirrors mirrorMap
}

type mirrorListHandler struct {
    config config
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
    mirrors := h.config.Mirrors
    var urls []string

    repos, ok := mirrors[q.Release]
    if !ok {
        repos, ok = mirrors["*"]
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

func loadConfig(path string, cfg *config) error {
    // Set defaults
    cfg.Listen = ":8080"

    // Open config file
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("Error opening config: %w", err)
    }

    // Decode yaml file into config object
    d := yaml.NewDecoder(f)
    err = d.Decode(cfg)
    if err != nil {
        return fmt.Errorf("Error reading config: %w", err)
    }

    // Verify config was populated
    if len(cfg.Mirrors) == 0 {
        return fmt.Errorf("Error reading config: failed to populate mirrors")
    }

    return nil
}

func main() {
    var err error
    var handler mirrorListHandler

    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: mirrorlist configfile\n")
        os.Exit(1)
    }
    configPath := os.Args[1]

    err = loadConfig(configPath, &handler.config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    http.Handle("/", &handler)

    addr := handler.config.Listen
    log.Println("Serving on " + addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
