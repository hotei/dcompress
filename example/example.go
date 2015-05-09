// example.go

package main

import (
	// currently go 1.4.2 std lib
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// local
	"github.com/hotei/dcompress"
	"github.com/hotei/mdr"
)

func main() {
	var filePair1 = [2]string{"kermit.Z",
		"d61611d13775c1f3a83675e81afcadfc4352b11e0f39f7c928bad62d25675b66"}

	var filePairs = [][2]string{filePair1 /*, filePair2,  filePair3*/}
	dcompress.Verbose = true
	for i := 0; i < len(filePairs); i++ {
		infname := filePairs[i][0]
		outsig := filePairs[i][1]
		fmt.Printf("\n working to dcompress %s\n", infname)

		r, err := os.Open(infname)
		if err != nil {
			log.Panicf("open file failed for", infname)
		}
		rdr, err := dcompress.NewReader(r)
		if err != nil {
			fmt.Printf("FAILED - err from NewReader\n")
			return
		}
		dBuf, err := ioutil.ReadAll(rdr)
		if err != nil {
			fmt.Printf("FAILED - err from rdr.ReadAll()\n")
			return
		}
		fmt.Printf("dcompress would create %d bytes in new file\n", len(dBuf))
		bufSig := mdr.BufSHA256(dBuf)
		fmt.Printf("dcompress buffer has sha256 sig of %s\n", bufSig)
		fmt.Printf("expected sha256 of                 %s\n", outsig)
		if bufSig != outsig {
			fmt.Printf("FAILED - output files sha256 not correct")
		} else {
			fmt.Printf("PASS \n")
		}
	}
}
