package analyzer

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	commentRegex = regexp.MustCompile(`//.*|/\*.*?\*/`)
)

var (
	flagSet flag.FlagSet
	groups  string
)

func init() {
	flagSet.StringVar(
		&groups,
		"groups",
		".*",
		"left associative boolean expression of import path regex patterns",
	)
}

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "goimportgroups",
		Doc:   "Checks if go imports are separated into user-defined groups.",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	fileNames, poses := getFileNamesAndPoses(pass)

	for i, filename := range fileNames {
		pos, msg, err := check(filename)
		if err != nil {
			return nil, err
		}

		if msg != "" {
			pass.Reportf(poses[i]+token.Pos(pos), msg)
		}
	}

	return nil, nil
}

func check(filename string) (int, string, error) {
	groupPatterns := strings.Split(groups, ";")

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return 0, "", err
	}

	fileNode, err := parser.ParseFile(token.NewFileSet(), filename, fileBytes, parser.ImportsOnly)
	if err != nil {
		fmt.Println("Error parsing:", err)
		return 0, "", err
	}

	importsStart, importsEnd, errorMessage := getImports(fileNode)
	if errorMessage != "" {
		return importsStart, fmt.Sprintf("File is not goimportgroups-ed: %s", errorMessage), nil
	}

	if importsStart == importsEnd {
		return 0, "", nil
	}

	src := string(fileBytes)

	importsSrc := src[importsStart-1 : importsEnd-1]
	importsSrc = strings.TrimSpace(importsSrc)
	importsSrc = strings.TrimPrefix(importsSrc, "import")
	importsSrc = strings.TrimSpace(importsSrc)
	importsSrc = strings.TrimPrefix(importsSrc, "(")
	importsSrc = strings.TrimSuffix(importsSrc, ")")
	importsSrc = strings.TrimSpace(importsSrc)
	importsSrc = commentRegex.ReplaceAllString(importsSrc, "")

	importLines := strings.Split(importsSrc, "\n")

	var groups [][]string
	var currGroup []string

	for _, line := range importLines {
		line = strings.TrimSpace(line)

		if line == "" {
			groups = append(groups, currGroup)
			currGroup = []string{}
			continue
		}

		importPath := ""

		if strings.HasPrefix(line, "\"") {
			importPath = strings.TrimSuffix(line[1:], "\"")
		} else {
			_, importPath, _ = strings.Cut(line, " ")
			importPath = strings.TrimSpace(importPath)
		}

		currGroup = append(currGroup, importPath)
	}

	groups = append(groups, currGroup)
	currGroup = []string{}

	currPatternI := 0
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}

		for currPatternI < len(groupPatterns) { // ignoring empty groups
			matches, err := match(g[0], groupPatterns[currPatternI])
			if err != nil {
				return 0, "", err
			}

			if matches {
				break
			}

			currPatternI++
		}

		if currPatternI >= len(groupPatterns) {
			return importsStart, "File is not goimportgroups-ed", nil
		}

		for _, imp := range g {
			matches, err := match(imp, groupPatterns[currPatternI])
			if err != nil {
				return 0, "", err
			}

			if !matches {
				return importsStart, "File is not goimportgroups-ed", nil
			}
		}
	}

	return 0, "", nil
}

func getImports(node *ast.File) (int, int, string) {
	start := 0
	end := 0
	found := false
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		if found {
			return int(genDecl.Pos()), int(genDecl.End()), "cannot have two import sections"
		}

		start = int(genDecl.Pos())
		end = int(genDecl.End())
		found = true
	}

	return start, end, ""
}

func match(s string, patterns string) (bool, error) {
	lastAnd := strings.LastIndex(patterns, ",")
	lastOr := strings.LastIndex(patterns, ":")

	if lastAnd > lastOr {
		l, err := match(s, patterns[:lastAnd])
		if err != nil {
			return false, err
		}

		r, err := match(s, patterns[lastAnd+1:])
		if err != nil {
			return false, err
		}

		return l && r, nil
	}

	if lastOr > lastAnd {
		l, err := match(s, patterns[:lastOr])
		if err != nil {
			return false, err
		}

		r, err := match(s, patterns[lastOr+1:])
		if err != nil {
			return false, err
		}

		return l || r, nil
	}

	r, err := regexp.MatchString(fmt.Sprintf("^%s$", patterns), s)
	if err != nil {
		return false, fmt.Errorf("cannot compile regex %s: %w", patterns, err)
	}

	return r, err
}

func getFileNamesAndPoses(pass *analysis.Pass) ([]string, []token.Pos) {
	var fileNames []string
	var poses []token.Pos
	for _, f := range pass.Files {
		fileName := pass.Fset.PositionFor(f.Pos(), true).Filename
		ext := filepath.Ext(fileName)
		if ext != "" && ext != ".go" {
			// position has been adjusted to a non-go file, revert to original file
			fileName = pass.Fset.PositionFor(f.Pos(), false).Filename
		}
		fileNames = append(fileNames, fileName)
		poses = append(poses, f.Pos())
	}
	return fileNames, poses
}
