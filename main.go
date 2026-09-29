package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/athoune/redistop/cli"
	"github.com/athoune/redistop/version"
)

func main() {
	fFlag := flag.Duration("f", 2*time.Second, "Frequency")
	hFlag := flag.Bool("help", false, "Help")
	flag.BoolVar(hFlag, "h", false, "Help")
	vFlag := flag.Bool("V", false, "Version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *hFlag {
		fmt.Printf(`RedisTop %s
top for Valkey (and Redis), group by command and client IP

Usage:
  redistop [host:port] [password]
Options:
  -f 2s : Refresh frequency
  -h, -help : Help
  -V : Version
  -cpuprofile file : Write cpu profile to file

You can set REDISTOP_PASSWORD instead of passing the password
(it takes precedence over the command line argument).
`, version.Version())
		return
	}
	if *vFlag {
		fmt.Println(version.Version())
		return
	}

	host := "localhost:6379"
	args := flag.Args()
	if len(args) > 0 {
		host = args[0]
	}
	var password string
	if len(args) > 1 {
		password = args[1]
	}
	p := os.Getenv("REDISTOP_PASSWORD")
	if p != "" {
		password = p
	}

	app := cli.NewApp(&cli.AppConfig{
		Host:      host,
		Password:  password,
		Frequency: *fFlag,
	})

	err := app.Serve()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Bye")
	}

}
