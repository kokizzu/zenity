// +build !windows,!darwin

package zenity

import (
	"strings"

	"github.com/ncruces/zenity/internal/zenutil"
)

func password(opts options) (string, string, bool, error) {
	args := []string{"--password"}
	args = appendTitle(args, opts)
	args = appendButtons(args, opts)
	if opts.username {
		args = append(args, "--username")
	}

	out, err := zenutil.Run(opts.ctx, args)
	str, ok, err := strResult(opts, out, err)
	if ok && opts.username {
		if split := strings.SplitN(string(out), "|", 2); len(split) == 2 {
			return split[0], split[1], true, nil
		}
	}
	return "", str, ok, err
}
