// Copyright 2019 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package commands

// const (
// 	// AptConfFilePath is the full file path for the proxy settings that are
// 	// written by cloud-init and the machine environ worker.
// 	AptConfFilePath = "/etc/apt/apt.conf.d/95-juju-proxy-settings"

// 	// the basic command for all dpkg calls:
// 	dpkg = "dpkg"

// 	// the basic command for all dpkg-query calls:
// 	dpkgquery = "dpkg-query"

// 	// the basic command for all apt-get calls:
// 	//		--force-confold is passed to dpkg to never overwrite config files
// 	//		--force-unsafe-io makes dpkg less sync-happy
// 	//		--assume-yes to never prompt for confirmation
// 	aptget = "apt-get --option=Dpkg::Options::=--force-confold --option=Dpkg::options::=--force-unsafe-io --assume-yes --quiet"

// 	// the basic command for all apt-cache calls:
// 	aptcache = "apt-cache"

// 	// the basic command for all add-apt-repository calls:
// 	//		--yes to never prompt for confirmation
// 	addaptrepo = "add-apt-repository --yes"

// 	// the basic command for all apt-config calls:
// 	aptconfig = "apt-config dump"

// 	// the basic format for specifying a proxy option for apt:
// 	aptProxySettingFormat = "Acquire::%s::Proxy %q;"

// 	// disable proxy for a specific host
// 	aptNoProxySettingFormat = "Acquire::%s::Proxy::%q \"DIRECT\";"

var getProxy = `
{
	grep -s ^http_proxy= /etc/environment  | sed 's/^http_proxy=//'  | sed 's/"//g';
	grep -s ^https_proxy= /etc/environment | sed 's/^https_proxy=//' | sed 's/"//g';

}
`[1:]

var setProxy = `
if ! grep -qs ^http_proxy= /etc/environment; then
	echo 'http_proxy="%s"' >> /etc/environment
else
	sed -i 's/^http_proxy=.*/http_proxy=%s/' /etc/environment
fi

if ! grep -qs ^https_proxy= /etc/environment; then
	echo 'https_proxy="%s"' >> /etc/environment
else
	sed -i 's/^https_proxy=.*/https_proxy=%s/' /etc/environment
fi
`[1:]

var unsetProxy = `
if grep -qs ^http_proxy= /etc/environment; then
	sed -i 's/^http_proxy=.*//' /etc/environment
fi

if grep -qs ^https_proxy= /etc/environment; then
	sed -i 's/^https_proxy=.*//' /etc/environment
fi
`[1:]

const (
	snap  = "snap"
	quiet = "2> /dev/null"
	noOp  = "true"
)

// snapCmder is the packageCommander instantiation for snap-based systems.
var snapCmder = packageCommander{
	prereq:           noOp,
	addRepository:    buildCommand(noOp, "%q"),
	removeRepository: buildCommand(noOp, "%q"),
	listRepositories: noOp,
	update:           buildCommand(snap, "refresh"),
	upgrade:          buildCommand(snap, "refresh", "%s"),
	listInstalled:    buildCommand(snap, "list"),
	install:          buildCommand(snap, "install"),
	cleanup:          buildCommand(snap, "refresh", quiet),
	listAvailable:    buildCommand(snap, "list"),
	remove:           buildCommand(snap, "remove"),
	purge:            buildCommand(snap, "remove"),
	search:           buildCommand(snap, "find", "%s"),
	isInstalled:      buildCommand(snap, "list", "|", `grep "^%s"`),
	getProxy:         getProxy,
	setProxy:         setProxy,
	setNoProxy:       unsetProxy,
}
