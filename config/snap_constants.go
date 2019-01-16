// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package config

// SnapDefaultPackages is a slice of snaps that are installed
// by default.
var SnapDefaultPackages = []string{
	"juju-db",
}

var cloudArchivePackagesSnap = map[string]struct{}{}
