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

License
-------
The 'dcompress' go package is distributed under the Simplified BSD License:

> Copyright (c) 2015 David Rook. All rights reserved.
> 
> Redistribution and use in source and binary forms, with or without modification, are
> permitted provided that the following conditions are met:
> 
>    1. Redistributions of source code must retain the above copyright notice, this list of
>       conditions and the following disclaimer.
> 
>    2. Redistributions in binary form must reproduce the above copyright notice, this list
>       of conditions and the following disclaimer in the documentation and/or other materials
>       provided with the distribution.
> 
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDER ``AS IS'' AND ANY EXPRESS OR IMPLIED
> WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
> FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> OR
> CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
> CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
> SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
> ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
> NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
> ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

Documentation (c) 2015 David Rook 

// EOF README-dcompress.md  (this is a markdown document and tested OK with blackfriday)
