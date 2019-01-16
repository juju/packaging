// Copyright 2019 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/proxy"
)

// snap is the PackageManager implementation for snap-based systems.
type snap struct {
	basePackageManager
}

var snapMissingConfig = regexp.MustCompile(`error: snap "[a-z-]+" has no "[a-z-]+" configuration option`)

// Search is defined on the PackageManager interface.
func (snap *snap) Search(pack string) (bool, error) {
	out, _, err := RunCommandWithRetry(snap.cmder.SearchCmd(pack), nil)
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(out, "No matching snaps") {
		return false, nil
	}
	return true, nil
}

// Install is defined on the PackageManager interface.
func (snap *snap) Install(packs ...string) error {
	fatalErr := func(output string) error {
		if strings.Contains(output, "not found") {
			return errors.New("unable to locate package") // error message from apt
		}
		return nil
	}
	_, _, err := RunCommandWithRetry(snap.cmder.InstallCmd(packs...), fatalErr)
	return err
}

// GetProxySettings is defined on the PackageManager interface.
func (snap *snap) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings

	if res == (proxy.Settings{}) {
		err := snap.readEtcEnvironmentForProxySettings(&res)
		if err != nil {
			return proxy.Settings{}, err
		}
	}

	return res, nil
}

func (snap *snap) readProxySettingsFromSnapSystemConfig(settings *proxy.Settings) error {
	configKeys := []string{"http", "https", "ftp"}

	for _, proxyType := range configKeys {
		// for snap <2.3.6, this should be "core"
		cmd := fmt.Sprintf("snap get system proxy.%s", proxyType)
		out, _, err := RunCommandWithRetry(cmd, nil)
		if err != nil {
			return err
		}

		if snapMissingConfig.MatchString(out) {
			continue
		}

		switch proxyType {
		case "http":
			settings.Http = out
		case "https":
			settings.Https = out
		case "ftp":
			settings.Ftp = out
		default:
			return errors.New("unsupported proxy setting")
		}
	}

	return nil
}

func (snap *snap) readEtcEnvironmentForProxySettings(settings *proxy.Settings) error {
	envVars := []string{"http", "https", "ftp", "no"}

	for _, proxyType := range envVars {
		read := fmt.Sprintf(`grep -is ^%s_proxy= /etc/environment || echo ""`, proxyType)
		format := fmt.Sprintf(`| sed 's/^%s_proxy=//I' | sed 's/"//g`, proxyType)

		out, _, err := RunCommandWithRetry(read+" | "+format, nil)
		if err != nil {
			return err
		}

		if snapMissingConfig.MatchString(out) {
			continue
		}

		switch proxyType {
		case "http":
			settings.Http = out
		case "https":
			settings.Https = out
		case "ftp":
			settings.Ftp = out
		case "no":
			settings.NoProxy = out
		default:
			return errors.New("unsupported proxy setting")
		}
	}
	return nil
}

// InstallPrerequisite is defined on the PackageManager interface.
//
// It is a no-op for snaps.
func (snap *basePackageManager) InstallPrerequisite() error {
	return nil
}

// Update is defined on the PackageManager interface.
func (snap *basePackageManager) Update() error {
	_, _, err := RunCommandWithRetry(pm.cmder.UpdateCmd(), nil)
	return err
}

// Upgrade is defined on the PackageManager interface.
func (snap *basePackageManager) Upgrade() error {
	_, _, err := RunCommandWithRetry(snap.cmder.UpgradeCmd(), nil)
	return err
}

// Install is defined on the PackageManager interface.
func (snap *basePackageManager) Install(packs ...string) error {
	_, _, err := RunCommandWithRetry(snap.cmder.InstallCmd(packs...), nil)
	return err
}

// Remove is defined on the PackageManager interface.
func (snap *basePackageManager) Remove(packs ...string) error {
	_, _, err := RunCommandWithRetry(snap.cmder.RemoveCmd(packs...), nil)
	return err
}

// Purge is defined on the PackageManager interface.
func (snap *basePackageManager) Purge(packs ...string) error {
	_, _, err := RunCommandWithRetry(snap.cmder.PurgeCmd(packs...), nil)
	return err
}

// IsInstalled is defined on the PackageManager interface.
func (snap *basePackageManager) IsInstalled(pack string) bool {
	args := strings.Fields(snap.cmder.IsInstalledCmd(pack))

	_, err := RunCommand(args[0], args[1:]...)
	return err == nil
}

// AddRepository is defined on the PackageManager interface.
func (snap *basePackageManager) AddRepository(repo string) error {
	return errors.New("not implemented")
}

// RemoveRepository is defined on the PackageManager interface.
func (snap *basePackageManager) RemoveRepository(repo string) error {
	return errors.New("not implemented")
}

// Cleanup is defined on the PackageManager interface.
func (snap *basePackageManager) Cleanup() error {
	return nil
}

// SetProxy is defined on the PackageManager interface.
func (snap *basePackageManager) SetProxy(settings proxy.Settings) error {
	cmds := pm.cmder.SetProxyCmds(settings)

	for _, cmd := range cmds {
		args := []string{"bash", "-c", fmt.Sprintf("%q", cmd)}
		out, err := RunCommand(args[0], args[1:]...)
		if err != nil {
			logger.Errorf("command failed: %v\nargs: %#v\n%s", err, args, string(out))
			return fmt.Errorf("command failed: %v", err)
		}
	}

	return nil
}
