// dcompress.go  (c) 2013 David Rook All rights reserved.

/*
Test suite passes
*/
package dcompress

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const g_version = "dcompress.go version 0.1.0"

// these need not be visible, lowercase them eventually
const (
	hSize     = 69001 // hash table size => 95% occupancy
	bitMask   = 0x1f
	blockMode = 0x80
	nBits     = uint8(16)
	//BUFSIZ     = 8192
	iBufSize = 8192
	oBufSize = iBufSize
	DEBUG    = 0
	BUG      = 0
)

var (
	VerboseFlag bool
	//g_cBytes    []byte // compressed bytes
	g_inbuf    []byte // unpacking temp area
	g_outbuf   []byte // output staging area
	MagicBytes = []byte{0x1f, 0x9d}

	// hashTable - [hSize] unsigned long  in original, but used with byte level access
	// using *--stackp as one example  go doesn't have pointer arith so we adapt
	g_htab    [hSize * 8]byte
	g_codetab [hSize]uint16 // codeTable must be uint16
)

func fatalErr(erx error) {
	log.Panicf("%v\n", erx)
	os.Exit(1)
}

func dumpIn(s string) {
	if len(g_inbuf) < 8 {
		log.Panicf("cant dump inbuf")
	}
	fmt.Printf("%s  => ", s)
	fmt.Printf("inbuf[0:100] \n")
	for j := 0; j < 4; j++ {
		for i := j * 25; i < (j+1)*25; i++ {
			fmt.Printf("%02x ", g_inbuf[i])
		}
		fmt.Printf(")\n")
	}
}

func dumpOut(s string) {
	if len(g_inbuf) < 8 {
		log.Panicf("cant dump outbuf")
	}
	fmt.Printf("%s  => ", s)
	fmt.Printf("outbuf[0:100] \n")
	for j := 0; j < 4; j++ {
		for i := j * 25; i < (j+1)*25; i++ {
			fmt.Printf("%02x ", g_outbuf[i])
		}
		fmt.Printf(")\n")
	}
}

func dumpHTAB(s string) {
	fmt.Printf("%s\n", s)
	for ndx, val := range g_htab {
		// if ndx < 255 { continue }	// these always equal self
		if val != 0 {
			fmt.Printf("htab[%d]=%02x %c\n", ndx, val, byte(val))
		}
	}
}

