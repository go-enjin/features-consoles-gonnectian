//go:build curses || all

// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gonnectian

import (
	"fmt"
	"sync"

	cenums "github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/ctk"
	"github.com/go-curses/ctk/lib/enums"

	gonnectian "github.com/go-enjin/features-gonnectian"
)

var _ Panel = (*AppInfoPanel)(nil)

type AppInfoPanel struct {
	curses *CCurses

	frame  ctk.Frame
	scroll ctk.ScrolledViewport
	label  ctk.Label

	sync.RWMutex
}

func (a *AppInfoPanel) Init(c *CCurses) (err error) {
	a.curses = c
	a.frame = ctk.NewFrame("Application Info")
	a.frame.Show()
	a.scroll = ctk.NewScrolledViewport()
	a.scroll.Show()
	a.scroll.SetPolicy(enums.PolicyNever, enums.PolicyAutomatic)
	a.frame.Add(a.scroll)
	a.label = ctk.NewLabel("")
	a.label.Show()
	a.label.SetJustify(cenums.JUSTIFY_NONE)
	a.label.SetLineWrap(false)
	a.label.SetLineWrapMode(cenums.WRAP_NONE)
	a.label.SetSizeRequest(70, 20)
	a.scroll.Add(a.label)
	return
}

func (a *AppInfoPanel) Key() string {
	return "app-info"
}

func (a *AppInfoPanel) Name() string {
	return "App Info"
}

func (a *AppInfoPanel) Show() {
	a.frame.Show()
}

func (a *AppInfoPanel) Hide() {
	a.frame.Hide()
}

func (a *AppInfoPanel) Refresh() {

	info := make(map[string]string)
	numVersions := 0
	for _, f := range a.curses.console.Enjin.Features() {
		if af, ok := f.(*gonnectian.CFeature); ok {
			url := af.GetPluginInstallationURL()
			dsc := af.GetPluginDescriptor()
			if _, ok := info[dsc.Name]; !ok {
				info[dsc.Name] = ""
			} else {
				info[dsc.Name] += "\n"
			}
			info[dsc.Name] += fmt.Sprintf(" - [%v] %v", dsc.Version, url)
			numVersions += 1
		}
	}
	var infoText string
	for k, v := range info {
		if infoText != "" {
			infoText += "\n"
		}
		infoText += fmt.Sprintf("%v\n%v", k, v)
	}
	a.label.SetText(infoText)
	a.frame.SetLabel(fmt.Sprintf("%d applications, %d total versions", len(info), numVersions))
}

func (a *AppInfoPanel) Container() ctk.Container {
	return a.frame
}