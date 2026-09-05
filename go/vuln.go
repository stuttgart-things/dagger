package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// govulncheck exit codes. 0 and 3 are the two documented outcomes of a
// successful scan; anything else means the tool itself failed and the report
// is not trustworthy.
const (
	govulncheckExitClean   = 0
	govulncheckExitFinding = 3
)

// reportExtensions maps a -format value to the extension its output belongs in.
// openvex emits JSON, so it shares the .json extension.
var reportExtensions = map[string]string{
	"text":    "txt",
	"json":    "json",
	"sarif":   "sarif",
	"openvex": "json",
}

// Govulncheck scans Go source for known vulnerabilities that are reachable
// from the call graph.
//
// This is not the same question as SecurityScan (gosec: did we write something
// unsafe) or the trivy module (does this artefact contain a package with a
// known CVE). govulncheck reads the Go vulnerability database and then does
// call-graph analysis, so it reports only vulnerabilities on a path the binary
// can actually reach, and stays quiet about a CVE in a dependency whose
// affected function is never called. That is what makes it usable as a build
// gate rather than as noise.
//
// Two behaviours of the tool shape this function.
//
// First, the report is always written, including when vulnerabilities are
// found. govulncheck exits 3 in that case, and a WithExec that propagated the
// exit code would leave no file to export -- the report would go missing
// precisely when it is wanted. The exit code is captured separately instead.
//
// Second, that exit code is only meaningful for -format text. With json,
// sarif and openvex govulncheck exits 0 even when it finds reachable
// vulnerabilities, on the assumption that the caller parses the output. So
// when a gate is asked for in one of those formats, a second text-mode pass
// supplies the verdict; the requested format still produces the report. Both
// passes run in the same container and share the module cache and the
// downloaded vulnerability database.
//
// A consequence worth knowing before wiring this into a workflow: with
// failOnFinding set, this function returns an error, so a `dagger call
// govulncheck ... export --path report.txt` never reaches the export. To both
// keep the report and gate the build, call twice -- once to export, once to
// gate. The second call hits Dagger's cache and costs nothing.
func (m *Go) Govulncheck(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="v1.7.0"
	govulncheckVersion string,
	// +optional
	// +default="text"
	format string, // text | json | sarif | openvex
	// Base toolchain for the scan container. A floor, not a pin: GOTOOLCHAIN
	// is set to auto, so a go.mod requiring something newer wins.
	// +optional
	// +default="1.25.5"
	goVersion string,
	// +optional
	// +default="alpine"
	variant string,
	// +optional
	// +default=false
	failOnFinding bool,
	// Skip Dagger's result cache. The vulnerability database changes
	// independently of the source, so a cached hit can report a clean tree
	// against a database from weeks ago.
	// +optional
	// +default=false
	noCache bool,
) (*dagger.File, error) {

	ext, ok := reportExtensions[format]
	if !ok {
		return nil, fmt.Errorf(
			"unsupported format %q: want one of text, json, sarif, openvex", format)
	}

	const (
		srcDir     = "/src"
		reportDir  = "/report"
		binaryPath = "/go/bin/govulncheck"

		reportErrPath = reportDir + "/report.stderr"
		reportExitAt  = reportDir + "/report.exit"
		gateOutPath   = reportDir + "/gate.txt"
		gateExitAt    = reportDir + "/gate.exit"
	)
	reportPath := fmt.Sprintf("%s/govulncheck.%s", reportDir, ext)

	ctr := m.GetGoLangContainer(goVersion, variant).
		// The golang images ship GOTOOLCHAIN=local, which pins the scan to
		// goVersion and makes it fail outright against a go.mod that requires
		// something newer. Reachability analysis depends on the toolchain, so
		// the toolchain the code declares is the one that answers the right
		// question: let Go fetch it. goVersion is the floor, not the ceiling.
		WithEnvVariable("GOTOOLCHAIN", "auto").
		WithDirectory(srcDir, src).
		WithWorkdir(srcDir).
		WithExec([]string{"mkdir", "-p", reportDir}).
		WithExec([]string{
			"go", "install",
			"golang.org/x/vuln/cmd/govulncheck@" + govulncheckVersion,
		})

	if noCache {
		// Changing an env var changes the container definition, which is what
		// Dagger keys its cache on. A timestamp makes every call a miss.
		ctr = ctr.WithEnvVariable(
			"GOVULNCHECK_CACHE_BUSTER",
			strconv.FormatInt(time.Now().UnixNano(), 10),
		)
	}

	// Run the scan without letting a non-zero exit abort the pipeline. stderr
	// is kept out of the report so json and sarif output stays parseable.
	ctr = ctr.WithExec([]string{"sh", "-c", fmt.Sprintf(
		"%s -format %s ./... > %s 2> %s; echo $? > %s",
		binaryPath, format, reportPath, reportErrPath, reportExitAt,
	)})

	// A gate in a machine format needs a text-mode pass, because only text
	// mode sets exit 3 on a finding.
	needsGatePass := failOnFinding && format != "text"
	if needsGatePass {
		ctr = ctr.WithExec([]string{"sh", "-c", fmt.Sprintf(
			"%s ./... > %s 2>&1; echo $? > %s",
			binaryPath, gateOutPath, gateExitAt,
		)})
	}

	reportExit, err := readExitCode(ctx, ctr, reportExitAt)
	if err != nil {
		return nil, err
	}

	// Whether the report run's exit code is allowed to be 3 depends on the
	// format: text signals findings that way, the others never do.
	scanFailed := reportExit != govulncheckExitClean
	if format == "text" && reportExit == govulncheckExitFinding {
		scanFailed = false
	}
	if scanFailed {
		detail, readErr := ctr.File(reportErrPath).Contents(ctx)
		if readErr != nil || strings.TrimSpace(detail) == "" {
			detail = "(no error output captured)"
		}
		return nil, fmt.Errorf(
			"govulncheck failed with exit code %d:\n%s", reportExit, detail)
	}

	report := ctr.File(reportPath)
	if !failOnFinding {
		return report, nil
	}

	// Decide the gate from whichever run carries a meaningful exit code.
	gateExit := reportExit
	gateOutput := reportPath
	if needsGatePass {
		gateExit, err = readExitCode(ctx, ctr, gateExitAt)
		if err != nil {
			return nil, err
		}
		gateOutput = gateOutPath

		if gateExit != govulncheckExitClean && gateExit != govulncheckExitFinding {
			detail, readErr := ctr.File(gateOutPath).Contents(ctx)
			if readErr != nil {
				detail = "(no output captured)"
			}
			return nil, fmt.Errorf(
				"govulncheck gate pass failed with exit code %d:\n%s", gateExit, detail)
		}
	}

	if gateExit == govulncheckExitFinding {
		summary, readErr := ctr.File(gateOutput).Contents(ctx)
		if readErr != nil {
			summary = "(report unavailable)"
		}
		return nil, fmt.Errorf(
			"govulncheck found reachable vulnerabilities:\n%s", summary)
	}

	return report, nil
}

// readExitCode reads a single integer written by `echo $?`.
func readExitCode(ctx context.Context, ctr *dagger.Container, path string) (int, error) {
	raw, err := ctr.File(path).Contents(ctx)
	if err != nil {
		return 0, fmt.Errorf("running govulncheck: %w", err)
	}

	code, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("reading govulncheck exit code %q: %w", raw, err)
	}

	return code, nil
}
