// Copyright 2019 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package commands

// snapCmder provides commands that are relevant for snap-based systems.
var snapCmder = packageCommander{
	update:      "snap refresh",
	upgrade:     "snap refresh",
	install:     "snap install",
	remove:      "snap remove",
	purge:       "snap remove",
	search:      "snap find %s",
	isInstalled: `snap list | grep "^%s"`,
}
