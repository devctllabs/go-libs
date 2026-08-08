package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var schemaFiles = []string{
	"v1/InitializeParams.json",
	"v1/InitializeResponse.json",
	"v2/ModelListParams.json",
	"v2/ModelListResponse.json",
	"v2/PermissionProfileListParams.json",
	"v2/PermissionProfileListResponse.json",
	"v2/ThreadStartParams.json",
	"v2/ThreadStartResponse.json",
	"v2/ThreadResumeParams.json",
	"v2/ThreadResumeResponse.json",
	"v2/ThreadReadParams.json",
	"v2/ThreadReadResponse.json",
	"v2/TurnStartParams.json",
	"v2/TurnStartResponse.json",
	"v2/TurnInterruptParams.json",
	"v2/TurnInterruptResponse.json",
	"v2/TurnStartedNotification.json",
	"v2/TurnCompletedNotification.json",
	"v2/ItemStartedNotification.json",
	"v2/ItemCompletedNotification.json",
	"v2/AgentMessageDeltaNotification.json",
	"v2/PlanDeltaNotification.json",
	"v2/ReasoningSummaryTextDeltaNotification.json",
	"v2/ReasoningSummaryPartAddedNotification.json",
	"v2/ReasoningTextDeltaNotification.json",
	"v2/CommandExecutionOutputDeltaNotification.json",
	"v2/TurnDiffUpdatedNotification.json",
	"v2/TurnPlanUpdatedNotification.json",
	"v2/ErrorNotification.json",
	"v2/WarningNotification.json",
	"v2/ThreadTokenUsageUpdatedNotification.json",
	"v2/ServerRequestResolvedNotification.json",
	"CommandExecutionRequestApprovalParams.json",
	"CommandExecutionRequestApprovalResponse.json",
	"FileChangeRequestApprovalParams.json",
	"FileChangeRequestApprovalResponse.json",
	"PermissionsRequestApprovalParams.json",
	"PermissionsRequestApprovalResponse.json",
	"ToolRequestUserInputParams.json",
	"ToolRequestUserInputResponse.json",
	"McpServerElicitationRequestParams.json",
	"McpServerElicitationRequestResponse.json",
}

const generatedGoArtifactName = "protocol.gen.go"

func generatedSchemaName(source string) string {
	return strings.TrimSuffix(source, filepath.Ext(source)) + ".gen.json"
}

func artifactRoots() []string {
	return []string{"schema", generatedGoArtifactName, "schema-version.txt", "version.gen.go"}
}

func main() {
	check := flag.Bool("check", false, "compare generated artifacts without modifying the repository")
	codex := flag.String("codex", "codex", "Codex CLI executable")
	quicktype := flag.String("quicktype", "quicktype", "quicktype executable")
	flag.Parse()

	if err := run(*check, *codex, *quicktype); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(check bool, codex, quicktype string) error {
	protocolDir, err := protocolDirectory()
	if err != nil {
		return err
	}
	temporary, err := os.MkdirTemp("", "codexapp-protocol-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporary)

	upstream := filepath.Join(temporary, "upstream")
	//nolint:gosec // Generated source trees use conventional directory permissions.
	if err := os.MkdirAll(upstream, 0o755); err != nil {
		return fmt.Errorf("create upstream schema directory: %w", err)
	}
	if err := command(codex, "app-server", "generate-json-schema", "--experimental", "--out", upstream); err != nil {
		return err
	}

	artifacts := filepath.Join(temporary, "artifacts")
	schemaDir := filepath.Join(artifacts, "schema")
	for _, name := range schemaFiles {
		if err := copyFile(filepath.Join(upstream, name), filepath.Join(schemaDir, generatedSchemaName(name))); err != nil {
			return err
		}
	}

	generated := filepath.Join(artifacts, generatedGoArtifactName)
	arguments := []string{"--lang", "go", "--src-lang", "schema", "--package", "protocol"}
	for _, name := range schemaFiles {
		arguments = append(arguments, "--src", filepath.Join(schemaDir, generatedSchemaName(name)))
	}
	arguments = append(arguments, "--out", generated)
	if err := command(quicktype, arguments...); err != nil {
		return err
	}
	if err := formatGo(generated); err != nil {
		return err
	}

	versionOutput, err := exec.Command(codex, "--version").Output()
	if err != nil {
		return fmt.Errorf("run %s --version: %w", codex, err)
	}
	version := strings.TrimSpace(string(versionOutput))
	//nolint:gosec // Generated source artifacts use conventional file permissions.
	if err := os.WriteFile(filepath.Join(artifacts, "schema-version.txt"), []byte(version+"\n"), 0o644); err != nil {
		return fmt.Errorf("write schema version: %w", err)
	}
	if err := writeVersionGo(filepath.Join(artifacts, "version.gen.go"), version); err != nil {
		return err
	}

	if check {
		return compareArtifacts(artifacts, protocolDir)
	}
	return installArtifacts(artifacts, protocolDir)
}

func protocolDirectory() (string, error) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("resolve generator source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..")), nil
}

