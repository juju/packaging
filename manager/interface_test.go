// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager_test

import (
	"github.com/juju/packaging/manager"
)

var _ manager.PackageManager = manager.NewAptPackageManager()
var _ manager.PackageManager = manager.NewYumPackageManager()
var _ manager.PackageManager = manager.NewZypperPackageManager()