// NOTES:
//  Z is a compression technique, not an archiver.
// NewReader will create a local copy of the uncompressed data.
//Note that this will set upper limit on the size of an individual readable file.
//	Memory use expected to be < (CompressedFileSize * 10) for non-pathological cases.
//	Unpacks 23 MB in less than one second on 4 Ghz AMD 64 (8120^OC).
// Go code is based on literal translation of compress42.c (ie it's not idiomatic
// nor is it pretty) See doc.go for credits to the original writer(s)
//	Kludge is to fix a problem with first character being written to output buffer as zero always.
//	We save first output character and then patch it into outbuf at end.  Not sure why this happens.
//	Other than the kludge it's a very literal translation of compress42.c
//
// NewReader takes compressed data from input source r and returns a reader for the uncompressed version
func NewReader(r io.Reader) (io.Reader, error) {
	// NOTE BENE: sections that start with if DEBUG or if BUG will be removed by the compiler
	// since they compile to if false.  Leaving them in for now won't hurt anything.  Once code is
	// tested fully they can be stripped.

	type code_int int64

	type count_int int64

	var (
		INIT_BITS       = uint8(9)
		FIRST           = code_int(257)
		CLEAR           = code_int(256)
		maxbits         = nBits
		block_mode      = blockMode
		stackNdx        int
		code            code_int
		finchar         int
		oldcode         code_int
		incode          code_int
		inbits          int
		posbits         int
		outpos          int
		insize          int
		bitmask         int
		free_ent        code_int
		maxcode         code_int = (1 << nBits) - 1
		maxmaxcode      code_int = 1 << nBits
		n_bits          uint
		rsize           int
		bytes_in        int
		err             error
		ErrBadMagic     = errors.New("dcompress: Bad magic number ")
		ErrCorruptInput = errors.New("dcompress: Corrupt input ")
		ErrMaxBitsExcd  = errors.New("dcompress: maxbits exceeded ")
		ErrOther        = errors.New("dcompress: Other error :-( ")
		firstChar       byte
		codesRead       int
		outBuf          []byte // this is what we return
	)
	if DEBUG == 1 {
		fmt.Printf("entered decompress()\n")
	}

	g_inbuf = make([]byte, iBufSize+64, iBufSize+64)
	g_outbuf = make([]byte, oBufSize+2048, oBufSize+2048)
	outBuf = make([]byte, 0, 10000)
	if BUG == 1 {
		fmt.Printf("Sizeof(htab)= %d\n", len(g_htab)*1)
		fmt.Printf("Sizeof(codetab)= %d\n", len(g_codetab)*2)
		fmt.Printf("Sizeof(inbuf)= %d\n", len(g_inbuf)*1)
		fmt.Printf("Sizeof(outbuf)= %d\n", len(g_outbuf)*1)
	}

	if DEBUG == 1 {
		dumpIn("program startup")
	}

	rsize, err = r.Read(g_inbuf[0:iBufSize])
	if err != nil {
		fatalErr(err)
	}
	insize += rsize
	if DEBUG == 1 {
		dumpIn("after first read")
	}
	if DEBUG == 1 {
		fmt.Printf("insize(%d) rsize(%d)\n", insize, rsize)
	}
	if (g_inbuf[0] != MagicBytes[0]) || (g_inbuf[1] != MagicBytes[1]) {
		fatalErr(ErrBadMagic)
	}

	if DEBUG == 1 {
		fmt.Printf("nBits(%d) iBufSize(%d)\n", nBits, iBufSize)
	}

	maxbits = uint8(g_inbuf[2] & bitMask)
	block_mode = int(g_inbuf[2] & blockMode)
	maxmaxcode = (1 << maxbits)
	if DEBUG == 1 {
		fmt.Printf("maxbits(%d) block_mode(%d) maxmaxcode(%d)\n",
			maxbits, block_mode, maxmaxcode)
	}
	if maxbits > nBits {
		fatalErr(ErrMaxBitsExcd)
	}
	// --- line 1650 in compress42.c ---
	bytes_in = insize
	n_bits = uint(INIT_BITS)
	maxcode = (1 << n_bits) - 1
	bitmask = (1 << n_bits) - 1
	oldcode = -1
	finchar = 0
	outpos = 0
	posbits = 3 << 3

	if block_mode != 0 {
		free_ent = FIRST
	} else {
		free_ent = 256
	}
	if DEBUG == 1 {
		fmt.Printf("bytes_in(%d) maxcode(%d) bitmask(%d) posbits(%d) free_ent(%d)\n",
			bytes_in, int(maxcode), bitmask, posbits, int(free_ent))
	}
	if DEBUG == 1 {
		fmt.Printf("elements-in-codetab(%d) hSize(%d)\n", len(g_codetab), hSize)
	}

	// clear_tab_prefixof() =>  memset(codetab,0,256)  not req in go

	for code = 255; code >= 0; code-- {
		g_htab[code] = byte(code)
	}

	// D O    W H I L E
	for {
	resetbuf:
		{
			var (
				e int
				o int
			)
			o = posbits >> 3
			if o <= insize {
				e = insize - o
			} else {
				e = 0
			}
			if DEBUG == 1 {
				fmt.Printf("before o(%d) e(%d)\n", o, e)
			}
			for i := 0; i < e; i++ {
				g_inbuf[i] = g_inbuf[i+o]
			}
			if DEBUG == 1 {
				fmt.Printf("after o(%d) e(%d)\n", o, e)
			}
			insize = e
			posbits = 0
		}
		if DEBUG == 1 {
			fmt.Printf("iBufSize(%d)\n", iBufSize)
		}
		if DEBUG == 1 {
			dumpIn("after copy")
		}
		//																			? B U F E M P T Y
		if BUG == 1 {
			fmt.Printf("insize(%d) Sizeof(inbuf)-iBufSize = %d\n",
				insize, len(g_inbuf)-iBufSize)
		}
		// 																					R E A D
		if insize < (len(g_inbuf) - iBufSize) {
			rsize, err = r.Read(g_inbuf[insize : insize+iBufSize])
			if err != nil {
				if err == io.EOF {
					// not an error
				} else {
					fmt.Printf("rsize(%d), insize(%d) len(inbuf)- iBufSize(%d)\n",
						rsize, insize, len(g_inbuf)-iBufSize)
					fatalErr(err)
				}
			}
			if BUG == 1 {
				fmt.Printf("read() added rsize(%d) to insize(%d) )\n", rsize, insize)
				dumpIn("after read insize")
			}
			insize += rsize
		}
		//	inbits = ((rsize > 0) ? (insize - insize%n_bits)<<3 : (insize<<3)-(n_bits-1));

		if rsize > 0 {
			inbits = (insize - (insize % int(n_bits))) << 3
		} else {
			inbits = (insize << 3) - (int(n_bits - 1))
		}
		if BUG == 1 {
			fmt.Printf("rsize(%d) insize(%d) inbits(%d) posbits(%d) free_ent(%d) maxcode(%d)\n",
				rsize, insize, inbits, posbits, free_ent, maxcode)
		}

		//																					 W H I L E
		for { // while inbits > posbits
			if inbits <= posbits {
				break
			}

			if free_ent > maxcode {
				posbits = ((posbits - 1) + ((int(n_bits) << 3) - (posbits-1+(int(n_bits)<<3))%(int(n_bits)<<3)))
				if BUG == 1 {
					fmt.Printf("free_ent > maxcode => inbits(%d) posbits(%d) free_ent(%d) maxcode(%d)\n", inbits, posbits, free_ent, maxcode)
				}
				n_bits++
				if n_bits == uint(maxbits) {
					maxcode = maxmaxcode
				} else {
					maxcode = (1 << n_bits) - 1
				}
				bitmask = (1 << n_bits) - 1
				goto resetbuf
			}
			//  --- line 1715 in compress42.c ---
			if BUG == 1 {
				fmt.Printf("before input() posbits(%d) code(%02x) n_bits(%d) bitmask(%x) ",
					posbits, uint(code), n_bits, bitmask)
			}
			if DEBUG == 1 {
				dumpIn("before input of code")
			}

			/*	#define	input(inbuf,posbits,code,n_bits,bitmask){
				REG1 char_type 		*p = &(inbuf)[(posbits)>>3];			\
						(code) = ((((long)(p[0]))|((long)(p[1])<<8)|		\
						 ((long)(p[2])<<16))>>((posbits)&0x7))&(bitmask);	\
				(posbits) += (n_bits);										\
				}
			*/
			//																				  I N P U T
			//input(inbuf,posbits,code,n_bits,bitmask);
			nBufNdx := posbits >> 3
			if DEBUG == 1 {
				fmt.Printf("nBufNdx(%d)\n", nBufNdx)
			}
			p1 := uint(g_inbuf[nBufNdx])
			p2 := uint(g_inbuf[nBufNdx+1]) << 8
			p3 := uint(g_inbuf[nBufNdx+2]) << 16 // bad index on larger files
			t1 := p1 | p2 | p3
			t2 := t1 >> uint(posbits&0x7)
			posbits += int(n_bits)
			if DEBUG == 1 {
				fmt.Printf("nBufNdx(%d) posbits&0x7(%d) t1(%d) t2(%d) p1(%d) p2(%d) p3(%d)\n",
					nBufNdx, posbits&0x7, t1, t2, p1, p2, p3)
			}
			code = code_int(int(t2) & bitmask)
			codesRead++
			if BUG == 1 {
				fmt.Printf("after -> code(%02x) posbits(%d) codesRead(%d)\n", code, posbits, codesRead)
			}
			// BUG(mdr): <kludge alert>
			if firstChar == 0 {
				firstChar = byte(code)
			}
			// </kludge alert>

			if DEBUG == 1 {
				fmt.Printf("   after input() code(%02x) posbits(%d) outpos(%d) n_bits(%d) bitmask(%x)\n",
					code, posbits, outpos, n_bits, bitmask)
			}
			if oldcode == -1 {
				if code >= 256 {
					fatalErr(ErrOther)
				}
				oldcode = code
				finchar = int(oldcode)

				if DEBUG == 1 {
					fmt.Printf("finchar(%02x)\n", byte(finchar))
				}
				g_outbuf = append(g_outbuf, byte(finchar))
				outpos++
				continue
			}
			// 																				C L E A R
			if (code == CLEAR) && (block_mode != 0) {
				if BUG == 1 {
					fmt.Printf("CLEAR CODE\n")
				}
				// clear_tab_prefixof();#	define	clear_tab_prefixof()	memset(codetab, 0, 256);
				for i := 0; i < 256; i++ {
					g_codetab[i] = 0
				}
				free_ent = FIRST - 1
				posbits = ((posbits - 1) + ((int(n_bits) << 3) - (posbits-1+(int(n_bits)<<3))%(int(n_bits)<<3)))
				n_bits = uint(INIT_BITS)
				maxcode = (1 << n_bits) - 1
				bitmask = (1 << n_bits) - 1
				goto resetbuf
			}
			incode = code

			//		stackp = de_stack
			//#	define	de_stack				((char_type *)&(htab[hSize-1]))
			stackNdx = hSize*8 - 1 // index of last element of htab
			// 																					KwKwK
			if code >= free_ent { // BUG(mdr): original text ? Special case for KwKwK string
				if BUG == 1 {
					fmt.Printf("KwKwK\n")
				}
				if code > free_ent { // core dump no real help so dont print details
					//var p * byte
					//posbits -= int(n_bits)
					//p = &inbuf[posbits>>3]
					//fmt.Printf("insize:%d posbits:%d inbuf:%02X %02X %02X %02X %02X (%d)\n", insize, posbits,
					//	p[-1],p[0],p[1],p[2],p[3], (posbits&07))
					if VerboseFlag {
						fmt.Printf("!Err-> dcompress: code(%d) > free_ent(%d)\n", code, free_ent)
						fmt.Printf("!Err-> dcompress: insize(%d) posbits(%d)\n", insize, posbits-int(n_bits))
					}
					return nil, ErrCorruptInput
				}

				/* #define	htabof(i)				htab[i]
				#	define	codetabof(i)			codetab[i]
				#	define	tab_prefixof(i)			codetabof(i)
				#	define	tab_suffixof(i)			((char_type *)(htab))[i]
				#	define	de_stack				((char_type *)&(htab[hSize*8-1]))
				#	define	clear_htab()			memset(htab, -1, sizeof(htab))
				#	define	clear_tab_prefixof()	memset(codetab, 0, 256);
				*/
				// x := byte(finchar)
				// stackp--
				// *stackp = x
				// htab[]int64
				// finchar is int type trunc to char, stuffed into int64 arrary

				stackNdx--
				g_htab[stackNdx] = byte(finchar)
				code = oldcode
			}
			// Generate output characters in reverse order
			//while ((cmp_code_int)code >= (cmp_code_int)256){
			//	*--stackp = tab_suffixof(code);
			//	code = tab_prefixof(code);
			//}
			for {
				if code < 256 {
					break
				}

				//	*--stackp = htab[code]
				stackNdx--
				g_htab[stackNdx] = g_htab[code]
				code = code_int(g_codetab[code])
				if DEBUG == 1 {
					fmt.Printf("genout %02x %c\n", code, code)
				}
			}

			//*--stackp =	(char_type)(finchar = tab_suffixof(code));
			finchar = int(byte(g_htab[code]))
			stackNdx--
			g_htab[stackNdx] = byte(finchar)

			// seems ok up to here, output tokens match so far
			// compress42 has empty htab at end - we have a full one.  WHY?
			// --- line 1792
			//																		 O U T P U T
			{ // brace 2 stuff here...
				var i int
				i = (hSize*8 - 1) - stackNdx
				if BUG == 1 {
					fmt.Printf("outpos(%d)\n", outpos)
				}
				tmp := i + outpos
				if tmp >= oBufSize {
					if DEBUG == 1 {
						fmt.Printf("{2} i(%d)\n", i)
					}
					for { // do while ((i=de_stack-stackp))>0
						if i > (oBufSize - outpos) {
							i = oBufSize - outpos
						}
						//	void *memcpy(void *dest, const void *src, size_t n);
						if i > 0 {
							//  memcpy(outbuf+outpos, stackp, i)
							j := outpos
							k := 0
							for {
								if k >= i { // not likely but i could be zero
									break
								}
								if DEBUG == 1 {
									fmt.Printf("j(%d) stackNdx(%d) k(%d)\n", j, stackNdx, k)
								}
								g_outbuf[j] = byte(g_htab[stackNdx+k])
								j++
								k++
							}
							outpos += i
						}
						if outpos >= oBufSize { // output buffer needs to be flushed
							outBuf = append(outBuf, g_outbuf[0:outpos]...)
							//if (write(fdout, outbuf, outpos) != outpos) {
							//write_error()
							//}
							outpos = 0
						}
						stackNdx += i
						i = (hSize*8 - 1) - stackNdx
						if i <= 0 {
							break
						}
					} // end of dowhile
				} else {
					if DEBUG == 1 {
						fmt.Printf("{2} i(%d)\n", i)
					}

					//  memcpy(outbuf+outpos, stackp, i);
					j := outpos
					k := 0
					for {
						if k >= i { // not likely but i could be zero
							break
						}
						if DEBUG == 1 {
							fmt.Printf("j(%d) stackNdx(%d) k(%d)\n", j, stackNdx, k)
						}
						g_outbuf[j] = byte(g_htab[stackNdx+k])
						j++
						k++
					}
					outpos += i
				}
			} // brace 2 end

			// 																		N E W   E N T R Y
			//	Generate the new entry in code table
			code = free_ent
			if code < maxmaxcode {
				g_codetab[code] = uint16(oldcode) // codetab is uint16
				g_htab[code] = byte(finchar)
				free_ent = code + 1
			}
			oldcode = incode // remember previous code
		} // end of while inbits > posbits
		bytes_in += rsize
		if rsize <= 0 {
			break
		}
	} // end of do while (rsize > 0)
	// man2write -> ssize_t write(int fd, const void *buf, size_t count);

	// flush remaining output
	if outpos > 0 {
		outBuf = append(outBuf, g_outbuf[0:outpos]...)
	}
	if BUG == 1 {
		fmt.Printf("outpos(%d) \n", outpos)
	}
	if DEBUG == 1 {
		dumpIn("end of run")
		dumpOut("end of run")
		dumpHTAB("end of run")
	}
	// BUG(mdr): <kludge alert>
	outBuf[0] = firstChar
	// </kludge>
	byteReader := bytes.NewReader(outBuf)
	return byteReader, nil
} // end of ReadAll aka decompress()
