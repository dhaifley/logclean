// Logclean - A small utility to purge old data from the log dataabse
package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"sync"
	"time"
)

const version = "1.1.1"

func showUsage() {
	log.Println(`Usage: logclean [options] index_name
Options:
	-h, --help			Shows this usage information
	-a, --age			Sets maximum log entry age in days (default is 30)
	-l, --log			Logs output to syslog`)
}

func main() {
	var age int
	var showHelp, logSys bool
	flag.BoolVar(&showHelp, "h", false, "Show this usage information")
	flag.BoolVar(&showHelp, "help", false, "Show this usage information")
	flag.IntVar(&age, "a", 30, "Sets maximum log entry age in days")
	flag.IntVar(&age, "age", 30, "Sets maximum log entry age in days")
	flag.BoolVar(&logSys, "l", false, "Logs output to syslog")
	flag.BoolVar(&logSys, "log", false, "Logs output to syslog")
	flag.Usage = showUsage
	flag.Parse()

	var slw *syslog.Writer
	if logSys {
		var err error
		slw, err = syslog.New(syslog.LOG_NOTICE, "logclean")
		if err != nil {
			log.Println("Error: Unable to attach syslog")
		} else {
			defer slw.Close()
			log.SetOutput(slw)
		}
	}

	log.Println("logclean - version:", version)
	if showHelp || flag.NArg() < 1 {
		showUsage()
		return
	}

	cp := "Command parameters:"
	flag.Visit(func(f *flag.Flag) {
		cp += fmt.Sprintf(" %s: %v", f.Name, f.Value)
	})
	for _, a := range flag.Args() {
		cp += fmt.Sprintf(" %v", a)
	}
	log.Println(cp)
	log.Println("Begin processing")

	cli := &http.Client{
		Timeout: time.Second * 30,
	}
	ec := ELKClient{
		Index:  flag.Arg(0),
		Age:    age,
		Client: cli,
	}

	var wg sync.WaitGroup
	ci := ec.GetIndexes()
	for ri := range ci {
		if ri.Err != nil {
			log.Printf("Error: %s\n", ri.Err.Error())
			continue
		}

		wg.Add(1)
		go func(index string) {
			rd := ec.DeleteIndex(index)
			if rd.Err != nil {
				log.Printf("Error: %s\n", rd.Err.Error())
				return
			}

			log.Printf("Deleted Index: %s\n", rd.Msg)
			wg.Done()
		}(ri.Msg)
	}

	wg.Wait()
	log.Println("End processing")
}
