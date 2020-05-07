// hexc -- a hex dumping utility that supports color

package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
	flag "github.com/ogier/pflag"
)

type colors struct {
	counter color.Color
	def     color.Color
	listing color.Color
	newl    color.Color
	num     color.Color
	other   color.Color
	punc    color.Color
	space   color.Color
}

type settings struct {
	color   colors
	start   int
	width   int
	decimal bool
	limit   int
	list    bool
	length  int
	unicode bool
	verbose bool
	legend  bool
}

func defaultColors() colors {
	self := colors{}
	self.counter = color.FgDefault
	self.def = color.FgDefault
	self.listing = color.FgYellow

	self.newl = color.BgRed
	self.num = color.FgMagenta
	self.other = color.FgGreen
	self.punc = color.BgYellow
	self.space = color.BgBlue
	return self
}

var codes = map[string][]byte{
	"punc": {'.', ',', '?', '!', ':', '&', '\'', '"', '$', '#', '@', '-', '_'},
	"num":  {'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
	"nlt":  {0x09, 0x0a, 0x0d}}

func main() {
	clr := defaultColors()

	// flags
	deciPtr := flag.BoolP("decimal", "d", false, "Show all numbers in decimal instead of hex")
	helpPtr := flag.BoolP("legend", "h", false, "Display a color legend before the dump")
	lmtPtr := flag.IntP("limit", "l", 0, "Limit the dump to an arbitrary number of bytes (0 = no limit)")
	startPtr := flag.IntP("start", "s", 0, "Choose which byte in the file to begin the dump, in decimal")
	textPtr := flag.BoolP("text", "t", false, "Show a text listing next to the dump")
	verbPtr := flag.BoolP("verbose", "v", false, "Show extra information at the top of the dump")
	widthPtr := flag.IntP("width", "w", 16, "Specify the width of the dump in bytes")
	flag.Parse()

	var setup settings
	setup.color = clr
	setup.decimal = *deciPtr
	setup.limit = *lmtPtr
	setup.legend = *helpPtr
	setup.list = *textPtr
	setup.start = *startPtr
	// setup.unicode = *uniPtr
	setup.verbose = *verbPtr
	setup.width = *widthPtr

	// args
	var args []string = flag.Args()
	var filename string
	if len(args) > 1 || len(args) < 1 {
		errMessage("Missing filename.", true, true)
	} else {
		filename = args[0]
	}

	// file io
	f, err := os.Open(filename)
	if err != nil {
		errMessage(err.Error(), false, true)
	}
	fs, err := f.Stat()
	if err != nil {
		errMessage("Could not get file stats.", false, true)
	}
	setup.length = int(fs.Size())
	buf := make([]byte, fs.Size())
	f.Read(buf)
	defer f.Close()

	hexPrint(buf, fs, setup)
}

// Pretty prints the hex dump with color
func hexPrint(buf []byte, stats os.FileInfo, s settings) {
	if s.start > len(buf) {
		errMessage("File is too short for start position.", false, true)
	}

	var cntfmt string
	var hord string
	if s.decimal {
		cntfmt = "%03d"
		hord = "d"
	} else {
		cntfmt = "%02x"
		hord = "h"
	}
	buf = buf[s.start:len(buf)]

	if s.limit > 0 {
		if s.limit <= len(buf) {
			buf = buf[:s.limit]
			s.length = s.limit
		} else {
			errMessage("Desired limit is too long for file given start position.", false, true)
		}
	}

	var clr colors = s.color
	var wid int = int(s.width)
	var count int = 0
	colw, _ := strconv.Atoi(cntfmt[2:3])
	var lines int = len(buf) / wid
	var remn int = len(buf) % wid

	if s.verbose {
		clr.def.Print(stats.ModTime().String() + " " + stats.Name() + ", size: " +
			strconv.FormatInt(stats.Size(), 10) + " bytes, showing: " + strconv.Itoa(s.length) + " bytes.")
		clr.def.Print("\n")
	}

	for count < lines {
		clr.counter.Printf("%06"+cntfmt[3:]+hord+"\t", count*wid)
		for _, chr := range buf[count*wid : count*wid+wid-1] {
			if isMember(chr, codes["punc"]) {
				clr.punc.Printf(cntfmt, chr)
			} else if isMember(chr, codes["num"]) {
				clr.num.Printf(cntfmt, chr)
			} else if chr < 0x21 {
				if isMember(chr, codes["nlt"]) {
					clr.newl.Printf(cntfmt, chr)
				} else {
					clr.space.Printf(cntfmt, chr)
				}
			} else {
				clr.other.Printf(cntfmt, chr)
			}
			clr.def.Print(" ")
		}

		if s.list {
			if s.unicode {
				// add unicode listing implementation
			} else {
				clr.def.Print("\t")
				for _, chr := range buf[count*wid : count*wid+wid-1] {
					if isMember(chr, codes["punc"]) {
						clr.punc.Print(string(chr))
					} else if isMember(chr, codes["num"]) {
						clr.num.Print(string(chr))
					} else if chr < 0x21 {
						if isMember(chr, codes["nlt"]) {
							clr.newl.Printf(" ")
						} else {
							clr.space.Printf(" ")
						}
					} else {
						clr.other.Print(string(chr))
					}
				}
			}
		}
		count++
		clr.def.Print("\n")
	}

	if remn > 0 {
		clr.counter.Printf("%06"+cntfmt[3:]+hord+"\t", count*wid)
		for _, chr := range buf[count*wid : count*wid+remn] {
			if isMember(chr, codes["punc"]) {
				clr.punc.Printf(cntfmt, chr)
			} else if isMember(chr, codes["num"]) {
				clr.num.Printf(cntfmt, chr)
			} else if chr < 0x21 {
				if isMember(chr, codes["nlt"]) {
					clr.newl.Printf(cntfmt, chr)
				} else {
					clr.space.Printf(cntfmt, chr)
				}
			} else {
				clr.other.Printf(cntfmt, chr)
			}
			clr.def.Print(" ")
		}
		for i := 0; i < wid-remn-1; i++ {
			clr.def.Print(strings.Repeat(" ", colw+1))
		}

		if s.list {
			if s.unicode {
				// add unicode listing implementation
			} else {
				clr.def.Print("\t")
				for _, chr := range buf[count*wid : count*wid+remn] {
					if isMember(chr, codes["punc"]) {
						clr.punc.Print(string(chr))
					} else if isMember(chr, codes["num"]) {
						clr.num.Print(string(chr))
					} else if chr < 0x21 {
						if isMember(chr, codes["nlt"]) {
							clr.newl.Printf(" ")
						} else {
							clr.space.Printf(" ")
						}
					} else {
						clr.other.Print(string(chr))
					}
				}
			}
		}
	}
}

// Checks if a byte is a member of a byte array
func isMember(chr byte, arr []byte) bool {
	for _, mem := range arr {
		if mem == chr {
			return true
		}
	}
	return false
}

// General error handler
func errMessage(message string, defaults bool, exit bool) {
	color.FgRed.Print("Error: " + message)
	color.FgDefault.Print("\n")
	if defaults {
		flag.PrintDefaults()
	}
	if exit {
		os.Exit(1)
	}
}
