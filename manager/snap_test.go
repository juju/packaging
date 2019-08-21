// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package manager_test

import (
	"os"
	"os/exec"
	"strings"

	"github.com/juju/proxy"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/packaging/commands"
	"github.com/juju/packaging/manager"
)

var _ = gc.Suite(&SnapSuite{})

type SnapSuite struct {
	testing.IsolationSuite
	paccmder commands.PackageCommander
	pacman   manager.PackageManager
}

func (s *SnapSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.paccmder = commands.NewSnapPackageCommander("stable", "classic")
	s.pacman = manager.NewSnapPackageManager("stable", "classic")
}

func (s *SnapSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
}

func (s *SnapSuite) TearDownTest(c *gc.C) {
	s.IsolationSuite.TearDownTest(c)
}

func (s *SnapSuite) TearDownSuite(c *gc.C) {
	s.IsolationSuite.TearDownSuite(c)
}

func (s *SnapSuite) TestGetProxySettingsEmpty(c *gc.C) {
	const expected = `error: snap "core" has no "proxy" configuration option`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), s.mockExitError(1))

	out, err := s.pacman.GetProxySettings()
	c.Assert(err, jc.ErrorIsNil)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.GetProxyCmd()))
	c.Assert(out, gc.Equals, proxy.Settings{})
}

func (s *SnapSuite) TestGetProxySettingsConfigured(c *gc.C) {
	const expected = `Key          Value
proxy.http   localhost:8080
proxy.https  localhost:8181
proxy.ftp  localhost:2121`
	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), nil)

	out, err := s.pacman.GetProxySettings()
	c.Assert(err, gc.IsNil)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.GetProxyCmd()))

	c.Assert(out, gc.Equals, proxy.Settings{
		Http:  "localhost:8080",
		Https: "localhost:8181",
		Ftp:   "localhost:2121",
	})
}

func (s *SnapSuite) TestSearchForExistingPackage(c *gc.C) {
	const expected = `name:      juju
summary:   juju client
publisher: Canonical✓
contact:   https://jaas.ai/
license:   unset
description: |
  Juju is an open source modelling tool for operating software in the cloud.  Juju allows you to
  ...

  https://discourse.jujucharms.com/
  https://docs.jujucharms.com/
  https://github.com/juju/juju
commands:
  - juju
snap-id:      e2CPHpB1fUxcKtCyJTsm5t3hN9axJ0yj
tracking:     2.6/stable
refresh-date: today at 15:58 BST
channels:
  stable:        2.6.6                     2019-07-31 (8594) 68MB classic
  candidate:     ↑
  beta:          ↑
  edge:          2.7-beta1+develop-93d21f2 2019-08-19 (8756) 75MB classic
  2.6/stable:    2.6.6                     2019-07-31 (8594) 68MB classic
  ...
  2.3/beta:      ↑
  2.3/edge:      2.3.10+2.3-41313d1        2019-03-25 (7080) 55MB classic
installed:       2.6.6                                (8594) 68MB classic
`
	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), nil)
	exists, err := s.pacman.Search("juju")
	c.Assert(err, gc.IsNil)
	c.Assert(exists, jc.IsTrue)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.SearchCmd("juju")))
}

func (s *SnapSuite) TestSearchForUnknownPackage(c *gc.C) {
	const expected = `error: no snap found for "foo"`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), s.mockExitError(1))
	exists, err := s.pacman.Search("foo")
	c.Assert(err, gc.IsNil)
	c.Assert(exists, jc.IsFalse)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.SearchCmd("foo")))
}

func (s *SnapSuite) TestIsInstalled(c *gc.C) {
	const expected = `Name  Version  Rev   Tracking  Publisher   Notes
juju  2.6.6    8594  2.6       canonical✓  classic
`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), nil)
	installed := s.pacman.IsInstalled("juju")
	c.Assert(installed, jc.IsTrue)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.IsInstalledCmd("juju")))
}

func (s *SnapSuite) TestIsInstalledForUnknownPackage(c *gc.C) {
	const expected = `error: no matching snaps installed`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), s.mockExitError(1))
	installed := s.pacman.IsInstalled("foo")
	c.Assert(installed, jc.IsFalse)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.IsInstalledCmd("foo")))
}

func (s *SnapSuite) TestInstall(c *gc.C) {
	const expected = `juju 2.6.6 from Canonical✓ installed`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), nil)
	err := s.pacman.Install("juju")
	c.Assert(err, gc.IsNil)

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.InstallCmd("juju")))
}

func (s *SnapSuite) TestInstallForUnknownPackage(c *gc.C) {
	const expected = `error: snap "foo" not found`

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expected), s.mockExitError(1))
	err := s.pacman.Install("foo")
	c.Assert(err, gc.ErrorMatches, ".*unable to locate package")

	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, strings.Fields(s.paccmder.InstallCmd("foo")))
}

func (s *SnapSuite) mockExitError(code int) error {
	err := &exec.ExitError{ProcessState: new(os.ProcessState)}
	s.PatchValue(&manager.ProcessStateSys, func(*os.ProcessState) interface{} {
		return mockExitStatuser(code)
	})
	return err
}
