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
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/ctk"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/store"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
)

var (
	_ Console = (*CConsole)(nil)
)

const (
	Tag     feature.Tag = "console-gonnectian"
	Version             = "0.2.1"
)

type Console interface {
	feature.Console
}

type MakeConsole interface {
	Make() Console

	SetGormDB(tag string) MakeConsole
	SetTableName(table string) MakeConsole
}

type CConsole struct {
	feature.CConsole

	prefix string

	dbTag   string
	dbTable string

	db *gorm.DB

	curses *CCurses

	infoLabel ctk.Label
	frame     ctk.Frame
	scroll    ctk.ScrolledViewport
	vbox      ctk.VBox
}

func New() MakeConsole {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeConsole {
	f := new(CConsole)
	f.Init(f)
	f.PackageTag = Tag
	f.ConsoleTag = tag
	return f
}

func (f *CConsole) Init(this interface{}) {
	f.CConsole.Init(this)
}

func (f *CConsole) SetGormDB(tag string) MakeConsole {
	f.dbTag = tag
	return f
}

func (f *CConsole) SetTableName(table string) MakeConsole {
	f.dbTable = table
	return f
}

func (f *CConsole) Make() (c Console) {
	if f.dbTag == "" {
		log.FatalDF(1, "%v feature requires .SetGormDB and .SetTableName", f.Tag())
	} else if f.dbTable == "" {
		log.FatalDF(1, "%v feature requires .SetTableName and .SetGormDB", f.Tag())
	}
	return f
}

func (f *CConsole) Title() (title string) {
	if f.prefix != "" {
		title = fmt.Sprintf("Gonnectian v%v (%v %v) [%v]", Version, globals.BinName, globals.Version, f.prefix)
		return
	}
	title = fmt.Sprintf("Gonnectian v%v (%v %v)", Version, globals.BinName, globals.Version)
	return
}

func (f *CConsole) Build(b feature.Buildable) (err error) {
	log.DebugF("%v (v%v) build", Tag, Version)
	return
}

func (f *CConsole) Setup(ctx *cli.Context, ei feature.Internals) {
	f.CConsole.Setup(ctx, ei)
	f.prefix = ctx.String("prefix")
}

func (f *CConsole) Prepare(app ctk.Application) {
	f.CConsole.Prepare(app)
	f.db = f.Enjin.MustDB(f.dbTag).(*gorm.DB)
	// no auto-migrate, manipulates features-gonnectian data
}

func (f *CConsole) Startup(display cdk.Display) {
	f.CConsole.Startup(display)

	var err error
	if f.curses, err = NewCurses(f); err != nil {
		f.App().NotifyStartupComplete()
		display.RequestQuit()
		time.Sleep(100 * time.Millisecond)
		display.Destroy()
		_, _ = fmt.Fprintf(os.Stderr, "error constructing curses user interface: %v\n", err)
		os.Exit(1)
		return
	}

	f.curses.Refresh()

	f.Window().Show()
	f.App().NotifyStartupComplete()
}

func (f *CConsole) Resized(w, h int) {
	log.DebugF("refreshing on resized: %v, %v", w, h)
	f.Refresh()
}

func (f *CConsole) Refresh() {
	window := f.Window()

	window.Freeze()
	window.GetVBox().Freeze()

	defer func() {
		window.GetVBox().Thaw()
		window.Thaw()
		window.Resize()
		f.Display().RequestDraw()
		f.Display().RequestShow()
	}()

	f.curses.Refresh()
}

func (f *CConsole) tx() (tx *gorm.DB) {
	tx = f.db.Scopes(func(tx *gorm.DB) *gorm.DB {
		if f.dbTable == "" {
			return tx.Table(store.DefaultTableName)
		}
		return tx.Table(f.dbTable)
	})
	return
}