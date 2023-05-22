// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package gonnectian

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-curses/cdk"
	cenums "github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/ctk"
	"github.com/go-curses/ctk/lib/enums"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/store"

	"github.com/go-enjin/be/pkg/log"
)

var _ Panel = (*TenantsPanel)(nil)

var (
	PanelFirstFrameTheme   paint.ThemeName = "panel-frame-theme-first"
	PanelDefaultFrameTheme paint.ThemeName = "panel-frame-theme-default"
)

func init() {
	// style := paint.GetDefaultColorStyle()
	// borderStyle := style.Foreground(paint.ColorDarkBlue).Background(paint.ColorNavy)

	borderRunes, _ := paint.GetDefaultBorderRunes(paint.StockBorder)
	borderRunes.BottomLeft = paint.DefaultNilRune
	borderRunes.Bottom = paint.DefaultNilRune
	borderRunes.BottomRight = paint.DefaultNilRune
	borderRunes.Right = paint.DefaultNilRune
	borderRunes.TopRight = paint.DefaultNilRune
	borderRunes.Top = paint.RuneHLine
	borderRunes.TopLeft = paint.DefaultNilRune
	borderRunes.Left = paint.DefaultNilRune
	frameTheme := paint.GetDefaultColorTheme()
	// frameTheme.Border.Normal = borderStyle
	// frameTheme.Border.Active = borderStyle
	// frameTheme.Border.Prelight = borderStyle
	// frameTheme.Border.Selected = borderStyle
	// frameTheme.Border.Insensitive = borderStyle
	frameTheme.Border.BorderRunes = borderRunes

	paint.RegisterTheme(PanelDefaultFrameTheme, frameTheme)

	borderRunes, _ = paint.GetDefaultBorderRunes(paint.StockBorder)
	borderRunes.BottomLeft = paint.DefaultNilRune
	borderRunes.Bottom = paint.DefaultNilRune
	borderRunes.BottomRight = paint.DefaultNilRune
	borderRunes.Right = paint.DefaultNilRune
	borderRunes.TopRight = paint.DefaultNilRune
	borderRunes.Top = paint.DefaultNilRune
	borderRunes.TopLeft = paint.DefaultNilRune
	borderRunes.Left = paint.DefaultNilRune
	frameTheme = paint.GetDefaultColorTheme()
	// frameTheme.Border.Normal = borderStyle
	// frameTheme.Border.Active = borderStyle
	// frameTheme.Border.Prelight = borderStyle
	// frameTheme.Border.Selected = borderStyle
	// frameTheme.Border.Insensitive = borderStyle
	frameTheme.Border.BorderRunes = borderRunes

	paint.RegisterTheme(PanelFirstFrameTheme, frameTheme)
}

type TenantsPanel struct {
	curses *CCurses

	frame  ctk.Frame
	scroll ctk.ScrolledViewport
	list   ctk.VBox

	firstFrameTheme   paint.Theme
	defaultFrameTheme paint.Theme

	sync.RWMutex
}

func (t *TenantsPanel) Init(c *CCurses) (err error) {
	t.curses = c
	t.firstFrameTheme, _ = paint.GetTheme(PanelFirstFrameTheme)
	t.defaultFrameTheme, _ = paint.GetTheme(PanelDefaultFrameTheme)

	t.frame = ctk.NewFrame("tenants")
	t.frame.Show()

	t.scroll = ctk.NewScrolledViewport()
	t.scroll.Show()
	t.scroll.SetPolicy(enums.PolicyAutomatic, enums.PolicyNever)
	t.frame.Add(t.scroll)

	t.list = ctk.NewVBox(false, 0)
	t.list.Show()
	t.scroll.Add(t.list)
	return
}

func (t *TenantsPanel) Key() string {
	return "tenants"
}

func (t *TenantsPanel) Name() string {
	return "Tenants"
}

func (t *TenantsPanel) Show() {
	t.frame.Show()
}

func (t *TenantsPanel) Hide() {
	t.frame.Hide()
}

