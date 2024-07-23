package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitFile(t *testing.T) {
	k := Krate{}
	k.initFile("test.go", "test")
	as := assert.New(t)
	as.FileExists("test.go")
	os.Remove("test.go")
}

func TestCreate(t *testing.T) {
	k := Krate{}
	f, err := k.create("test.go")
	as := assert.New(t)
	as.Nil(err)
	as.NotNil(f)
	os.Remove("test.go")
}

func TestEditFiles(t *testing.T) {
	k := Krate{}
	k.editFiles()
	as := assert.New(t)
	as.FileExists(filepath.Join(".pio", "libdeps", "esp32dev", "TFT_eSPI", "User_Setup.h"))
	os.RemoveAll(".pio")
}

func TestBuildProj(t *testing.T) {
	k := Krate{}
	k.buildProj()
	as := assert.New(t)
	as.DirExists(".pio")
	os.RemoveAll(".pio")
}

func TestInstallDeps(t *testing.T) {
	k := Krate{}
	k.installDeps()
	as := assert.New(t)
	as.DirExists(".pio")
	os.RemoveAll(".pio")
}

func TestInitProj(t *testing.T) {
	k := Krate{}
	k.initProj([]string{"init", "_test"})
	as := assert.New(t)
	as.DirExists("_test")
	as.FileExists(filepath.Join("_test", "platformio.ini"))
	as.DirExists(filepath.Join("_test", "src"))
	as.FileExists(filepath.Join("_test", "src", "main.cpp"))
	os.RemoveAll("_test")
}