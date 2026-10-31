// Package junit writes test results as JUnit XML, the format CI systems read to show test
// reports. It uses only encoding/xml.
package junit

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// Suites is the root element, <testsuites>.
type Suites struct {
	XMLName  xml.Name `xml:"testsuites"`
	Name     string   `xml:"name,attr,omitempty"`
	Tests    int      `xml:"tests,attr"`
	Failures int      `xml:"failures,attr"`
	Errors   int      `xml:"errors,attr"`
	Skipped  int      `xml:"skipped,attr"`
	Time     float64  `xml:"time,attr"`
	Suites   []Suite  `xml:"testsuite"`
}

// Suite is one <testsuite>.
type Suite struct {
	Name     string  `xml:"name,attr"`
	Tests    int     `xml:"tests,attr"`
	Failures int     `xml:"failures,attr"`
	Errors   int     `xml:"errors,attr"`
	Skipped  int     `xml:"skipped,attr"`
	Time     float64 `xml:"time,attr"`
	Cases    []Case  `xml:"testcase"`
}

// Case is one <testcase>. Failure and Skipped are nil for a test that passed.
type Case struct {
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr,omitempty"`
	Time      float64  `xml:"time,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
	Skipped   *Skipped `xml:"skipped,omitempty"`
}

// Failure marks a test that failed. Message is the short reason and Text the detail, such as
// a stack trace.
type Failure struct {
	Message string `xml:"message,attr,omitempty"`
	Type    string `xml:"type,attr,omitempty"`
	Text    string `xml:",chardata"`
}

// Skipped marks a test that did not run.
type Skipped struct {
	Message string `xml:"message,attr,omitempty"`
}

// Add appends a suite to s and updates the totals. It counts the suite's cases itself, so a
// caller does not set the counts.
func (s *Suites) Add(suite Suite) {
	suite.Tests, suite.Failures, suite.Skipped, suite.Time = 0, 0, 0, 0
	for _, c := range suite.Cases {
		suite.Tests++
		suite.Time += c.Time
		if c.Failure != nil {
			suite.Failures++
		}
		if c.Skipped != nil {
			suite.Skipped++
		}
	}
	s.Suites = append(s.Suites, suite)
	s.Tests += suite.Tests
	s.Failures += suite.Failures
	s.Errors += suite.Errors
	s.Skipped += suite.Skipped
	s.Time += suite.Time
}

// Write writes the XML, with its declaration, to w.
func (s Suites) Write(w io.Writer) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("junit: encode: %w", err)
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// Marshal returns the XML as bytes.
func (s Suites) Marshal() ([]byte, error) {
	var b bytes.Buffer
	if err := s.Write(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
