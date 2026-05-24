//go:build ignore

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifest struct {
	Schema   int             `json:"schema"`
	Header   artifact        `json:"header"`
	Libs     []libArtifact   `json:"libs"`
	Versions []versionRecord `json:"versions"`
}

type artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type libArtifact struct {
	Platform string `json:"platform"`
	Target   string `json:"target"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
}

type versionRecord struct {
	Platform string `json:"platform"`
	Target   string `json:"target"`
	Path     string `json:"path"`
	Commit   string `json:"commit"`
	SHA256   string `json:"sha256"`
}

var platformTargets = map[string]string{
	"darwin_amd64": "x86_64-apple-darwin",
	"darwin_arm64": "aarch64-apple-darwin",
	"linux_amd64":  "x86_64-unknown-linux-gnu",
	"linux_arm64":  "aarch64-unknown-linux-gnu",
}

func main() {
	write := flag.Bool("write", false, "rewrite internal/clib/MANIFEST.json")
	flag.Parse()

	got, err := buildManifest()
	if err != nil {
		fatal(err)
	}

	const manifestPath = "internal/clib/MANIFEST.json"
	if *write {
		if err := writeManifest(manifestPath, got); err != nil {
			fatal(err)
		}
		return
	}

	want, err := readManifest(manifestPath)
	if err != nil {
		fatal(err)
	}
	if !equalManifest(want, got) {
		fmt.Fprintln(os.Stderr, "internal/clib/MANIFEST.json is out of date; run `go run internal/clib/manifest.go -write`")
		fatal(diffManifest(want, got))
	}
}

func buildManifest() (manifest, error) {
	platforms := make([]string, 0, len(platformTargets))
	for platform := range platformTargets {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)

	m := manifest{
		Schema: 1,
		Header: artifact{
			Path: "internal/clib/nono.h",
		},
	}
	var err error
	m.Header.SHA256, err = fileSHA256(m.Header.Path)
	if err != nil {
		return manifest{}, err
	}

	for _, platform := range platforms {
		target := platformTargets[platform]
		libPath := filepath.ToSlash(filepath.Join("internal/clib", platform, "libnono_ffi.a"))
		libHash, err := fileSHA256(libPath)
		if err != nil {
			return manifest{}, err
		}
		m.Libs = append(m.Libs, libArtifact{
			Platform: platform,
			Target:   target,
			Path:     libPath,
			SHA256:   libHash,
		})

		versionPath := filepath.ToSlash(filepath.Join("internal/clib", platform, "VERSION"))
		versionHash, err := fileSHA256(versionPath)
		if err != nil {
			return manifest{}, err
		}
		commit, err := versionCommit(versionPath)
		if err != nil {
			return manifest{}, err
		}
		m.Versions = append(m.Versions, versionRecord{
			Platform: platform,
			Target:   target,
			Path:     versionPath,
			Commit:   commit,
			SHA256:   versionHash,
		})
	}
	return m, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func versionCommit(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) != 40 {
			return "", fmt.Errorf("%s: expected 40-character commit SHA, got %q", path, line)
		}
		return line, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s: missing commit SHA", path)
}

func readManifest(path string) (manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return manifest{}, err
	}
	defer f.Close()

	var m manifest
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return manifest{}, err
	}
	if m.Schema != 1 {
		return manifest{}, fmt.Errorf("%s: unsupported schema %d", path, m.Schema)
	}
	return m, nil
}

func writeManifest(path string, m manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func equalManifest(a, b manifest) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aj) == string(bj)
}

func diffManifest(want, got manifest) error {
	if want.Header != got.Header {
		return fmt.Errorf("header mismatch: manifest has %s, current is %s", want.Header.SHA256, got.Header.SHA256)
	}
	for i := range want.Libs {
		if i >= len(got.Libs) || want.Libs[i] != got.Libs[i] {
			return fmt.Errorf("library mismatch at index %d", i)
		}
	}
	if len(want.Libs) != len(got.Libs) {
		return fmt.Errorf("library count mismatch: manifest has %d, current is %d", len(want.Libs), len(got.Libs))
	}
	for i := range want.Versions {
		if i >= len(got.Versions) || want.Versions[i] != got.Versions[i] {
			return fmt.Errorf("version mismatch at index %d", i)
		}
	}
	if len(want.Versions) != len(got.Versions) {
		return fmt.Errorf("version count mismatch: manifest has %d, current is %d", len(want.Versions), len(got.Versions))
	}
	return errors.New("manifest mismatch")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
