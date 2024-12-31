// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/mkloubert/go-package-manager/utils"
)

// ProjectVersionManager manages versions of a project
// based on Git tags
type ProjectVersionManager struct {
	app *AppContext
}

// BumpProjectVersionOptions stores options for `Bump()“ method
// of `ProjectVersionManager“ instance
type BumpProjectVersionOptions struct {
	Breaking *bool   // increase major part
	Feature  *bool   // increase minor part
	Fix      *bool   // increase patch part
	Force    *bool   // force bump even if latest version is newer
	Major    *int64  // if defined, the initial value for new major part
	Message  *string // the custom git message
	Minor    *int64  // if defined, the initial value for minor part
	Patch    *int64  // if defined, the initial value for patch part
}

// pvm.Bump() - bumps the version of the current project, based on the current settings
// by default minor version is increased
func (pvm *ProjectVersionManager) Bump(options ...BumpProjectVersionOptions) (*version.Version, error) {
	latestVersion, err := pvm.GetLatestVersion()
	if err != nil {
		return nil, err
	}

	breaking := false
	feature := false
	fix := false
	force := false
	var major int64 = -1
	message := ""
	var minor int64 = -1
	var patch int64 = -1
	for _, o := range options {
		if o.Breaking != nil {
			breaking = *o.Breaking
		}
		if o.Feature != nil {
			feature = *o.Feature
		}
		if o.Fix != nil {
			fix = *o.Fix
		}
		if o.Force != nil {
			force = *o.Force
		}
		if o.Major != nil {
			major = *o.Major
		}
		if o.Message != nil {
			message = strings.TrimSpace(*o.Message)
		}
		if o.Minor != nil {
			minor = *o.Minor
		}
		if o.Patch != nil {
			patch = *o.Patch
		}
	}

	if latestVersion == nil {
		latestVersion, _ = version.NewVersion("0.0.0")
	}

	segments := latestVersion.Segments64()
	currentMajor := segments[0]
	currentMinor := segments[1]
	currentPatch := segments[2]

	newMajor := currentMajor
	if major > -1 {
		newMajor = major
	}
	newMinor := currentMinor
	if minor > -1 {
		newMinor = minor
	}
	newPatch := currentPatch
	if patch > -1 {
		newPatch = patch
	}

	if !breaking && !feature && !fix {
		// default: 1.2.3 => 1.3.0

		newMinor++
		newPatch = 0
	} else {
		if breaking {
			newMajor++ // by default e.g.: 1.2.3 => 2.0.0
			if !feature {
				newMinor = 0
			}
			if !fix {
				newPatch = 0
			}
		}
		if feature {
			newMinor++ // by default e.g.: 1.2.3 => 1.3.0
			if !fix {
				newPatch = 0
			}
		}
		if fix {
			newPatch++ // e.g. 1.2.3 => 1.2.4
		}
	}

	nextVersion, err := version.NewVersion(
		fmt.Sprintf(
			"%v.%v.%v",
			newMajor, newMinor, newPatch,
		),
	)
	if err != nil {
		return nextVersion, err
	}

	if !force && nextVersion.LessThanOrEqual(latestVersion) {
		return nextVersion, fmt.Errorf("new version is not greater than latest one")
	}

	gitMessage := strings.TrimSpace(message)
	if gitMessage == "" {
		gitMessage = fmt.Sprintf("version %v", nextVersion.String())
	}

	tagName := fmt.Sprintf("v%v", nextVersion.String())

	p := utils.CreateShellCommandByArgs("git", "tag", "-a", tagName, "-m", gitMessage)
	p.Dir = pvm.app.Cwd

	err = p.Run()

	return nextVersion, err
}

// pvm.GetLatestVersion() - Returns the latest version based on the Git tags
// of the current repository or nil if not found.
func (pvm *ProjectVersionManager) GetLatestVersion() (*version.Version, error) {
	allVersions, err := pvm.GetVersions()
	if err != nil {
		return nil, err
	}

	var latestVersion *version.Version
	for _, v := range allVersions {
		updateVersion := func() {
			latestVersion = v
		}

		if latestVersion != nil {
			if latestVersion.LessThanOrEqual(v) {
				updateVersion()
			}
		} else {
			updateVersion()
		}
	}

	return latestVersion, nil
}

// pvm.GetVersions() - Returns all versions represented by Git tags
// inside the current working directory.
func (pvm *ProjectVersionManager) GetVersions() ([]*version.Version, error) {
	var versions []*version.Version

	tags, err := pvm.app.GetGitTags()
	if err != nil {
		return versions, err
	}

	for _, t := range tags {
		v, err := version.NewVersion(t)
		if err == nil {
			versions = append(versions, v)
		}
	}

	return versions, nil
}
