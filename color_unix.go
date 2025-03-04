// +build !windows,!darwin

package zenity

import (
	"image/color"

	"github.com/ncruces/zenity/internal/zenutil"
)

func selectColor(opts options) (color.Color, error) {
	args := []string{"--color-selection"}

	args = appendTitle(args, opts)
	if opts.color != nil {
		args = append(args, "--color", zenutil.UnparseColor(opts.color))
	}
	if opts.showPalette {
		args = append(args, "--show-palette")
	}

	out, err := zenutil.Run(opts.ctx, args)
	str, ok, err := strResult(opts, out, err)
	if ok {
		return zenutil.ParseColor(str), nil
	}
	return nil, err
}
