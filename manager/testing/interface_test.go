// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"github.com/juju/packaging/v2/manager"
	"github.com/juju/packaging/v2/manager/testing"
)

var _ manager.PackageManager = &testing.MockPackageManager{}
