// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package config_test

import (
	"github.com/juju/packaging/v3/config"
)

var _ config.PackagingConfigurer = config.NewAptPackagingConfigurer()
var _ config.PackagingConfigurer = config.NewYumPackagingConfigurer()
var _ config.PackagingConfigurer = config.NewZypperPackagingConfigurer()
