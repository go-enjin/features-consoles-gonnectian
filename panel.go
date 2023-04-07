// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package gonnectian

import "github.com/go-curses/ctk"

type Panel interface {
	Key() string
	Name() string
	Init(c *CCurses) (err error)

	Show()
	Hide()
	Refresh()
	Container() ctk.Container
}