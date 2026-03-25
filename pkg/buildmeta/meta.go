package buildmeta

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// These can be injected at build time when VCS metadata is unavailable:
// -ldflags "-X github.com/mr-cheeezz/dankbot/pkg/buildmeta.BuildBranch=stable \
//          -X github.com/mr-cheeezz/dankbot/pkg/buildmeta.BuildRevision=abc1234 \
//          -X github.com/mr-cheeezz/dankbot/pkg/buildmeta.BuildCommitTime=2026-03-24T20:10:00Z"
var (
	BuildBranch     = ""
	BuildRevision   = ""
	BuildCommitTime = ""
)

type Info struct {
	Version      string
	Branch       string
	Revision     string
	CommitTime   string
	CommitShort  string
	CommitSource string
}

func Detect(version string) Info {
	info := Info{
		Version: strings.TrimSpace(version),
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		if info.Version == "" {
			info.Version = "dev"
		}
		return info
	}

	settings := map[string]string{}
	for _, setting := range buildInfo.Settings {
		settings[setting.Key] = strings.TrimSpace(setting.Value)
	}

	revision := strings.TrimSpace(settings["vcs.revision"])
	branch := strings.TrimSpace(settings["vcs.branch"])
	commitTime := strings.TrimSpace(settings["vcs.time"])

	if branch == "" || revision == "" {
		gitBranch, gitRevision := detectFromGitDir()
		if branch == "" {
			branch = gitBranch
		}
		if revision == "" {
			revision = gitRevision
		}
	}

	if branch == "" {
		branch = strings.TrimSpace(BuildBranch)
	}
	if revision == "" {
		revision = strings.TrimSpace(BuildRevision)
	}
	if commitTime == "" {
		commitTime = strings.TrimSpace(BuildCommitTime)
	}

	if info.Version == "" || info.Version == "dev" {
		mainVersion := strings.TrimSpace(buildInfo.Main.Version)
		if mainVersion != "" && mainVersion != "(devel)" {
			info.Version = mainVersion
		}
	}
	if info.Version == "" {
		if revision != "" {
			info.Version = "dev-" + shortRevision(revision)
		} else {
			info.Version = "dev"
		}
	}

	info.Branch = branch
	info.Revision = revision
	info.CommitShort = shortRevision(revision)
	info.CommitTime = normalizeCommitTime(commitTime)

	if strings.EqualFold(strings.TrimSpace(settings["vcs.modified"]), "true") {
		info.CommitSource = "dirty"
	} else {
		info.CommitSource = "clean"
	}

	return info
}

func shortRevision(revision string) string {
	revision = strings.TrimSpace(revision)
	if len(revision) > 8 {
		return revision[:8]
	}
	return revision
}

func normalizeCommitTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}

	return parsed.UTC().Format(time.RFC3339)
}

func detectFromGitDir() (string, string) {
	execPath, err := os.Executable()
	if err != nil {
		return "", ""
	}

	dir := filepath.Dir(execPath)
	root, gitDir := findGitDir(dir)
	if gitDir == "" {
		return "", ""
	}

	headBytes, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", ""
	}
	head := strings.TrimSpace(string(headBytes))
	if head == "" {
		return "", ""
	}

	if !strings.HasPrefix(head, "ref: ") {
		return "", strings.TrimSpace(head)
	}

	ref := strings.TrimSpace(strings.TrimPrefix(head, "ref: "))
	branch := ref
	const headPrefix = "refs/heads/"
	if strings.HasPrefix(ref, headPrefix) {
		branch = strings.TrimPrefix(ref, headPrefix)
	}

	revision := resolveGitRef(root, gitDir, ref)
	return branch, revision
}

func findGitDir(start string) (string, string) {
	current := filepath.Clean(start)
	for {
		candidate := filepath.Join(current, ".git")
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				return current, candidate
			}
			if !info.IsDir() {
				content, readErr := os.ReadFile(candidate)
				if readErr == nil {
					line := strings.TrimSpace(string(content))
					if strings.HasPrefix(line, "gitdir:") {
						path := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
						if !filepath.IsAbs(path) {
							path = filepath.Clean(filepath.Join(current, path))
						}
						return current, path
					}
				}
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", ""
		}
		current = parent
	}
}

func resolveGitRef(root, gitDir, ref string) string {
	if ref == "" {
		return ""
	}

	refPath := filepath.Clean(filepath.Join(gitDir, ref))
	if strings.HasPrefix(refPath, gitDir+string(os.PathSeparator)) || refPath == filepath.Clean(gitDir) {
		if bytes, err := os.ReadFile(refPath); err == nil {
			return strings.TrimSpace(string(bytes))
		}
	}

	packedRefsPath := filepath.Join(gitDir, "packed-refs")
	packedRefs, err := os.ReadFile(packedRefsPath)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(packedRefs), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "^") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if parts[1] == ref {
			return strings.TrimSpace(parts[0])
		}
	}

	_ = root
	return ""
}
