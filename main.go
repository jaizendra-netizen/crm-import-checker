package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	req := flag.String("require", "", "comma-separated fields that must not be empty")
	stages := flag.String("stages", "", "allowed values for the stage column (default mql,sql,opportunity)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: crm-import-checker [-require f1,f2] [-stages a,b,c] file.csv")
		os.Exit(2)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	defer f.Close()
	issues, rows, err := Check(f, Options{Required: split(*req), Stages: split(*stages)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	for _, i := range issues {
		fmt.Printf("row %d  %-13s %s\n", i.Row, i.Kind, i.Message)
	}
	fmt.Printf("%d rows checked, %d issue(s)\n", rows, len(issues))
	if len(issues) > 0 {
		os.Exit(1)
	}
}
