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
	hFlag := flag.Bool("h", false, "Help")
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
top for Redis, group by command and client IP

Usage:
  redistop [[localhost:6379] password]
Options:
  -f 2s : Refresh frequency
  -h : Help
  -V : Version

You can set REDISTOP_PASSWORD
`, version.Version())
		return
	}
	if *vFlag {
		fmt.Println(version.Version())
		return
	}

	host := "localhost:6379"
	args := flag.Args()
	if len(args) > 1 {
		host = args[1]
	}
	var password string
	if len(args) > 2 {
		password = args[2]
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
