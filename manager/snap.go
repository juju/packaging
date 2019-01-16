// Copyright 2019 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/proxy"
)

// snap is the PackageManager implementation for snap-based systems.
type snap struct {
	basePackageManager
}

var snapMissingConfig = regexp.MustCompile(`error: snap "[a-z-]+" has no "[a-z-\.]+" configuration option`)

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

	err := snap.readProxySettingsFromSnapSystemConfig(&res)
	if err != nil {
		return proxy.Settings{}, err
	}

	return res, nil
}

func (snap *snap) readProxySettingsFromSnapSystemConfig(settings *proxy.Settings) error {
	configKeys := []string{"http", "https", "ftp"}

	for _, proxyType := range configKeys {
		// for snap <2.3.6, this should be "core"
		cmd := fmt.Sprintf("snap get system proxy.%s", proxyType)
		out, _, err := RunCommandWithRetry(cmd, nil)
		if snapMissingConfig.MatchString(out) {
			continue
		}
		if err != nil {
			return err
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

// InstallPrerequisite is defined on the PackageManager interface.
//
// It is a no-op for snaps.
func (snap *snap) InstallPrerequisite() error {
	return nil
}

// Update is defined on the PackageManager interface.
func (snap *snap) Update() error {
	_, _, err := RunCommandWithRetry(snap.cmder.UpdateCmd(), nil)
	return err
}

// Upgrade is defined on the PackageManager interface.
func (snap *snap) Upgrade() error {
	_, _, err := RunCommandWithRetry(snap.cmder.UpgradeCmd(), nil)
	return err
}

// Remove is defined on the PackageManager interface.
func (snap *snap) Remove(packs ...string) error {
	_, _, err := RunCommandWithRetry(snap.cmder.RemoveCmd(packs...), nil)
	return err
}

// Purge is defined on the PackageManager interface.
func (snap *snap) Purge(packs ...string) error {
	_, _, err := RunCommandWithRetry(snap.cmder.PurgeCmd(packs...), nil)
	return err
}

// IsInstalled is defined on the PackageManager interface.
func (snap *snap) IsInstalled(pack string) bool {
	pack = strings.ToLower(strings.TrimSpace(pack))
	cmd := snap.cmder.IsInstalledCmd(pack)
	cmd = fmt.Sprintf(cmd, pack)
	stdOut, err := RunCommand(cmd)
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(stdOut)) == pack
}

// AddRepository is defined on the PackageManager interface.
func (snap *snap) AddRepository(repo string) error {
	return nil
}

// RemoveRepository is defined on the PackageManager interface.
func (snap *snap) RemoveRepository(repo string) error {
	return nil
}

// Cleanup is defined on the PackageManager interface.
func (snap *snap) Cleanup() error {
	return nil
}

// SetProxy is defined on the PackageManager interface.
func (snap *snap) SetProxy(settings proxy.Settings) error {
	return snap.setProxySettingsInSnapSystemConfig(&settings)
}

func (snap *snap) setProxySettingsInSnapSystemConfig(settings *proxy.Settings) error {
	proxyConfig := map[string]string{
		"http":  settings.Http,
		"https": settings.Https,
		"ftp":   settings.Ftp,
	}

	for proxyType, proxyValue := range proxyConfig {
		if proxyValue == "" {
			continue
		}

		cmd := fmt.Sprintf("snap set system proxy.%s %s", proxyType, proxyValue)
		_, _, err := RunCommandWithRetry(cmd, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
