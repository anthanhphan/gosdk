package logger

import (
	"os"
	"os/exec"
	"testing"
)

func runFatalTest(t *testing.T, testName string) {
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), "BE_FATAL_TEST=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		// Test passed successfully, it exited with code 1
		return
	}
	t.Fatalf("process ran with err %v, want exit error", err)
}

func TestLogger_Fatal_Subprocess(t *testing.T) {
	if os.Getenv("BE_FATAL_TEST") == "1" {
		logger := NewLogger(&Config{LogLevel: LevelDebug}, nil)
		logger.Fatal("fatal message")
		return // Should not reach here
	}
	runFatalTest(t, "TestLogger_Fatal_Subprocess")
}

func TestLogger_Fatalf_Subprocess(t *testing.T) {
	if os.Getenv("BE_FATAL_TEST") == "1" {
		logger := NewLogger(&Config{LogLevel: LevelDebug}, nil)
		logger.Fatalf("fatal %s", "message")
		return // Should not reach here
	}
	runFatalTest(t, "TestLogger_Fatalf_Subprocess")
}

func TestLogger_Fatalw_Subprocess(t *testing.T) {
	if os.Getenv("BE_FATAL_TEST") == "1" {
		logger := NewLogger(&Config{LogLevel: LevelDebug}, nil)
		logger.Fatalw("fatal message", "key", "value")
		return // Should not reach here
	}
	runFatalTest(t, "TestLogger_Fatalw_Subprocess")
}

func TestGlobal_Fatalw_Subprocess(t *testing.T) {
	if os.Getenv("BE_FATAL_TEST") == "1" {
		InitLogger(&Config{LogLevel: LevelDebug})
		Fatalw("global fatal", "key", "value")
		return // Should not reach here
	}
	runFatalTest(t, "TestGlobal_Fatalw_Subprocess")
}
