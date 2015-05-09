// dcompress_test.go

package dcompress

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/hotei/mdr"
)

func Test_000(t *testing.T) { // template
	if false {
		t.Errorf("print fail, but keep testing")
	}
	fmt.Printf("Test_0000 (%d)\n")
}

//
var filePair1 = [2]string{"testdata/kermit.Z",
	"d61611d13775c1f3a83675e81afcadfc4352b11e0f39f7c928bad62d25675b66"}

var filePairs = [][2]string{filePair1 /*, filePair2,  filePair3*/}

func Test_001(t *testing.T) {
}
