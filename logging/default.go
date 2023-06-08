package logging

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
)

func SetupDefaultLogger(level zerolog.Level) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "01-02 15:04:05.000"}
	defaultLogger := zerolog.New(output).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	log.Logger = defaultLogger

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
				fmt.Printf("=%+s:%d\r\n", frame, frame)
			}
		} else {
			return pkgerrors.MarshalStack(err)
		}
		return nil
	}
}
