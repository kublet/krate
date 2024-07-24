package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type Publish struct {
	
}
type config struct {
	name       string
	ID         string
	typ        string
	ddData     string
	defaultTxt string
}

func (p *Publish) Publish() {
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