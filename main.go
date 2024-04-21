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
	"syscall"
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
Preferences preferences;

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
	kublet/KGFX@^0.0.8
	kublet/OTAServer@^1.0.4
monitor_speed = 460800
`

func initFile(fpath, v string) {
	f, err := os.Create(fpath)
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
		fmt.Println("krate: please specify name of folder")
		os.Exit(1)
	}

	initProjFiles(folderName)
}

func buildProj() {
	editFiles()

	binary, lookErr := exec.LookPath("pio")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"pio", "run"}

	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}

func monitor() {

	binary, lookErr := exec.LookPath("pio")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"pio", "device", "monitor"}

	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
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
		fmt.Println(" help                    Shows list of commands")
		fmt.Println(" send <firmware>         Send firmware OTA to kublet. If arg provided, file must end in .bin. Defaults to firmware.bin path in pio")

	case "send":
		sendFileOTA(args)

	case "init":
		initProj(args)

	case "build":
		buildProj()

	case "monitor":
		monitor()
	}

}
