// Copyright (c) 2022  The Go-Enjin Authors
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
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/go-curses/cdk"
	cenums "github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/ctk"
	"github.com/go-curses/ctk/lib/enums"

	databaseFeature "github.com/go-enjin/be/features/database"
	"github.com/go-enjin/be/pkg/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/features-gonnectian"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/store"
)

var (
	_ Console = (*CConsole)(nil)
)

const (
	Tag     feature.Tag = "AtlasGonnect"
	Name                = "atlas-gonnect"
	Version             = "0.1.0"
)

type Console interface {
	feature.MakeConsole
	feature.Console
}

type CConsole struct {
	feature.CConsole

	prefix string

	db *gorm.DB
	ei feature.Internals

	infoLabel ctk.Label
	frame     ctk.Frame
	scroll    ctk.ScrolledViewport
	vbox      ctk.VBox
}

func New() feature.MakeConsole {
	f := new(CConsole)
	f.Init(f)
	return f
}

func (f *CConsole) Depends() (deps feature.Tags) {
	deps = feature.Tags{
		databaseFeature.Tag,
	}
	return
}

func (f *CConsole) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CConsole) Name() (name string) {
	name = Name
	return
}

func (f *CConsole) Title() (title string) {
	if f.prefix != "" {
		title = fmt.Sprintf("Atlas-Gonnect v%v (%v %v) [%v]", Version, globals.BinName, globals.Version, f.prefix)
		return
	}
	title = fmt.Sprintf("Atlas-Gonnect v%v (%v %v)", Version, globals.BinName, globals.Version)
	return
}

func (f *CConsole) Build(b feature.Buildable) (err error) {
	log.DebugF("%v (v%v) build", Tag, Version)
	return
}

func (f *CConsole) Setup(ctx *cli.Context, ei feature.Internals) {
	f.prefix = ctx.String("prefix")
	f.ei = ei
}

func (f *CConsole) Prepare(app ctk.Application) {
	f.CConsole.Prepare(app)

	var err error
	if f.db, err = database.Get(); err != nil {
		log.FatalF("error getting database connection")
	}
}

func (f *CConsole) Startup(display cdk.Display) {
	f.CConsole.Startup(display)

	window := f.Window()

	vbox := window.GetVBox()
	vbox.SetSpacing(1)

	f.infoLabel = ctk.NewLabel("")
	f.infoLabel.Show()
	f.infoLabel.SetLineWrapMode(cenums.WRAP_NONE)
	f.infoLabel.SetJustify(cenums.JUSTIFY_NONE)
	vbox.PackStart(f.infoLabel, false, false, 0)

	f.frame = ctk.NewFrame("loading...")
	f.frame.Show()
	f.frame.SetLabelAlign(0.0, 0.5)
	ft := f.frame.GetTheme()
	ft.Border.BorderRunes.TopLeft = paint.DefaultFillRune
	ft.Border.BorderRunes.Left = paint.DefaultFillRune
	ft.Border.BorderRunes.BottomLeft = paint.DefaultFillRune
	ft.Border.BorderRunes.Bottom = paint.DefaultFillRune
	ft.Border.BorderRunes.BottomRight = paint.DefaultFillRune
	ft.Border.BorderRunes.Right = paint.DefaultFillRune
	ft.Border.BorderRunes.TopRight = paint.DefaultFillRune
	f.frame.SetTheme(ft)
	vbox.PackStart(f.frame, true, true, 0)

	f.scroll = ctk.NewScrolledViewport()
	f.scroll.Show()
	f.scroll.SetPolicy(enums.PolicyNever, enums.PolicyAutomatic)
	f.frame.Add(f.scroll)

	f.vbox = ctk.NewVBox(false, 1)
	f.vbox.Show()
	f.scroll.Add(f.vbox)

	f.Refresh()

	window.Show()
	f.App().NotifyStartupComplete()
}

func (f *CConsole) Resized(w, h int) {
	log.DebugF("refreshing on resized: %v, %v", w, h)
	f.Refresh()
}

