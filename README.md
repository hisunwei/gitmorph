# GitMorph

GitMorph is a powerful CLI tool that allows you to seamlessly switch between multiple Git identities on your local machine. Perfect for developers who work on different projects with various Git accounts.

<img width="826" alt="Screenshot 2025-05-25 at 9 59 25 AM" src="https://github.com/user-attachments/assets/c0801555-546a-4b69-a9b0-508d0b9c60ad" />

## Features

- Create and manage multiple Git profiles
- Easily switch between different Git identities (incl. per-profile SSH key)
- List all available profiles (shows SSH key path)
- Edit existing profiles
- Delete profiles
- Simple and intuitive command-line interface

## Installation

To install GitMorph, make sure you have Go installed on your system, then run:

```bash
go install github.com/abhigyan-mohanta/gitmorph@latest
````

### Update PATH

After installation, you may need to add the Go binaries directory to your system's `PATH` so you can run `gitmorph` from anywhere:

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

## Usage

GitMorph provides the following commands:

### Create a new profile

```bash
gitmorph new
```

Prompts for:

* **Profile name**
* **Git username**
* **Git email**
* **SSH private key path** (leave blank for `~/.ssh/id_ed25519`)

### List all profiles

```bash
gitmorph list
```

Shows:

```
Available Git profiles:
- work     (Username: alice, Email: alice@corp.com, SSH: ~/.ssh/id_ed25519_work)
- personal (Username: alice123, Email: alice@gmail.com, SSH: ~/.ssh/id_ed25519)
```

### Switch to a profile

```bash
gitmorph switch <profile-name>
```

* Sets `user.name` and `user.email` globally
* Sets or unsets `core.sshCommand` to use the profile’s SSH key

### Edit a profile

```bash
gitmorph edit <profile-name>
```

Interactively update any of:

* Username
* Email
* SSH key path

Leave a prompt blank to keep the current value.

### Delete a profile

```bash
gitmorph delete <profile-name>
```

Removes the profile entry from `~/.gitmorph.json`.

## How It Works

GitMorph stores your Git profiles in a JSON file located at `~/.gitmorph.json`. Commands:

* `new` / `edit` / `delete` modify the JSON
* `switch` updates your global Git config:

   * `git config --global user.name <username>`
   * `git config --global user.email <email>`
   * `git config --global core.sshCommand "ssh -i <sshKey>"`

If you delete or switch back to a profile with no custom key, it unsets `core.sshCommand`.

## Code Structure

* `main.go`: Entry point
* `cmd/root.go`: Root command, loading/saving JSON
* `cmd/new.go`: `new` command
* `cmd/list.go`: `list` command
* `cmd/switch.go`: `switch` command
* `cmd/edit.go`: `edit` command
* `cmd/delete.go`: `delete` command

## SSH Configuration

You can still use `~/.ssh/config` if you like; GitMorph’s per-profile `core.sshCommand` will override it when set.

Example `~/.ssh/config`:

```plaintext
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519

Host github.com-work
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_work
```

## Auto switch by .gitmorph config

1.  Add `.gitmorph` in root directory of git repo:
    ```
    echo 'personal' > .gitmorph
    ```
2. Add the following to your shell profile (`.bashrc`/`.zshrc` or `.profile`), replacing the `DEFAULT_GITMORPH_PROFILE` value as needed:

```bash
# Define the default profile
export DEFAULT_GITMORPH_PROFILE="work"

# Get the path of the original git command
export ORIGINAL_GIT=$(command -v git)

# Override git to auto-switch gitmorph profile
function git() {
    gitmorph exec "$@"
}
```

The `gitmorph exec` command handles all the auto-switch logic:
- Detects the git repository root
- Reads the `.gitmorph` file for the target profile name
- Falls back to `DEFAULT_GITMORPH_PROFILE` if no `.gitmorph` file or if it's empty
- Switches to the profile (sets `user.name`, `user.email`, `core.sshCommand`)
- Aborts the git command with an error if the profile switch fails
- Executes the real git command with all arguments

## Dependencies

```go
require (
    github.com/spf13/cobra v1.8.1
    github.com/spf13/pflag v1.0.5
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
