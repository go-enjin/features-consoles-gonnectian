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
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/ctk"

	databaseFeature "github.com/go-enjin/be/features/database"
	"github.com/go-enjin/be/pkg/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
)

var (
	_ Console = (*CConsole)(nil)
)

const (
	Tag     feature.Tag = "Gonnectian"
	Name                = "gonnectian"
	Version             = "0.2.0"
)

type Console interface {
	feature.Console
}

type MakeConsole interface {
	Make() feature.Console
}

type CConsole struct {
	feature.CConsole

	prefix string

	db *gorm.DB
	ei feature.Internals

	curses *CCurses

	infoLabel ctk.Label
	frame     ctk.Frame
	scroll    ctk.ScrolledViewport
	vbox      ctk.VBox
}

func New() MakeConsole {
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
		title = fmt.Sprintf("Gonnectian v%v (%v %v) [%v]", Version, globals.BinName, globals.Version, f.prefix)
		return
	}
	title = fmt.Sprintf("Gonnectian v%v (%v %v)", Version, globals.BinName, globals.Version)
	return
}

func (f *CConsole) Make() (c feature.Console) {
	return f.Self()
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