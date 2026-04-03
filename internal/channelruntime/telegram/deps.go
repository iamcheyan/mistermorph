package telegram

import "github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"

type HandleModelCommandFunc func(text string) (string, bool, error)

type Dependencies struct {
	depsutil.CommonDependencies
	HandleModelCommand HandleModelCommandFunc
}
