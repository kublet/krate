package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var tftSetup string = `
#define USER_SETUP_INFO "User_Setup"
#define ST7789_DRIVER  
#define TFT_RGB_ORDER TFT_BGR
#define TFT_WIDTH  240
#define TFT_HEIGHT 240
#define TFT_MISO -1
#define TFT_MOSI 23
#define TFT_SCLK 18
#define TFT_CS 5  
#define TFT_DC 2
#define TFT_RST 4
#define TFT_BL 15
#define LOAD_GLCD   
#define LOAD_FONT2  
#define LOAD_FONT4
#define LOAD_FONT6
#define LOAD_FONT7
#define LOAD_FONT8
#define LOAD_GFXFF
#define SMOOTH_FONT
#define SPI_FREQUENCY 1000000
#define SPI_READ_FREQUENCY 27000000
#define SPI_TOUCH_FREQUENCY 2500000
`

var mainCpp string = `
#include <Arduino.h>
#include <otaserver.h>
#include <kgfx.h>

OTAServer otaserver;
KGFX ui;

void setup() {
  Serial.begin(460800);
  Serial.println("Starting app");

  otaserver.connectWiFi(); // DO NOT EDIT.
  otaserver.run(); // DO NOT EDIT

  ui.init();
  ui.clear();
  ui.drawText("hello", Arial_28, TFT_YELLOW, 0, 0);
}

void loop() {
  if((WiFi.status() == WL_CONNECTED)) {
    otaserver.handle(); // DO NOT EDIT
  }

  delay(1);
}
`

var pioini string = `
; PlatformIO Project Configuration File
;
;   Build options: build flags, source filter
;   Upload options: custom upload port, speed and extra flags
;   Library options: dependencies, extra library storages
;   Advanced options: extra scripting
;
; Please visit documentation for the other options and examples
; https://docs.platformio.org/page/projectconf.html

[env:esp32dev]
platform = espressif32
board = esp32dev
framework = arduino
lib_deps = 
	bodmer/TFT_eSPI@^2.5.0
	kublet/KGFX@^0.0.10
	kublet/OTAServer@^1.0.4
monitor_speed = 460800
`

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func initFile(fpath, v string) {
	f, err := create(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.WriteString(v); err != nil {
		log.Fatal(err)
	}

	w.Flush()
}

func sendFileOTA(args []string) {
	ip := ""
	if len(args) > 1 {
		ip = args[1]
		pip := net.ParseIP(ip)
		if pip == nil {
			fmt.Println("krate: invalid IP address")
			os.Exit(1)
		}
	} else {
		ip = os.Getenv("KUBLET_IP_ADDR")
		pip := net.ParseIP(ip)
		if pip == nil {
			fmt.Println("krate: invalid IP address. Set env by running export KUBLET_IP_ADDR=<IP Addr>")
			os.Exit(1)
		}
	}

	fpath := ".pio/build/esp32dev/firmware.bin"

	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	fw, err := writer.CreateFormFile("filedata", filepath.Base("firmware.bin"))
	if err != nil {
		log.Fatal(err)
	}
	fd, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Fatal(err)
	}

	writer.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://"+ip+"/update", form)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bodyText)
}

func initProjFiles(folder string) {
	initFile(filepath.Join(folder, "src/main.cpp"), mainCpp)
	initFile(filepath.Join(folder, "platformio.ini"), pioini)
}

func editFiles() {
	fpath := ".pio/libdeps/esp32dev/TFT_eSPI/User_Setup.h"
	initFile(fpath, tftSetup)
}

func initProj(args []string) {
	var folderName string
	if len(args) > 1 {
		folderName = args[1]
	} else {
		fmt.Println("krate: please specify name of directory")
		os.Exit(1)
	}

	initProjFiles(folderName)

	fmt.Printf("Initialized files in %s:\n", folderName)
	fmt.Println("    src/")
	fmt.Println("    platformio.ini")
	fmt.Println("")
	fmt.Printf("cd into %s project and begin development\n", folderName)
	fmt.Println("Happy coding!")
}

