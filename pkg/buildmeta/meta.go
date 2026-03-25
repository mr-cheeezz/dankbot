package buildmeta

import (
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
