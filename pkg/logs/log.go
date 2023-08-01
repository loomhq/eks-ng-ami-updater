package logs

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

func Setup(debug bool) context.Context {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return zerolog.New(os.Stdout).WithContext(context.Background())
}
