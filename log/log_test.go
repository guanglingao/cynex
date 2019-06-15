package log

import "testing"

func TestDebug(t *testing.T) {
	Debug("this is debug output.")
}

func TestError(t *testing.T) {
	Error("This is error output.")
}

func TestInfo(t *testing.T) {
	Info("This is info output")
}

func TestWarning(t *testing.T) {
	Warning("This is warning output")
}
