<center>
# dcompress
</center>

## OVERVIEW

dcompress is a go (1.4.2 currently) package to decompress .Z files.

This is a <strike>conversion</strike> literal translation of (n)compress42.c into go.
The original program was a Unix utility to compress/dcompress files. Files
compressed with this utility have a .Z extension.  It's relatively rare as of
2015 since better compression utilities are available now. 

### WHY?

I converted this utility to go so I could use it with a program I wrote to look
inside tar and zip archives and gather data on the files therein. Some of the files in my
tars are in .Z format and are themselves tars (ie tarofsomedir.tar.Z.  I could
of course have uncompressed them manually, but that's not what programmers do.

### Installation

If you have a working go installation on a Unix-like OS:

> ```go get github.com/hotei/dcompress```

Will copy github.com/hotei/program to the first entry of your $GOPATH

or if go is not installed yet :

> ```cd DestinationDirectory```

> ```git clone https://github.com/hotei/dcompress.git```

### Features

* Designed to decompress files with .Z extension ('compress' output)
* Same interface as the other packages in the go stdlib compress directory (flate,bzip2 etc)

### Limitations

* package is read only, there is no plan modify it to write .Z files (gz and bz2 are
much better)

Comments can be sent to <hotei1352@gmail.com> or registered at github.com/hotei/dcompress


### Things to be aware of

*	Passes test suite with input of 9 KB, 500 KB and 22 MB in 0.75 seconds (4Ghz AMD64).
*	It will not work with .z extension (that would be the output from _pack_ , not _compress_)
*	Has not been tested on 32bit hardware due to lack of time/need/availability.
*	The software uses LZW.
*	LZW aka the [Lempel-Ziv-Welch algorithm][1] was patented.
That patent (US Patent 4,558,302) has expired in the US.
*	I have no plans for a function to create compressed files (with ext.Z). If you
are interested in one, see http://golang.org/pkg/compress/lzw/ for an LZW reader/writer.  
   *   It is my impression from reading the documentation that the
package compress/lzw only decodes up to 12 bit codes (const maxWidth=12). The Unix _compress_
utility uses up to 16 bit codes according to the documentation.

Typical usage to read a compressed file can be found in the example directory.


[1]: http://en.wikipedia.org/wiki/Lempel%E2%80%93Ziv%E2%80%93Welch "LZW at Wiki"
[2]: http://google.com/        "Google"


