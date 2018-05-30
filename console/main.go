package console

import (
	"flag"
)

var (
	cmdHelp = flag.Bool("h", false, "帮助")
)

func init() {
	flag.Parse()
}

func main() {
	if *cmdHelp {
		flag.PrintDefaults()
		return
	}

}
