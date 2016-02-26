package modules

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ModuleConsoleRunner is the ConsoleRunner module.
	ModuleConsoleRunner = "console_runner"

	// ActionConsoleRunnerRun is the ConsoleRunner search action.
	ActionConsoleRunnerRun = "consolerunner.run"

	// ConsoleRunnerTimeout is the timeout for any console calls.
	ConsoleRunnerTimeout = 10 * time.Second
)

// ConsoleRunner is the google image search for ConsoleRunner module
type ConsoleRunner struct{}

// Name returns the module name.
func (cr *ConsoleRunner) Name() string {
	return ModuleConsoleRunner
}

// Actions returns the module actions.
func (cr *ConsoleRunner) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionConsoleRunnerRun, MessagePattern: "^run", Description: "Runs a given console command", Handler: cr.handleConsoleRunnerRun},
	}
}

func (cr *ConsoleRunner) isWhitelistedCommand(cmd string) bool {
	switch cmd {
	case "traceroute", "whois", "ping", "w", "ps", "uptime":
		return true
	default:
		return false
	}
}

func (cr *ConsoleRunner) handleConsoleRunnerRun(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessSpecificMention(m.Text, b.ID()))
	cleanedMessage := core.FixLinks(messageWithoutMentions)

	pieces := strings.Split(cleanedMessage, " ")
	if len(pieces) < 2 {
		return exception.Newf("invalid arguments for `%s`", ActionConsoleRunnerRun)
	}

	commandWithArguments := pieces[1:]
	command := commandWithArguments[0]
	args := []string{}
	if len(commandWithArguments) > 1 {
		args = commandWithArguments[1:]
	}

	if !cr.isWhitelistedCommand(command) {
		return exception.Newf("`%s` cannot run %s", ActionConsoleRunnerRun, command)
	}

	cmdFullPath, lookErr := exec.LookPath(command)
	if lookErr != nil {
		return exception.Wrap(lookErr)
	}

	stdoutBuffer := bytes.NewBuffer([]byte{})
	subCmd := exec.Command(cmdFullPath, args...)
	subCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	subCmd.StdoutPipe()
	subCmd.StderrPipe()
	subCmd.Stdout = stdoutBuffer
	subCmd.Stderr = stdoutBuffer

	startErr := subCmd.Start()
	if startErr != nil {
		return startErr
	}

	started := time.Now().UTC()

	didTimeout := false
	go func() {
		for {
			now := time.Now().UTC()
			if now.Sub(started) > ConsoleRunnerTimeout {
				didTimeout = true
				pgid, err := syscall.Getpgid(subCmd.Process.Pid)
				if err != nil {
					return
				}
				syscall.Kill(-pgid, 15)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	subCmd.Wait()

	timedOutText := ""
	if didTimeout {
		timedOutText = " (timed out)"
	}

	stdout := stdoutBuffer.String()
	outputText := fmt.Sprintf("console runner stdout%s:\n", timedOutText)
	if len(stdout) != 0 {
		prefixed := strings.Replace(stdout, "\n", "\n>", -1)
		outputText = outputText + fmt.Sprintf(">%s", prefixed)
	} else {
		outputText = "> empty"
	}

	return b.Say(m.Channel, outputText)
}
