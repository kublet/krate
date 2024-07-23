package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// Krate struct to hold all the things needed
type Krate struct {
	username       string
	device_address string
	path           string
	cmd            string
}

const tftSetup string = `
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

const mainCpp string = `
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
  Serial.println("Drew hello");
}

void loop() {
  if((WiFi.status() == WL_CONNECTED)) {
    otaserver.handle(); // DO NOT EDIT
  }

  delay(1);
}
`

const pioini string = `
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
	kublet/KGFX@^0.0.14
	kublet/OTAServer@^1.0.4
monitor_speed = 460800
`

/**
 * Initialize the Krate struct
 */
func (k *Krate) Initialize() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	k.username = user.Username     // get the current user
	if runtime.GOOS != "windows" { // Windows Path
		k.path = filepath.Join(user.HomeDir, ".platformio", "penv", "bin")
	} else { // Non-Windows Path
		k.path = filepath.Join(user.HomeDir, ".platformio", "penv", "Scripts")
	}
	k.cmd = "pio" // the command to run
	k.device_address = ""
	return nil
}

/**
* print help message
 */
func (k *Krate) Help() {
	fmt.Println("Usage: krate [options...] <arg>")
	fmt.Println(" help                    List all available commands")
	fmt.Println(" init                    Initializes project with basic libraries and code")
	fmt.Println(" deps install            Install dependencies specified in platformio.ini file")
	fmt.Println(" build                   Compiles project")
	fmt.Println(" send <ip address>       Send firmware OTA to kublet by specifiying ip address")
	fmt.Println(" monitor                 Monitors logs")
	fmt.Println(" publish									Creates manifest file based on prompts and user inputs for publishing on Kublet community")
}

func (k *Krate) initProjFiles(folder string) {
	k.initFile(filepath.Join(folder, "src", "main.cpp"), mainCpp)
	k.initFile(filepath.Join(folder, "platformio.ini"), pioini)
}

/**
 * Initialize project
 */
func (k *Krate) initProj(args []string) {
	var folderName string
	if len(args) > 1 {
		folderName = args[1]
	} else {
		log.Fatal("krate: please specify name of directory")
	}
	k.initProjFiles(folderName)
	fmt.Printf("Initialized files in %s:\n", folderName)
	fmt.Println("    src/")
	fmt.Println("    platformio.ini")
	fmt.Println("")
	fmt.Printf("cd into %s project and begin development\n", folderName)
	fmt.Println("Happy coding!")
}

/**
 * Install dependencies
 */
func (k *Krate) installDeps() {
	cmd := exec.Command(filepath.Join(k.path, k.cmd), "lib", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (k *Krate) initFile(fpath, v string) {
	f, err := k.create(fpath)
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

func (k *Krate) editFiles() {
	fpath := filepath.Join(".pio", "libdeps", "esp32dev", "TFT_eSPI", "User_Setup.h")
	k.initFile(fpath, tftSetup)
}

/**
 * Build project
 */
func (k *Krate) buildProj() {
	k.editFiles()

	cmd := exec.Command(filepath.Join(k.path, k.cmd), "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}


func (k *Krate) create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
func (k *Krate) sendFileOTA(args []string) {
	ip := ""
	if len(args) > 1 {
		ip = args[1]
		pip := net.ParseIP(ip)
		if pip == nil {
			log.Fatal("krate: invalid IP address")
			// fmt.Println("krate: invalid IP address")
			// os.Exit(1)
		}
	} else {
		ip = os.Getenv("KUBLET_IP_ADDR")
		pip := net.ParseIP(ip)
		if pip == nil {
			log.Fatal("krate: invalid IP address. Set env by running export KUBLET_IP_ADDR=<IP Addr>")
			// fmt.Println("krate: invalid IP address. Set env by running export KUBLET_IP_ADDR=<IP Addr>")
			// os.Exit(1)
		}
	}
	k.device_address = ip

	fpath := filepath.Join(".pio", "build", "esp32dev", "firmware.bin")

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

/**
 * Monitor
 */
func (k *Krate) monitor() {
	cmd := exec.Command(filepath.Join(k.path, k.cmd), "device", "monitor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}
