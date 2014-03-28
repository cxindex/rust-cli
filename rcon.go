package main

import (
	"flag"
	"fmt"
	"github.com/james4k/rcon"
	"github.com/rakyll/globalconf"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func check(f *flag.Flag) {
	if len(f.Value.String()) == 0 && f.Name != "cmd" {
		log.Fatalln(f.Name, "is not set")
	}
}

var status = regexp.MustCompile("^hostname:")

func main() {
	listen := flag.Bool("listen", false, "listen")
	server := flag.String("server", "", "server")
	port := flag.String("port", "", "port")
	pass := flag.String("pass", "", "password")
	cmd := flag.String("cmd", "", "command")

	abs, _ := exec.LookPath(os.Args[0])
	l, _ := filepath.EvalSymlinks(abs)
	dir, _ := filepath.Split(l)

	opts := globalconf.Options{Filename: dir + "rcon.conf"}
	if conf, err := globalconf.NewWithOptions(&opts); err != nil {
		if err, ok := err.(*os.PathError); !ok {
			log.Fatalln(err)
		}
	} else {
		conf.ParseAll()
	}

	flag.Parse()
	flag.VisitAll(check)

	listenLogger := log.New(os.Stdout, "", log.LstdFlags)
start:
	s, err := rcon.Dial(*server+":"+*port, *pass)
	if err != nil {
		log.Println(err)
		if !*listen {
			os.Exit(1)
		}
		time.Sleep(5 * time.Second)
		goto start
	}
	defer s.Close()

	if !*listen {
		if _, err := s.Write(*cmd); err != nil {
			log.Fatalln(err)
		}
	}
	for {
		response, id, err := s.Read()
		if err != nil {
			log.Println(err)
			if !*listen {
				os.Exit(1)
			}
			goto start
		}
		if id != 0 { //empty anyway
			response = strings.Trim(response, "\n")
			if *listen && !status.MatchString(response) { //strip status flood
				listenLogger.Println(response)
			}
			if !*listen {
				fmt.Println(response)
			}
		}
		if len(response) != 4000 && id != 0 && !*listen {
			break
		} else if id == 0 && err == nil {
			continue
		}
	}
}
