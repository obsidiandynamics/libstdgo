package seelog

import (
	"testing"

	"github.com/cihub/seelog"
	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
)

func Example() {
	const logConfig = `
			<seelog type="sync">
					<outputs formatid="main">
							<console/>
					</outputs>
					<formats>
							<format id="main" format="%Time %Date %LEV %File:%Line: %Msg"/>
					</formats>
			</seelog>
	`
	binding := Bind(func() seelog.LoggerInterface {
		logger, err := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
		if err != nil {
			panic(err)
		}
		return logger
	})
	s := scribe.New(binding.Factories())

	// Do some logging
	s.I()("Important application message")

	// Eventually, when the logger is no longer required...
	binding.Close()
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
