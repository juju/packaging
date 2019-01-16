// Copyright 2019 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package config

import (
	"github.com/juju/packaging"
)

// snapConfigurer is the PackagingConfigurer implementation for snap-based systems.
type snapConfigurer struct {
	*baseConfigurer
}

// RenderSource is defined on the PackagingConfigurer interface.
func (c *snapConfigurer) RenderSource(src packaging.PackageSource) (string, error) {
	return "", nil
}

// RenderPreferences is defined on the PackagingConfigurer interface.
func (c *snapConfigurer) RenderPreferences(prefs packaging.PackagePreferences) (string, error) {
	return "", nil
}

// ApplyCloudArchiveTarget is defined on the PackagingConfigurer interface.
func (c *snapConfigurer) ApplyCloudArchiveTarget(pack string) []string {
	return []string{pack}
}
