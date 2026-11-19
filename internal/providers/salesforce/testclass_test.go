package salesforce

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsTestClassSource(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"lower case", "@isTest\npublic class TestFactory {}", true},
		{"mixed case", "@IsTest\nprivate class RestTestUtil {}", true},
		{"upper case", "@ISTEST public class X {}", true},
		{"with arguments", "@IsTest(SeeAllData=false)\npublic class X {}", true},
		{"space before arguments", "@IsTest (isParallel=true) public with sharing class X {}", true},
		{"same line", "@isTest private class X { }", true},
		{"after a header comment", "/**\n * Helpers.\n */\n// note\n@IsTest\npublic class X {}", true},
		{"comment between annotation and class", "@IsTest\n// why\n/* more */\n\npublic class X {}", true},
		{"other annotation first", "@SuppressWarnings('PMD')\n@IsTest\npublic class X {}", true},
		{"CRLF", "@IsTest\r\npublic class X {}\r\n", true},
		{"normal class", "public with sharing class Greeter { public String hi() { return 'hi'; } }", false},
		{"method level only", "public class Greeter {\n  @IsTest\n  static void t() {}\n}", false},
		{"commented line annotation", "// @IsTest\npublic class X {}", false},
		{"commented block annotation", "/* @IsTest */ public class X {}", false},
		{"annotation inside a string", "@SuppressWarnings('@IsTest')\npublic class X {}", false},
		{"string with escaped quote", "@SuppressWarnings('it\\'s @IsTest')\npublic class X {}", false},
		{"similar annotation", "@IsTestable\npublic class X {}", false},
		{"unterminated comment", "/* @IsTest\npublic class X {}", false},
		{"interface", "@IsTest\npublic interface X {}", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTestClassSource(tc.src); got != tc.want {
				t.Fatalf("IsTestClassSource(%q) = %v, want %v", tc.src, got, tc.want)
			}
		})
	}
}

// The plan records the changed classes whose source is annotated @IsTest, whatever their names.
func TestPlanRecordsTestClassesFromSource(t *testing.T) {
	dir := planProject(t)
	classes := filepath.Join(dir, "force-app", "main", "default", "classes")
	if err := os.MkdirAll(classes, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, src string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(classes, name+".cls"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The sgd fixture lists Greeter and GreeterTest as changed.
	write("Greeter", "public class Greeter {\n  @IsTest static void t() {}\n}")
	write("GreeterTest", "public class GreeterTest {}")
	var calls []Cmd
	cs, err := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls)).Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{}; !reflect.DeepEqual(cs.Facts["salesforce.apex_test_classes"], want) {
		t.Fatalf("without annotations: %#v, want %#v", cs.Facts["salesforce.apex_test_classes"], want)
	}

	write("GreeterTest", "// helper\n@isTest\nprivate class GreeterTest {}")
	cs, err = planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls)).Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"GreeterTest"}; !reflect.DeepEqual(cs.Facts["salesforce.apex_test_classes"], want) {
		t.Fatalf("with annotation: %#v, want %#v", cs.Facts["salesforce.apex_test_classes"], want)
	}
}

// A test class that is not in the change set is not listed, even when its source is annotated.
func TestTestClassNamesOnlyListsRequestedClasses(t *testing.T) {
	dir := t.TempDir()
	classes := filepath.Join(dir, "force-app", "classes")
	if err := os.MkdirAll(classes, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"A", "B"} {
		if err := os.WriteFile(filepath.Join(classes, n+".cls"), []byte("@IsTest class "+n+" {}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := testClassNames(dir, []string{"force-app"}, []string{"A", "Missing"})
	if want := []string{"A"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
