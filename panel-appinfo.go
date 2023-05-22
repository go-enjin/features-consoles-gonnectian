// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

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