func installDeps() {
	cmd := exec.Command("pio", "lib", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}

func buildProj() {
	editFiles()

	cmd := exec.Command("pio", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}

func monitor() {
	cmd := exec.Command("pio", "device", "monitor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}

type config struct {
	name       string
	ID         string
	typ        string
	ddData     string
	defaultTxt string
}

func publish() {
	var (
		summary, desc, author, isCfg string
		cfgs                         []*config
	)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Summary (max 40 chars): ")
	scanner.Scan()
	summary = scanner.Text()

	fmt.Print("Description (max 150 chars): ")
	scanner.Scan()
	desc = scanner.Text()

	fmt.Print("Author (max 20 chars): ")
	scanner.Scan()
	author = scanner.Text()

	for {
		fmt.Print("Configs (yes/no): ")
		scanner.Scan()
		isCfg = scanner.Text()
		if isCfg == "yes" {
			cfg := &config{}

			fmt.Printf("ID (the key stored in nvs/preferences): ")
			scanner.Scan()
			cfg.ID = scanner.Text()

			fmt.Printf("Name (the displayed name/label of the config): ")
			scanner.Scan()
			cfg.name = scanner.Text()

			fmt.Printf("Type (text, dropdown_search): ")
			scanner.Scan()
			cfg.typ = scanner.Text()

			if cfg.typ == "text" {
				fmt.Printf("Default text (the placeholder text in the textbox). Press enter if none: ")
				scanner.Scan()
				cfg.defaultTxt = scanner.Text()

			} else if cfg.typ == "dropdown_search" {
				fmt.Printf("Dropdown data (comma separated): ")
				scanner.Scan()
				cfg.ddData = scanner.Text()
			}
			cfgs = append(cfgs, cfg)
			continue

		} else {
			break
		}
	}

	f, err := os.Create("manifest.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.WriteString("summary: " + summary + "\n"); err != nil {
		log.Fatal(err)
	}
	if _, err := w.WriteString("desc: " + desc + "\n"); err != nil {
		log.Fatal(err)
	}
	if _, err := w.WriteString("author: " + author + "\n"); err != nil {
		log.Fatal(err)
	}

	if _, err := w.WriteString("configs:\n"); err != nil {
		log.Fatal(err)
	}

	for _, c := range cfgs {
		if _, err := w.WriteString("  id: " + c.ID + "\n"); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteString("  name: " + c.name + "\n"); err != nil {
			log.Fatal(err)
		}
		if _, err := w.WriteString("  type: " + c.typ + "\n"); err != nil {
			log.Fatal(err)
		}
		if c.defaultTxt != "" {
			if _, err := w.WriteString("  default_text: " + c.defaultTxt + "\n"); err != nil {
				log.Fatal(err)
			}
		}
		if c.ddData != "" {
			if _, err := w.WriteString("  dropdown: " + c.ddData + "\n"); err != nil {
				log.Fatal(err)
			}
		}

	}

	w.Flush()

}

func main() {

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("krate: run `krate help` to see list of commands")
		os.Exit(1)
	}

	action := args[0]

	switch action {
	case "help":
		fmt.Println("Usage: krate [options...] <arg>")
		fmt.Println(" help                    List all available commands")
		fmt.Println(" init                    Initializes project with basic libraries and code")
		fmt.Println(" deps install            Install dependencies specified in platformio.ini file")
		fmt.Println(" build                   Compiles project")
		fmt.Println(" send <ip address>       Send firmware OTA to kublet by specifiying ip address")
		fmt.Println(" monitor                 Monitors logs")
		fmt.Println(" publish									Creates manifest file based on prompts and user inputs for publishing on Kublet community")

	case "init":
		initProj(args)

	case "deps":
		if len(args) > 1 {
			if args[1] != "install" {
				fmt.Println("krate: Run `krate deps install` to install dependencies")
				os.Exit(1)
			}
		} else {
			fmt.Println("krate: Run `krate deps install` to install dependencies")
			os.Exit(1)
		}
		installDeps()

	case "build":
		buildProj()

	case "send":
		sendFileOTA(args)

	case "monitor":
		monitor()

	case "publish":
		publish()
	}

}
