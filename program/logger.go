package program

import (
	"github.com/latifrons/commongo/utilfuncs"
	"github.com/sirupsen/logrus"
	"os"
)

func SetupLogger(level string) {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "01-02 15:04:05.000000"
	formatter.FullTimestamp = true
	formatter.ForceColors = true

	lvl, err := logrus.ParseLevel(level)
	utilfuncs.PanicIfError(err, "log level")
	logrus.SetLevel(lvl)
	logrus.SetFormatter(formatter)
	logrus.StandardLogger().SetOutput(os.Stdout)

}
