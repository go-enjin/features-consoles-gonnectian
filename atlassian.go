//go:build all || (curses && atlassian)

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

package atlassian

import (
	"encoding/json"
	"fmt"

	"github.com/go-curses/cdk"
	cenums "github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/ctk"
	"github.com/go-curses/ctk/lib/enums"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	databaseFeature "github.com/go-enjin/be/features/database"
	"github.com/go-enjin/be/pkg/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/third_party/pkg/atlas-gonnect/store"
)

var _consoleAtlassian *Console

var _ feature.Console = (*Console)(nil)

const Tag feature.Tag = "AtlassianConsole"

const Version = "0.1.0"

type Console struct {
	feature.CConsole

	db *gorm.DB

	frame  ctk.Frame
	scroll ctk.ScrolledViewport
	vbox   ctk.VBox
}

type MakeConsole interface {
	feature.MakeConsole
}

func New() MakeConsole {
	if _consoleAtlassian == nil {
		_consoleAtlassian = new(Console)
		_consoleAtlassian.Init(_consoleAtlassian)
	}
	return _consoleAtlassian
}

func (f *Console) Depends() (deps feature.Tags) {
	deps = feature.Tags{
		databaseFeature.Tag,
	}
	return
}

func (f *Console) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Console) Name() (name string) {
	name = "atlassian-console"
	return
}

func (f *Console) Title() (title string) {
	title = fmt.Sprintf("Atlassian Console (v%v)", Version)
	return
}

func (f *Console) Build(b feature.Buildable) (err error) {
	log.DebugF("%v (v%v) build", Tag, Version)
	return
}

func (f *Console) Prepare(app ctk.Application) {
	f.CConsole.Prepare(app)

	var err error
	if f.db, err = database.Get(); err != nil {
		log.FatalF("error getting database connection")
	}
}

func (f *Console) Startup(display cdk.Display) {
	f.CConsole.Startup(display)

	window := f.Window()

	vbox := window.GetVBox()
	vbox.SetSpacing(1)

	f.frame = ctk.NewFrame("loading...")
	f.frame.Show()
	f.frame.SetLabelAlign(0.0, 0.5)
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

func (f *Console) Resized(w, h int) {

}

func (f *Console) Refresh() {
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
		tl := ctk.NewLabel("(no atlassian connect installations present")
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
		hbox.SetSizeRequest(-1, 1)
		f.vbox.PackStart(hbox, false, false, 0)

		tl := ctk.NewLabel(fmt.Sprintf("[%d] %v [debug=%v]", idx+1, tenant.BaseURL, debug))
		tl.SetJustify(cenums.JUSTIFY_LEFT)
		tl.SetSingleLineMode(true)
		tl.SetSizeRequest(-1, 1) // toggle-width box-child-space
		tl.Show()
		hbox.PackStart(tl, true, true, 0)

		bt := ctk.NewButtonWithLabel("Toggle Debug")
		bt.Show()
		bt.SetSizeRequest(-1, 1)
		bt.Connect(ctk.SignalActivate, "atlassian-console-activate-handler", f.toggleDebugHandler, tenant, ctx)
		hbox.PackStart(bt, false, false, 0)
	}
}

func (f *Console) toggleDebugHandler(data []interface{}, argv ...interface{}) cenums.EventFlag {
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