func (t *TenantsPanel) Refresh() {
	display := t.curses.console.Display()

	for _, child := range t.list.GetChildren() {
		t.list.Remove(child)
		child.Destroy()
	}

	var tenants []*store.Tenant
	t.curses.console.tx().Find(&tenants)
	numTenants := len(tenants)

	t.frame.SetLabel(fmt.Sprintf("%d tenants found:", numTenants))

	if numTenants == 0 {
		tl := ctk.NewLabel("(no gonnectian installations present")
		tl.SetAlignment(0.5, 0.5)
		tl.SetJustify(cenums.JUSTIFY_CENTER)
		tl.Show()
		t.list.PackStart(tl, true, true, 0)
		return
	}

	w, h := display.Screen().Size()
	width := w - 2 - 2 - 1 // borders frame-borders scroll
	height := numTenants * 5
	if height < h-7 {
		width += 1
	} else {
		width -= 1
	}
	t.list.SetSizeRequest(width, height)

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
		var allowedUnlicensed bool
		if v, ok := ctx["allowed-unlicensed"].(bool); ok {
			allowedUnlicensed = v
		}

		frame := ctk.NewFrame("")
		frame.Show()
		frame.SetLabelAlign(0.0, 0.5)
		frame.SetSizeRequest(-1, 5)
		if idx == 0 {
			frame.SetTheme(t.firstFrameTheme)
		} else {
			frame.SetTheme(t.defaultFrameTheme)
		}
		t.list.PackStart(frame, false, false, 0)

		hbox := ctk.NewHBox(false, 1)
		hbox.Show()
		hbox.SetSizeRequest(-1, 4)
		frame.Add(hbox)

		tenantText := fmt.Sprintf("[%d] %v (lic=%v)", idx+1, tenant.BaseURL, ctx["license"])
		tenantText += fmt.Sprintf("\n (c=%v / u=%v)", tenant.CreatedAt.Format("2006-01-02 15:04 MST"), tenant.UpdatedAt.Format("2006-01-02 15:04 MST"))
		if tenant.AddonInstalled {
			tenantText += "\n  (installed, "
		} else {
			tenantText += "\n  (not installed, "
		}
		if allowedUnlicensed {
			tenantText += " allowed unlicensed, "
		}
		if debug == "true" {
			tenantText += " debugging enabled)"
		} else {
			tenantText += " debugging disabled)"
		}

		tl := ctk.NewLabel(tenantText)
		tl.Show()
		tl.SetJustify(cenums.JUSTIFY_LEFT)
		tl.SetSingleLineMode(false)
		tl.SetLineWrap(false)
		tl.SetLineWrapMode(cenums.WRAP_NONE)
		tl.SetSizeRequest(-1, 4) // toggle-width box-child-space
		hbox.PackStart(tl, true, true, 0)

		vbox := ctk.NewVBox(false, 0)
		vbox.Show()
		hbox.PackEnd(vbox, false, true, 0)

		makeButton := func(buttonLabel, tooltipText, key string, handler cdk.SignalListenerFn) {
			bt := ctk.NewButtonWithLabel(buttonLabel)
			bt.Show()
			bt.SetSizeRequest(23, 1)
			bt.SetTooltipText(tooltipText)
			bt.SetHasTooltip(true)
			bt.Connect(ctk.SignalActivate, "gonnectian-console-"+key+"-handler", handler, tenant, ctx)
			vbox.PackStart(bt, false, false, 0)
		}

		var buttonLabel, tooltipText string
		if debug == "true" {
			buttonLabel = "Disable Debug"
			tooltipText = "Click to disable per-tenant UI debugging"
		} else {
			buttonLabel = "Enable Debug"
			tooltipText = "Click to enable per-tenant UI debugging"
		}
		makeButton(buttonLabel, tooltipText, "debug", t.toggleDebugHandler)

		if allowedUnlicensed {
			buttonLabel = "Reject Unlicensed"
			tooltipText = "Click to reject unlicensed installations for this tenant"
		} else {
			buttonLabel = "Allow Unlicensed"
			tooltipText = "Click to allow unlicensed installations for this tenant"
		}
		makeButton(buttonLabel, tooltipText, "unlicensed", t.toggleUnlicensedHandler)
	}

}

func (t *TenantsPanel) Container() ctk.Container {
	return t.frame
}

func (t *TenantsPanel) toggleDebugHandler(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if len(data) == 2 {
		if tenant, ok := data[0].(*store.Tenant); ok {
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
					tenant.Context = b
					if err := t.curses.console.tx().Save(&tenant).Error; err != nil {
						log.ErrorF("error saving tenant database change: %v", err)
					}
					t.curses.Refresh()
				}
			}
		}
	}
	return cenums.EVENT_STOP
}

func (t *TenantsPanel) toggleUnlicensedHandler(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if len(data) == 2 {
		if tenant, ok := data[0].(*store.Tenant); ok {
			if c, ok := data[1].(map[string]interface{}); ok {
				if v, ok := c["allowed-unlicensed"].(bool); ok {
					if v {
						c["allowed-unlicensed"] = false
					} else {
						c["allowed-unlicensed"] = true
						delete(c, "reject")
					}
				} else {
					c["allowed-unlicensed"] = true
					delete(c, "reject")
				}
				if b, err := json.Marshal(c); err != nil {
					log.ErrorF("error encoding tenant context change: %v", err)
				} else {
					tenant.Context = b
					if err := t.curses.console.tx().Save(&tenant).Error; err != nil {
						log.ErrorF("error saving tenant database change: %v", err)
					}
					t.curses.Refresh()
				}
			}
		}
	}
	return cenums.EVENT_STOP
}