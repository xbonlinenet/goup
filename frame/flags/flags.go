package flags

import (
	"fmt"
	"strings"
)

var (
	// Go flags
	GoVersion string

	// 	Git flags
	GitBranch string
	GitCommit string

	// Build flags
	BuildTime string
	BuildEnv  string
)

// flags=" \
// -X 'github.com/xbonlinenet/goup/frame/flags.GoVersion=$(go version)' \
// -X github.com/xbonlinenet/goup/frame/flags.GitBranch=`git rev-parse --abbrev-ref HEAD` \
// -X github.com/xbonlinenet/goup/frame/flags.GitCommit=`git rev-parse HEAD` \
// -X github.com/xbonlinenet/goup/frame/flags.BuildTime=`date -u '+%Y-%m-%d %H:%M:%S'` \
// -X github.com/xbonlinenet/goup/frame/flags.BuildEnv=$env \
// "
// go build -ldflags "$flags" -x -o main main.go

func init() {
	// default flag
	if BuildEnv == "" {
		BuildEnv = "dev"
	}
}

func DisplayCompileTimeFlags() {
	var sb strings.Builder

	sb.WriteString("[CompileTimeFlags]\n")
	sb.WriteString(fmt.Sprintf("GoVersion: %s\n", GoVersion))
	sb.WriteString(fmt.Sprintf("GitBranch: %s\n", GitBranch))
	sb.WriteString(fmt.Sprintf("GitCommit: %s\n", GitCommit))
	sb.WriteString(fmt.Sprintf("BuildEnv: %s\n", BuildEnv))
	sb.WriteString(fmt.Sprintf("BuildTime: %s\n", BuildTime))

	fmt.Println(sb.String())
}
