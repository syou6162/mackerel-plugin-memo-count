package mpmemocount

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// MemoCountPlugin is mackerel plugin
type MemoCountPlugin struct {
	prefix string
	dir    string
}

// GraphDefinition interface for mackerelplugin
func (p *MemoCountPlugin) GraphDefinition() map[string]mp.Graphs {
	ret := make(map[string]mp.Graphs)
	ret["memo_count"] = mp.Graphs{
		Label: p.prefix + " Count",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "file_count", Label: "File Count"},
			{Name: "line_count", Label: "Line Count"},
		},
	}
	return ret
}

// FetchMetrics interface for mackerelplugin
func (p *MemoCountPlugin) FetchMetrics() (map[string]float64, error) {
	ret := make(map[string]float64)

	files, err := getMarkdownFilenames(p.dir)
	if err != nil {
		return nil, err
	}

	ret["file_count"] = float64(len(files))

	lineCnt := 0
	for _, filename := range files {
		cnt := lineCount(p.dir + "/" + filename)
		lineCnt += cnt
	}

	ret["line_count"] = float64(lineCnt)
	return ret, nil
}

func filterMarkdown(files []string) []string {
	var newfiles []string
	for _, file := range files {
		if strings.HasSuffix(file, ".md") {
			newfiles = append(newfiles, file)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(newfiles)))
	return newfiles
}

func getMarkdownFilenames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	files, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	files = filterMarkdown(files)
	return files, nil
}

func lineCount(filename string) int {
	lineCnt := 0

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return lineCnt
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		lineCnt++
	}
	return lineCnt
}

// Do the plugin
func Do() {
	var (
		optPrefix = flag.String("metric-key-prefix", "Memo", "Metric key prefix")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION] /path/to/memo_dir\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	mp.NewMackerelPlugin(&MemoCountPlugin{
		prefix: *optPrefix,
		dir:    flag.Args()[0],
	}).Run()
}
