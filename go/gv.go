// This gv.go file contains some extra debug/visualization stuff
// unrelated to the actual train routing logic.
package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
)

func init() {
	http.HandleFunc("/gv", handleGV)
	http.HandleFunc("/rgv", handleGVRaw)
}

func (n Navi) GV(w io.Writer, g StationGraph, penWidth int) {
	fmt.Fprintln(w, `graph g {`)
	fmt.Fprintln(w, `  graph [overlap=scale]`)
	done := make(map[string]bool)
	keyFn := func(x, y string) string {
		if x < y {
			y, x = x, y
		}
		return x + ":" + y
	}
	for x, ym := range g {
		for y, lines := range ym {
			key := keyFn(x, y)
			if done[key] || x == y {
				continue
			}
			for line := range lines {
				fmt.Fprintf(w, `  "%s" -- "%s" [color="%s" penwidth=%d]`, x, y, n.Lines[line].Color, penWidth)
				fmt.Fprintf(w, "\n")
			}
			done[key] = true
		}
	}
	fmt.Fprintln(w, "}")
}

func handleGV(w http.ResponseWriter, r *http.Request) {
	_, n := LoadNavi(r)
	w.Header().Set("Content-Type", "text/gv; charset=utf-8")
	switch r.FormValue("adj") {
	case "lines":
		n.GV(w, n.LineAdjacency, 1)
	default:
		n.GV(w, n.Adjacency, 5)
	}
}

var rawLineTmpl = template.Must(template.New("gv").Funcs(template.FuncMap{
	"LineColor": func(s string) string { return "" },
	"Connect": func(l string, stations []string) string {
		if len(stations) < 1 {
			return ""
		}
		edges := []string{}
		nodes := []string{}
		asNode := func(s string) string {
			return "l_" + l + "_" + s
		}
		for _, s := range stations {
			n := asNode(s)
			nodes = append(nodes, fmt.Sprintf(`%s [label="%s"]`, n, s))
			edges = append(edges, n)
		}
		return fmt.Sprintf("%s\n%s", strings.Join(nodes, "\n"), strings.Join(edges, " -- "))
	},
}).Parse(`
graph g {
  graph [overlap=scale]

{{range .Network}}
subgraph line_{{.Name}} {
  {{$line := .}}
  node [rank=same]
  edge [color="{{LineColor .Name}}" penwidth=5]
  {{Connect .Name .Stations}}
  }
{{end}}
}
`))

func handleGVRaw(w http.ResponseWriter, r *http.Request) {
	_, n := LoadNavi(r)
	w.Header().Set("Content-Type", "text/gv; charset=utf-8")
	err := rawLineTmpl.Funcs(template.FuncMap{
		"LineColor": func(s string) string {
			return n.Lines[s].Color
		},
	}).ExecuteTemplate(w, "gv", n)
	if err != nil {
		fmt.Fprintln(w, "!!!", err)
	}
}
