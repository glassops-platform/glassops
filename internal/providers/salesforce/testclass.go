package salesforce

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	testAnnotation = regexp.MustCompile(`(?i)@\s*istest\b`)
	typeKeyword    = regexp.MustCompile(`(?i)\b(class|interface|enum)\b`)
)

// IsTestClassSource reports whether Apex source declares a test class: its top level type is a
// class and the annotations before it include @IsTest, in any case, with or without arguments
// such as (SeeAllData=false).
//
// This is not an Apex parser. It blanks comments and string literals, then looks at the text
// before the first class, interface or enum keyword. Limits:
//   - An annotation spelled in some way the pattern does not match is missed. The class is then
//     treated as a normal class, which is the safe side, because it must meet the coverage minimum.
//   - A @IsTest method in a normal class sits after the class keyword, so it does not count.
//   - Only the first top level type is read. Inner classes are not looked at.
func IsTestClassSource(src string) bool {
	clean := blankCommentsAndStrings(src)
	loc := typeKeyword.FindStringSubmatchIndex(clean)
	if loc == nil || !strings.EqualFold(clean[loc[2]:loc[3]], "class") {
		return false
	}
	return testAnnotation.MatchString(clean[:loc[0]])
}

// blankCommentsAndStrings replaces comments and string literals with spaces, keeping newlines.
func blankCommentsAndStrings(src string) string {
	b := []byte(src)
	for i := 0; i < len(b); {
		switch {
		case b[i] == '/' && i+1 < len(b) && b[i+1] == '/':
			for ; i < len(b) && b[i] != '\n'; i++ {
				b[i] = ' '
			}
		case b[i] == '/' && i+1 < len(b) && b[i+1] == '*':
			b[i], b[i+1] = ' ', ' '
			i += 2
			for i < len(b) && !(b[i] == '*' && i+1 < len(b) && b[i+1] == '/') {
				if b[i] != '\n' {
					b[i] = ' '
				}
				i++
			}
			for j := 0; j < 2 && i < len(b); j, i = j+1, i+1 {
				b[i] = ' '
			}
		case b[i] == '\'':
			b[i] = ' '
			i++
			for i < len(b) && b[i] != '\'' && b[i] != '\n' {
				if b[i] == '\\' && i+1 < len(b) {
					b[i] = ' '
					i++
				}
				b[i] = ' '
				i++
			}
			if i < len(b) && b[i] == '\'' {
				b[i] = ' '
				i++
			}
		default:
			i++
		}
	}
	return string(b)
}

// testClassNames returns the names in classes whose source file under the package directories
// declares a test class, sorted. A class with no file under classes/ is not listed, so a
// class that cannot be read is treated as a normal class.
func testClassNames(dir string, pkgDirs, classes []string) []string {
	want := map[string]bool{}
	for _, c := range classes {
		want[c] = true
	}
	found := map[string]bool{}
	for _, pd := range pkgDirs {
		root := filepath.Join(dir, filepath.FromSlash(pd))
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".cls") {
				return nil
			}
			name := strings.TrimSuffix(d.Name(), ".cls")
			if !want[name] || found[name] {
				return nil
			}
			if data, rerr := os.ReadFile(path); rerr == nil && IsTestClassSource(string(data)) {
				found[name] = true
			}
			return nil
		})
	}
	out := []string{}
	for n := range found {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
