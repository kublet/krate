package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)


func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("krate: run `krate help` to see list of commands")
		os.Exit(1)
	}
	krate := Krate{}
	err := krate.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	action := args[0]
	switch action {
	case "help":
		krate.Help()
	case "init":
		krate.initProj(args)
	case "deps":
		if len(args) > 1 && args[1] != "install" {
				fmt.Println("krate: Run `krate deps install` to install dependencies")
				os.Exit(1)
		} else if len(args) == 1 {
			fmt.Println("krate: Run `krate deps install` to install dependencies")
			os.Exit(1)
		}
		krate.installDeps()
	case "build":
		krate.buildProj()
	case "send":
		krate.sendFileOTA(args)

	case "monitor":
		krate.monitor()
	case "publish":
		pub := Publish{}
		pub.Publish()
	default:
		fmt.Printf("krate: %s is not a valid command\n", action)
		fmt.Println("krate: run `krate help` to see list of commands")
	}

}
