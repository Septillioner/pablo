package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	hostKeyVerificationOff = "off"
	trustOnFirstUseOn      = "on"
	defaultSSHPort         = "22"
	sshDirPerm             = 0o700
	knownHostsFilePerm     = 0o600
)

// HostKeyOptions controls how remote host keys are verified.
type HostKeyOptions struct {
	// Verification is "on" (default), "off", or empty (treated as on).
	Verification string
	// TrustOnFirstUse is "on", "off", or empty (treated as off).
	TrustOnFirstUse string
	// KnownHostsPath overrides the default ~/.ssh/known_hosts path when set.
	KnownHostsPath string
}

func (o HostKeyOptions) VerificationDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(o.Verification), hostKeyVerificationOff)
}

func (o HostKeyOptions) TrustOnFirstUseEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(o.TrustOnFirstUse), trustOnFirstUseOn)
}

func defaultKnownHostsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory for known_hosts: %w", err)
	}
	return filepath.Join(home, ".ssh", "known_hosts"), nil
}

func resolveKnownHostsPath(opts HostKeyOptions) (string, error) {
	if strings.TrimSpace(opts.KnownHostsPath) != "" {
		return expandPath(opts.KnownHostsPath), nil
	}
	return defaultKnownHostsPath()
}

func ensureKnownHostsFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat known_hosts %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), sshDirPerm); err != nil {
		return fmt.Errorf("create .ssh directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, knownHostsFilePerm)
	if err != nil {
		return fmt.Errorf("create known_hosts %s: %w", path, err)
	}
	return f.Close()
}

func buildHostKeyCallback(opts HostKeyOptions) (ssh.HostKeyCallback, error) {
	if opts.VerificationDisabled() {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	knownHostsPath, err := resolveKnownHostsPath(opts)
	if err != nil {
		return nil, err
	}
	if err := ensureKnownHostsFile(knownHostsPath); err != nil {
		return nil, err
	}

	baseCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts from %s: %w", knownHostsPath, err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := baseCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) {
			return err
		}

		fingerprint := ssh.FingerprintSHA256(key)

		// Want empty => host is unknown; Want non-empty => key mismatch.
		if len(keyErr.Want) > 0 {
			return fmt.Errorf(
				"SSH host key mismatch for %s (possible MITM).\n"+
					"Presented key fingerprint: %s\n"+
					"Remove the old entry from %s and add the new key only if you trust this host",
				hostname, fingerprint, knownHostsPath,
			)
		}

		if opts.TrustOnFirstUseEnabled() {
			if appendErr := appendKnownHost(knownHostsPath, hostname, remote, key); appendErr != nil {
				return fmt.Errorf("trust-on-first-use failed for %s: %w", hostname, appendErr)
			}
			return nil
		}

		hostForScan := hostnameForKeyScan(hostname)
		return fmt.Errorf(
			"SSH host key for %s is not in known_hosts (%s).\n"+
				"Presented key fingerprint: %s\n"+
				"Add the host key with:\n"+
				"  ssh-keyscan -H %s >> %s\n"+
				"Or connect once interactively:\n"+
				"  ssh %s\n"+
				"Optional: set remote.trust_on_first_use: on to record the key on first connect.\n"+
				"Not recommended: set remote.host_key_verification: off to skip verification",
			hostname, knownHostsPath, fingerprint, hostForScan, knownHostsPath, hostForScan,
		)
	}, nil
}

func appendKnownHost(path, hostname string, remote net.Addr, key ssh.PublicKey) error {
	addresses := knownHostAddresses(hostname, remote)
	line := knownhosts.Line(addresses, key)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, knownHostsFilePerm)
	if err != nil {
		return fmt.Errorf("open known_hosts for append: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, line); err != nil {
		return fmt.Errorf("write known_hosts entry: %w", err)
	}
	return nil
}

func knownHostAddresses(hostname string, remote net.Addr) []string {
	normalized := knownhosts.Normalize(hostname)
	addrs := []string{normalized}
	if remote != nil {
		remoteNormalized := knownhosts.Normalize(remote.String())
		if remoteNormalized != "" && remoteNormalized != normalized {
			addrs = append(addrs, remoteNormalized)
		}
	}
	return addrs
}

func hostnameForKeyScan(hostname string) string {
	normalized := knownhosts.Normalize(hostname)
	host, port, err := net.SplitHostPort(normalized)
	if err != nil {
		return strings.Trim(normalized, "[]")
	}
	if port == "" || port == defaultSSHPort {
		return host
	}
	return net.JoinHostPort(host, port)
}
