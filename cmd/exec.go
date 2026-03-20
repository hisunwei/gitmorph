package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:                "exec [git args...]",
	Short:              "Auto-switch profile and execute git command",
	Long:               `Detect .gitmorph config in the current git repository, switch to the specified profile, and execute the git command. Reads DEFAULT_GITMORPH_PROFILE and ORIGINAL_GIT from environment variables.`,
	DisableFlagParsing: true,
	SilenceUsage:       true,
	Run:                runExec,
}

func init() {
	RootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) {
	// Resolve the original git binary path from env, fallback to PATH lookup
	gitBin := os.Getenv("ORIGINAL_GIT")
	if gitBin == "" {
		var err error
		gitBin, err = exec.LookPath("git")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: could not find git binary")
			os.Exit(1)
		}
	}

	defaultProfile := os.Getenv("DEFAULT_GITMORPH_PROFILE")

	// Check if we're inside a git repository
	repoRoot, err := gitRevParseRoot(gitBin)
	if err == nil && repoRoot != "" {
		profileName := ""
		gitmorphFile := filepath.Join(repoRoot, ".gitmorph")
		data, readErr := os.ReadFile(gitmorphFile)
		if readErr == nil {
			profileName = strings.TrimSpace(string(data))
		}

		// Fallback to default profile if no profile determined
		if profileName == "" && defaultProfile != "" {
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "No .gitmorph file found, using default profile: %s\n", defaultProfile)
			} else {
				fmt.Fprintf(os.Stderr, ".gitmorph file is empty, using default profile: %s\n", defaultProfile)
			}
			profileName = defaultProfile
		}

		if profileName != "" {
			if switchErr := autoSwitchProfile(profileName, gitBin); switchErr != nil {
				fmt.Fprintf(os.Stderr, "\033[31m==========Gitmorph Error============\033[0m\n")
				fmt.Fprintf(os.Stderr, "\033[31mError: Failed to switch to gitmorph profile: %s\033[0m\n", profileName)
				fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", switchErr.Error())
				fmt.Fprintf(os.Stderr, "\033[31mPlease ensure the profile exists and the SSH key is correctly configured.\033[0m\n")
				fmt.Fprintf(os.Stderr, "\033[31mGit command aborted.\033[0m\n")
				fmt.Fprintf(os.Stderr, "\033[31m==========End Gitmorph Error============\033[0m\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Using gitmorph Profile: \033[31m%s\033[0m\n", profileName)
		}
	}

	// Replace current process with the real git binary
	execArgs := append([]string{"git"}, args...)
	if execErr := syscall.Exec(gitBin, execArgs, os.Environ()); execErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to execute git: %v\n", execErr)
		os.Exit(1)
	}
}

func gitRevParseRoot(gitBin string) (string, error) {
	out, err := exec.Command(gitBin, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func autoSwitchProfile(name string, gitBin string) error {
	profile, exists := profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	if err := setGlobalGitConfig(gitBin, "user.name", profile.Username); err != nil {
		return fmt.Errorf("error setting user.name: %w", err)
	}
	if err := setGlobalGitConfig(gitBin, "user.email", profile.Email); err != nil {
		return fmt.Errorf("error setting user.email: %w", err)
	}

	if profile.SSHKey != "" {
		sshCmd := fmt.Sprintf("ssh -i %s", profile.SSHKey)
		if err := setGlobalGitConfig(gitBin, "core.sshCommand", sshCmd); err != nil {
			return fmt.Errorf("error setting core.sshCommand: %w", err)
		}
	} else {
		// Remove custom SSH command if no key specified
		exec.Command(gitBin, "config", "--global", "--unset", "core.sshCommand").Run()
	}

	return nil
}

func setGlobalGitConfig(gitBin, key, value string) error {
	return exec.Command(gitBin, "config", "--global", key, value).Run()
}
