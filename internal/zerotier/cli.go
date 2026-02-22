package zerotier

import "os"

// known install locations for zerotier-cli across macOS and Linux
var cliPaths = []string{
	"/usr/local/bin/zerotier-cli",
	"/usr/sbin/zerotier-cli",
	"/Library/Application Support/ZeroTier/One/zerotier-cli",
}

// IsInstalled returns true if a zerotier-cli binary is found at a known path.
func IsInstalled() bool {
	return CLIPath() != ""
}

// CLIPath returns the path to the zerotier-cli binary, or empty string if not found.
func CLIPath() string {
	for _, p := range cliPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