func command(name string, arguments ...string) error {
	//nolint:gosec // Executables are explicit generator command-line inputs.
	cmd := exec.Command(name, arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", name, err)
	}
	return nil
}

func copyFile(source, target string) error {
	//nolint:gosec // Paths are derived from the generator's private temporary tree and repository schema tree.
	contents, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}
	//nolint:gosec // Generated source trees use conventional directory permissions.
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(target), err)
	}
	//nolint:gosec // Generated source artifacts use conventional file permissions.
	if err := os.WriteFile(target, contents, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	return nil
}

func formatGo(path string) error {
	//nolint:gosec // The path is an artifact inside the generator's private temporary tree.
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated Go: %w", err)
	}
	formatted, err := format.Source(contents)
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}
	//nolint:gosec // Generated Go source uses conventional file permissions.
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write formatted Go: %w", err)
	}
	return nil
}

func writeVersionGo(path, version string) error {
	source := fmt.Sprintf("// Code generated by codexapp protocol generator. DO NOT EDIT.\n\npackage protocol\n\nconst CodexVersion = %q\n", version)
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return fmt.Errorf("format generated version: %w", err)
	}
	//nolint:gosec // Generated Go source uses conventional file permissions.
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write generated version: %w", err)
	}
	return nil
}

func installArtifacts(source, target string) error {
	if err := os.RemoveAll(filepath.Join(target, "schema")); err != nil {
		return fmt.Errorf("remove old schema snapshot: %w", err)
	}
	for _, name := range artifactRoots() {
		from := filepath.Join(source, name)
		to := filepath.Join(target, name)
		info, err := os.Stat(from)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := copyTree(from, to); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(from, to); err != nil {
			return err
		}
	}
	return nil
}

func compareArtifacts(source, target string) error {
	wanted, err := filesUnder(source)
	if err != nil {
		return err
	}
	actual := make([]string, 0, len(wanted))
	for _, root := range artifactRoots() {
		path := filepath.Join(target, root)
		info, statErr := os.Stat(path)
		if statErr != nil {
			return fmt.Errorf("generated artifacts are stale: %s is missing", path)
		}
		if info.IsDir() {
			files, walkErr := filesUnder(path)
			if walkErr != nil {
				return walkErr
			}
			for _, name := range files {
				actual = append(actual, filepath.Join(root, name))
			}
		} else {
			actual = append(actual, root)
		}
	}
	sort.Strings(actual)
	if !equalStrings(wanted, actual) {
		return errors.New("generated artifacts are stale: file set differs; run mise run codexapp:generate")
	}
	for _, name := range wanted {
		//nolint:gosec // Both paths are confined to the known artifact and repository schema trees.
		expected, readErr := os.ReadFile(filepath.Join(source, name))
		if readErr != nil {
			return readErr
		}
		//nolint:gosec // Both paths are confined to the known artifact and repository schema trees.
		got, readErr := os.ReadFile(filepath.Join(target, name))
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(expected, got) {
			return fmt.Errorf("generated artifact %s is stale; run mise run codexapp:generate", name)
		}
	}
	return nil
}

func filesUnder(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, relative)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			//nolint:gosec // Generated source trees use conventional directory permissions.
			return os.MkdirAll(filepath.Join(target, relative), 0o755)
		}
		return copyFile(path, filepath.Join(target, relative))
	})
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
