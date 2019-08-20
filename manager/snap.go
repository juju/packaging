// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package manager

import (
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/proxy"
)

var (
	// snapProxyRe is a regexp which matches all proxy-related configuration
	// options in the snap proxy settings output
	snapProxyRE = regexp.MustCompile(`(?im)^proxy\.(?P<protocol>[a-z]+)\s+(?P<proxy>.+)$`)

	snapNotFoundRE = regexp.MustCompile(`(?i)error: snap "[^"]+" not found`)
)

// snap is the PackageManager implementation for snap-based systems.
type snap struct {
	basePackageManager
}

// Search is defined on the PackageManager interface.
func (snap *snap) Search(pack string) (bool, error) {
	out, _, err := RunCommandWithRetry(snap.cmder.SearchCmd(pack), nil)
	if strings.Contains(combinedOutput(out, err), "error: no snap found") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// IsInstalled is defined on the PackageManager interface.
func (snap *snap) IsInstalled(pack string) bool {
	out, _, err := RunCommandWithRetry(snap.cmder.IsInstalledCmd(pack), nil)
	if strings.Contains(combinedOutput(out, err), "error: no matching snaps installed") || err != nil {
		return false
	}
	return true
}

// Install is defined on the PackageManager interface.
func (snap *snap) Install(packs ...string) error {
	out, _, err := RunCommandWithRetry(snap.cmder.InstallCmd(packs...), nil)
	if snapNotFoundRE.MatchString(combinedOutput(out, err)) {
		return errors.New("unable to locate package")
	}
	return err
}

// GetProxySettings is defined on the PackageManager interface.
func (snap *snap) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings

	out, _, err := RunCommandWithRetry(snap.cmder.GetProxyCmd(), nil)
	if strings.Contains(combinedOutput(out, err), `no "proxy" configuration option`) {
		return res, nil
	} else if err != nil {
		return res, err
	}

	for _, match := range snapProxyRE.FindAllStringSubmatch(out, -1) {
		switch match[1] {
		case "http":
			res.Http = match[2]
		case "https":
			res.Https = match[2]
		case "ftp":
			res.Ftp = match[2]
		}
	}

	return res, nil
}

func combinedOutput(out string, err error) string {
	res := string(out)
	if err != nil {
		res += err.Error()
	}
	return res
}