func (f *CConsole) Refresh() {
	window := f.Window()
	display := f.Display()

	window.Freeze()
	f.vbox.Freeze()

	defer func() {
		f.vbox.Thaw()
		window.Thaw()
		window.Resize()
		f.Display().RequestDraw()
		f.Display().RequestShow()
	}()

	info := make(map[string]string)
	for _, f := range f.ei.Features() {
		if af, ok := f.(*gonnectian.Feature); ok {
			url := af.GetPluginInstallationURL()
			dsc := af.GetPluginDescriptor()
			if _, ok := info[dsc.Name]; !ok {
				info[dsc.Name] = ""
			} else {
				info[dsc.Name] += "\n"
			}
			info[dsc.Name] += fmt.Sprintf(" - [%v] %v", dsc.Version, url)
		}
	}
	var infoText string
	for k, v := range info {
		if infoText != "" {
			infoText += "\n"
		}
		infoText += fmt.Sprintf("%v\n%v", k, v)
	}
	f.infoLabel.SetText(infoText)

	for _, child := range f.vbox.GetChildren() {
		f.vbox.Remove(child)
		child.Destroy()
	}

	w, _ := display.Screen().Size()
	width := w - 2 - 2 - 1 // borders frame-borders scroll
	f.vbox.SetSizeRequest(width, -1)

	var tenants []*store.Tenant
	f.db.Find(&tenants)
	numTenants := len(tenants)

	f.frame.SetLabel(fmt.Sprintf("%d tenants found:", numTenants))

	if numTenants == 0 {
		tl := ctk.NewLabel("(no gonnectian installations present")
		tl.SetAlignment(0.5, 0.5)
		tl.SetJustify(cenums.JUSTIFY_CENTER)
		tl.Show()
		f.vbox.PackStart(tl, true, true, 0)
		return
	}

	for idx, tenant := range tenants {
		var ctx map[string]interface{}
		contextJson := tenant.Context.String()
		if contextJson == "" {
			contextJson = `{"debug":"false"}`
		}
		if err := json.Unmarshal([]byte(contextJson), &ctx); err != nil {
			log.ErrorF("error parsing tenant context: %v", err)
		}
		var debug string
		if v, ok := ctx["debug"].(string); ok {
			debug = v
		} else {
			debug = "false"
		}

		hbox := ctk.NewHBox(false, 1)
		hbox.Show()
		hbox.SetSizeRequest(-1, 3)
		f.vbox.PackStart(hbox, false, false, 0)

		tenantText := fmt.Sprintf("[%d] %v", idx+1, tenant.BaseURL)
		tenantText += fmt.Sprintf("\n (c=%v / u=%v)", tenant.CreatedAt.Format("2006-01-02 15:04 MST"), tenant.UpdatedAt.Format("2006-01-02 15:04 MST"))
		if tenant.AddonInstalled {
			tenantText += "\n  (installed, "
		} else {
			tenantText += "\n  (not installed, "
		}
		if debug == "true" {
			tenantText += " debugging enabled)"
		} else {
			tenantText += " debugging disabled)"
		}

		tl := ctk.NewLabel(tenantText)
		tl.SetJustify(cenums.JUSTIFY_LEFT)
		tl.SetSingleLineMode(false)
		tl.SetLineWrap(false)
		tl.SetLineWrapMode(cenums.WRAP_NONE)
		tl.SetSizeRequest(-1, 4) // toggle-width box-child-space
		tl.Show()
		hbox.PackStart(tl, true, true, 0)

		var buttonLabel, tooltipText string
		if debug == "true" {
			buttonLabel = "Disable Debug"
			tooltipText = "Click to disable per-tenant UI debugging"
		} else {
			buttonLabel = "Enable Debug"
			tooltipText = "Click to enable per-tenant UI debugging"
		}
		bt := ctk.NewButtonWithLabel(buttonLabel)
		bt.Show()
		bt.SetSizeRequest(20, 3)
		bt.SetTooltipText(tooltipText)
		bt.SetHasTooltip(true)
		bt.Connect(ctk.SignalActivate, "gonnectian-console-activate-handler", f.toggleDebugHandler, tenant, ctx)
		hbox.PackStart(bt, false, false, 0)
	}
}

func (f *CConsole) toggleDebugHandler(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if len(data) == 2 {
		if t, ok := data[0].(*store.Tenant); ok {
			if c, ok := data[1].(map[string]interface{}); ok {
				if v, ok := c["debug"]; ok {
					if v == "true" {
						c["debug"] = "false"
					} else {
						c["debug"] = "true"
					}
				} else {
					c["debug"] = "true"
				}
				if b, err := json.Marshal(c); err != nil {
					log.ErrorF("error encoding tenant context change: %v", err)
				} else {
					t.Context = datatypes.JSON(b)
					if err := f.db.Save(&t).Error; err != nil {
						log.ErrorF("error saving tenant database change: %v", err)
					}
					f.Refresh()
				}
			}
		}
	}
	return cenums.EVENT_STOP
}