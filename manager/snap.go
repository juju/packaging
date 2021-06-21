// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/packaging/v2/commands"
	"github.com/juju/proxy"
)

var (
	// snapProxyRe is a regexp which matches all proxy-related configuration
	// options in the snap proxy settings output
	snapProxyRE = regexp.MustCompile(`(?im)^proxy\.(?P<protocol>[a-z]+)\s+(?P<proxy>.+)$`)

	snapNotFoundRE = regexp.MustCompile(`(?i)error: snap "[^"]+" not found`)
	trackingRE     = regexp.MustCompile(`(?im)tracking:\s*(.*)$`)

	_ PackageManager = (*Snap)(nil)
)

// Snap is the PackageManager implementation for snap-based systems.
type Snap struct {
	basePackageManager
	installRetryable Retryable
}

// NewSnapPackageManager returns a PackageManager for snap-based systems.
func NewSnapPackageManager() *Snap {
	return &Snap{
		basePackageManager: basePackageManager{
			cmder:     commands.NewSnapPackageCommander(),
			retryable: dnsRetryable{},
		},
		installRetryable: snapInstallRetryable{},
	}
}

// Search is defined on the PackageManager interface.
func (snap *Snap) Search(pack string) (bool, error) {
	out, _, err := RunCommandWithRetry(snap.cmder.SearchCmd(pack), snap.retryable)
	if strings.Contains(combinedOutput(out, err), "error: no snap found") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// IsInstalled is defined on the PackageManager interface.
func (snap *Snap) IsInstalled(pack string) bool {
	out, _, err := RunCommandWithRetry(snap.cmder.IsInstalledCmd(pack), snap.retryable)
	if strings.Contains(combinedOutput(out, err), "error: no matching snaps installed") || err != nil {
		return false
	}
	return true
}

// InstalledChannel returns the snap channel for an installed package.
func (snap *Snap) InstalledChannel(pack string) string {
	out, _, err := RunCommandWithRetry(fmt.Sprintf("snap info %s", pack), snap.retryable)
	combined := combinedOutput(out, err)
	matches := trackingRE.FindAllStringSubmatch(combined, 1)
	if len(matches) == 0 {
		return ""
	}

	return matches[0][1]
}

// ChangeChannel updates the tracked channel for an installed snap.
func (snap *Snap) ChangeChannel(pack, channel string) error {
	out, _, err := RunCommandWithRetry(fmt.Sprintf("snap refresh --channel %s %s", channel, pack), snap.retryable)
	if err != nil {
		return err
	} else if strings.Contains(combinedOutput(out, err), "not installed") {
		return errors.Errorf("snap not installed")
	}

	return nil
}

// Install is defined on the PackageManager interface.
func (snap *Snap) Install(packs ...string) error {
	out, _, err := RunCommandWithRetry(snap.cmder.InstallCmd(packs...), snap.installRetryable)
	if snapNotFoundRE.MatchString(combinedOutput(out, err)) {
		return errors.New("unable to locate package")
	}
	return err
}

// GetProxySettings is defined on the PackageManager interface.
func (snap *Snap) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings

	out, _, err := RunCommandWithRetry(snap.cmder.GetProxyCmd(), snap.retryable)
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

// ConfigureStoreProxy sets up snapd to connect to the snap store proxy
// instance defined in the provided assertions using the provided store ID.
//
// If snap also needs to use HTTP/HTTPS proxies to talk to the outside world,
// these need to be configured separately before invoking this method via a
// call to SetProxy.
func (snap *Snap) ConfigureStoreProxy(assertions, storeID string) error {
	// Setup proxy based on the instructions from:
	// https://docs.ubuntu.com/snap-store-proxy/en/devices
	//
	// Note that while the above instructions run "snap ack /dev/stdin" the
	// code below will instead write the assertions to a temp file and pass
	// that to snap ack. This is purely done to make testing easier.
	assertFile, err := ioutil.TempFile("", "assertions")
	if err != nil {
		return errors.Annotate(err, "unable to create assertion file")
	}
	defer func() {
		_ = assertFile.Close()
		_ = os.Remove(assertFile.Name())
	}()
	if _, err = assertFile.WriteString(assertions); err != nil {
		return errors.Annotate(err, "unable to write to assertion file")
	}
	_ = assertFile.Close()

	ackCmd := fmt.Sprintf("snap ack %s", assertFile.Name())
	if _, _, err = RunCommandWithRetry(ackCmd, snap.retryable); err != nil {
		return errors.Annotate(err, "failed to execute 'snap ack'")
	}

	setCmd := fmt.Sprintf("snap set system proxy.store=%s", storeID)
	if _, _, err = RunCommandWithRetry(setCmd, snap.retryable); err != nil {
		return errors.Annotatef(err, "failed to configure snap to use store ID %q", storeID)
	}

	return nil
}

// DisableStoreProxy resets the snapd proxy store settings.
//
// If snap was also configured to use HTTP/HTTPS proxies these must be reset
// separately via a call to SetProxy.
// call to SetProxy.
func (snap *Snap) DisableStoreProxy() error {
	setCmd := "snap set system proxy.store="
	if _, _, err := RunCommandWithRetry(setCmd, snap.retryable); err != nil {
		return errors.Annotate(err, "failed to configure snap to not use a store proxy")
	}

	return nil
}

func combinedOutput(out string, err error) string {
	res := string(out)
	if err != nil {
		res += err.Error()
	}
	return res
}

type snapInstallRetryable struct{}

func (snapInstallRetryable) IsRetryable(err error, output string) (bool, int, error) {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return false, -1, errors.Annotatef(err, "unexpected error type %T", err)
	}
	waitStatus, ok := ProcessStateSys(exitError.ProcessState).(exitStatuser)
	if !ok {
		return false, -1, errors.Annotatef(err, "unexpected process state type %T", exitError.ProcessState.Sys())
	}

	code := waitStatus.ExitStatus()
	switch code {
	case 0:
		return false, code, nil
	case 1, 100:
		return true, code, errors.Trace(err)
	default:
		return false, code, errors.Trace(err)
	}
}
