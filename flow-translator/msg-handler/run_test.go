/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    run_test.go
 * details: Deals with the setup and teardown for Unit Tests for msghandler package
 *
 */

package msghandler

import (
	"log"
	"os"
	"testing"

	opts "github.com/Juniper/collector/flow-translator/options"
)

func VerifyError(name string, t *testing.T, expected interface{}, result interface{}) {
	if expected != result {
		t.Errorf("%s failed, expected '%v', got '%v'", name, expected, result)
	}
}

func setup() {
	InitMockData()
	opts.Verbose = true
	opts.Logger = log.New(os.Stderr, "[jFlow] ", log.Ldate|log.Ltime)
}

func shutdown(retCode int) {
	log.Println("Test Done!!!")
	os.Exit(retCode)
}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	shutdown(retCode)
}
