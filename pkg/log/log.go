package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Debugf output the debug message to console
// Deprecated: Use zap.L().Debug() instead
func Debugf(format string, args ...interface{}) {
	zap.L().Debug(fmt.Sprintf(format, args...))
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Infof output the log message to console
// Deprecated: Use zap.L().Info() instead
func Infof(format string, args ...interface{}) {
	zap.L().Info(fmt.Sprintf(format, args...))
	fmt.Printf(format+"\n", args...)
}

// Warnf output the warning message to console
// Deprecated: Use zap.L().Warn() instead
func Warnf(format string, args ...interface{}) {
	zap.L().Warn(fmt.Sprintf(format, args...))
	// _, _ = colorutil.ColorWarningMsg.Fprintf(os.Stderr, format+"\n", args...)
}

// Errorf output the error message to console
// Deprecated: Use zap.L().Error() instead
func Errorf(format string, args ...interface{}) {
	zap.L().Error(fmt.Sprintf(format, args...))
	// _, _ = colorutil.ColorErrorMsg.Fprintf(os.Stderr, format+"\n", args...)
}
