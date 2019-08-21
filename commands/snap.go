// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package commands

const (
	// snap binary name.
	snapBinary = "snap"

	snapNoProxySettingFormat = `proxy.%s=""`
	snapProxySettingFormat   = `proxy.%s=%q`
)

// NewSnapPackageCommander returns a PackageCommander for snap-based systems.
func NewSnapPackageCommander(channel, confinement string) PackageCommander {
	var installArgs string
	if channel != "" {
		installArgs += " --" + channel
	}
	if confinement != "" {
		installArgs += " --" + confinement
	}

	return &packageCommander{
		prereq:           makeNopCmd(),
		update:           makeNopCmd(),
		upgrade:          buildCommand(snapBinary, "refresh"),
		install:          buildCommand(snapBinary, "install "+installArgs),
		remove:           buildCommand(snapBinary, "remove"),
		purge:            buildCommand(snapBinary, "remove"),
		search:           buildCommand(snapBinary, "info %s"),
		isInstalled:      buildCommand(snapBinary, "list %s"),
		listAvailable:    makeNopCmd(),
		listInstalled:    buildCommand(snapBinary, "list"),
		addRepository:    makeNopCmd(),
		listRepositories: makeNopCmd(),
		removeRepository: makeNopCmd(),
		cleanup:          makeNopCmd(),
		// Note: proxy.{http,https} available since snapd 2.28
		getProxy:              buildCommand(snapBinary, "get system proxy"),
		proxySettingsFormat:   snapProxySettingFormat,
		noProxySettingsFormat: snapNoProxySettingFormat,
		setProxy:              buildCommand(snapBinary, "set system %s"),
	}
}

func makeNopCmd() string {
	return buildCommand(":", "#No action here")
}
