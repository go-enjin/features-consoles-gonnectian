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

	"gorm.io/gorm"

	"github.com/go-curses/cdk"
	cenums "github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/ctk"
	"github.com/go-curses/ctk/lib/enums"
)

var gKnownPanels []Panel
var ButtonActiveTheme paint.ThemeName = "toggle-button-active"

func init() {
	gKnownPanels = append(gKnownPanels,
		&AppInfoPanel{},
		&TenantsPanel{},
	)

	borders, _ := paint.GetDefaultBorderRunes(paint.StockBorder)
	arrows, _ := paint.GetArrows(paint.StockArrow)

	style := paint.GetDefaultColorStyle()
	styleNormal := style.Foreground(paint.ColorWhite).Background(paint.ColorDarkGreen)
	styleActive := style.Foreground(paint.ColorWhite).Background(paint.ColorForestGreen)
	styleInsensitive := style.Foreground(paint.ColorDarkSlateGray).Background(paint.ColorRosyBrown)

	paint.RegisterTheme(ButtonActiveTheme, paint.Theme{
		Content: paint.ThemeAspect{
			Normal:      styleNormal.Dim(false).Bold(true),
			Selected:    styleActive.Dim(false).Bold(true),
			Active:      styleActive.Dim(false).Bold(true).Reverse(true),
			Prelight:    styleActive.Dim(false),
			Insensitive: styleInsensitive.Dim(true),
			FillRune:    paint.DefaultFillRune,
			BorderRunes: borders,
			ArrowRunes:  arrows,
			Overlay:     false,
		},
		Border: paint.ThemeAspect{
			Normal:      styleNormal.Dim(true).Bold(false),
			Selected:    styleActive.Dim(false).Bold(true),
			Active:      styleActive.Dim(false).Bold(true).Reverse(true),
			Prelight:    styleActive.Dim(false),
			Insensitive: styleInsensitive.Dim(true),
			FillRune:    paint.DefaultFillRune,
			BorderRunes: borders,
			ArrowRunes:  arrows,
			Overlay:     false,
		},
	})
}

type CCurses struct {
	db      *gorm.DB
	console *CConsole

	pOrder []string
	panels map[string]Panel
	active string

	window     ctk.Window
	panelArea  ctk.VBox
	toggleArea ctk.HButtonBox
	toggles    map[string]ctk.Button

	defaultToggleTheme paint.Theme
	activeToggleTheme  paint.Theme

	sync.RWMutex
}

func NewCurses(console *CConsole) (c *CCurses, err error) {
	c = &CCurses{
		db:      console.db,
		window:  console.Window(),
		console: console,
		toggles: make(map[string]ctk.Button),
		panels:  make(map[string]Panel),
	}
	c.defaultToggleTheme, _ = paint.GetTheme(ctk.ButtonColorTheme)
	c.activeToggleTheme, _ = paint.GetTheme("toggle-button-active")

	vbox := c.window.GetVBox()

	c.panelArea = ctk.NewVBox(false, 0)
	c.panelArea.Show()
	vbox.PackStart(c.panelArea, true, true, 0)

	c.toggleArea = ctk.NewHButtonBox(false, 1)
	c.toggleArea.Show()
	vbox.PackEnd(c.toggleArea, false, true, 0)

	// accelMap := ctk.NewAccelerator("/quit")

	for idx, panel := range gKnownPanels {
		if err = panel.Init(c); err != nil {
			err = fmt.Errorf("error init %v panel: %v", panel.Key(), err)
			return
		}
		c.panels[panel.Key()] = panel
		c.pOrder = append(c.pOrder, panel.Key())
		toggle := c.makePanelToggle(idx+1, panel)
		c.toggles[panel.Key()] = toggle
		c.panelArea.PackStart(panel.Container(), true, true, 0)
		c.toggleArea.PackStart(toggle, false, false, 0)
		if idx == 0 {
			c.active = panel.Key()
			toggle.SetTheme(c.activeToggleTheme)
		} else {
			toggle.SetTheme(c.defaultToggleTheme)
		}
	}

	b := c.makeToggleButton("quit", "Quit <F10>", cdk.KeyF10, func(data []interface{}, argv ...interface{}) cenums.EventFlag {
		c.console.Display().RequestQuit()
		return cenums.EVENT_STOP
	})
	c.toggleArea.SetChildSecondary(b, true)
	sep := ctk.NewSeparator()
	sep.Show()
	sep.SetSizeRequest(-1, 1)
	c.toggleArea.SetChildSecondary(sep, true)
	c.toggleArea.SetChildPacking(sep, true, true, 0, enums.PackStart)
	return
}

func (c *CCurses) makePanelToggle(id int, p Panel) (b ctk.Button) {
	label := fmt.Sprintf("%s <F%d>", p.Name(), id)
	accelKey := cdk.Key(0)
	if id > 0 && id < 10 {
		accelKey = cdk.Key(int16(cdk.KeyF1) + int16(id-1))
	}
	return c.makeToggleButton(p.Key(), label, accelKey, c.togglePanelHandler, id, p)
}

func (c *CCurses) makeToggleButton(key, label string, accelKey cdk.Key, handler cdk.SignalListenerFn, data ...interface{}) (b ctk.Button) {
	b = ctk.NewButtonWithLabel(label)
	b.Show()
	b.SetSizeRequest(-1, 1)
	b.Connect(ctk.SignalActivate, key+"-toggle-handler", handler, data...)
	if accelKey > 0 {
		accelGroup := ctk.NewAccelGroup()
		accelGroup.AccelConnect(accelKey, cdk.ModNone, 0, key+"-toggle-accel", func(argv ...interface{}) (handled bool) {
			b.GrabFocus()
			b.Activate()
			return
		})
		c.console.Window().AddAccelGroup(accelGroup)
	}
	return
}

func (c *CCurses) togglePanelHandler(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if p, ok := data[1].(Panel); ok {
		c.active = p.Key()
		c.Refresh()
	}
	return cenums.EVENT_PASS
}

func (c *CCurses) Refresh() {
	p, _ := c.panels[c.active]
	c.window.Freeze()
	c.panelArea.Freeze()
	for key, panel := range c.panels {
		if key != p.Key() {
			panel.Hide()
			if b, ok := c.toggles[panel.Key()]; ok {
				b.SetTheme(c.defaultToggleTheme)
			}
		}
	}
	if b, ok := c.toggles[c.active]; ok {
		b.SetTheme(c.activeToggleTheme)
		b.GrabFocus()
	}
	p.Show()
	p.Refresh()
	c.panelArea.Thaw()
	c.window.Thaw()
	c.window.Resize()
	c.console.Display().RequestDraw()
	c.console.Display().RequestShow()
}