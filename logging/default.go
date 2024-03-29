package logging

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
)

func SetupDefaultLoggerWithColor(level zerolog.Level, noColor bool) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: noColor, TimeFormat: "01-02 15:04:05.000"}
	defaultLogger := zerolog.New(output).Level(level).With().Timestamp().Logger()
	log.Logger = defaultLogger
	zerolog.TimeFieldFormat = "01-02 15:04:05.000"

	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		type stackTracer interface {
			StackTrace() errors.StackTrace
		}
		e, ok := err.(stackTracer)
		if !ok {
			return nil
		}
		//It's mean when env=dev just print track
		if true {
			fmt.Printf("%v\n", e)
			for _, frame := range e.StackTrace() {
				fmt.Printf("%+s:%d\r\n", frame, frame)
			}
		} else {
			return pkgerrors.MarshalStack(err)
		}
		return nil
	}
}

func SetupDefaultLogger(level zerolog.Level) {
	SetupDefaultLoggerWithColor(level, true)
}
