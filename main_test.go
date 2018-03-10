package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHelloWorld(t *testing.T) {
	Convey("Testing", t, func() {
		So(1, ShouldEqual, 1)
	})
}
