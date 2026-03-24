package main

import (
	"runtime/debug"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/release"
)

// botVersion can be overridden at build time:
// go build -ldflags "-X main.botVersion=v0.9.1-beta2" ./cmd/bot
var botVersion = release.Current

func init() {
	botVersion = resolveBotVersion(botVersion)
}

func resolveBotVersion(current string) string {
	current = strings.TrimSpace(current)
	if current != "" && current != "dev" {
		return current
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	if v := strings.TrimSpace(buildInfo.Main.Version); v != "" && v != "(devel)" {
		return v
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key != "vcs.revision" {
			continue
		}
		revision := strings.TrimSpace(setting.Value)
		if revision == "" {
			break
		}
		if len(revision) > 7 {
			revision = revision[:7]
		}
		return "dev-" + revision
	}

	return "dev"
}
