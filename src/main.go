package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/cavaliercoder/grab"
)

const pineappleSrc string = "https://github.com/pineappleEA/pineapple-src/"
const pineappleSite string = "https://raw.githubusercontent.com/pineappleEA/pineappleEA.github.io/master/index.html"

//TODO: set path with settings inside app
const installPath string = "."

func main() {
	a := app.New()
	w := a.NewWindow("PinEApple Updater")
	w.SetIcon(resourceIconPng)
	versionSlice, linkMap := downloadList()
	w.SetContent(loadUI(versionSlice, linkMap))
	w.Resize(fyne.NewSize(500, 450))
	w.Show()
	a.Run()
}

func downloadList() ([]int, map[int]string) {
	//return variables
	linkMap := make(map[int]string)
	versionSlice := make([]int, 0)

	//download site into resp
	resp, err := http.Get(pineappleSite)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	//read response body through scanner
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan(); i++ {
		var line = scanner.Text()
		match, _ := regexp.MatchString("https://anonfiles.com", line)
		if match {
			// extract link
			linkPattern, _ := regexp.Compile("https://anonfiles.com/.*/YuzuEA-[0-9]*_7z")
			link := linkPattern.FindString(scanner.Text())

			// extract version number
			versionPattern, _ := regexp.Compile("EA [0-9]*")
			versionString := versionPattern.FindString(scanner.Text())
			numberPattern, _ := regexp.Compile("[0-9]*$")
			versionString = numberPattern.FindString(versionString)
			version, _ := strconv.Atoi(versionString)

			//save link in map
			linkMap[version] = link
			//add version number to slice
			versionSlice = append(versionSlice, version)

		} else if line == "</html>" {
			break
		}
	}
	return versionSlice, linkMap
}

func install(versionSlice []int, linkMap map[int]string, selectedVersion int) {
	resp, err := http.Get(pineappleSrc + "releases/download/EA-" + strconv.Itoa(versionSlice[selectedVersion]) + "/Windows-Yuzu-EA-" + strconv.Itoa(versionSlice[selectedVersion]) + ".7z")
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	var downloadLink string
	if resp.StatusCode == 200 {
		// Downloading from Github
		downloadLink = pineappleSrc + "releases/download/EA-" + strconv.Itoa(versionSlice[selectedVersion]) + "/Windows-Yuzu-EA-" + strconv.Itoa(versionSlice[selectedVersion]) + ".7z"
	} else {
		//Download from Anonfiles
		//Download Anonfiles page to grab direct download
		resp, err := http.Get(linkMap[versionSlice[selectedVersion]])
		if err != nil {
			// handle err
		}
		//go line through line and search for direct download link with regex
		//TODO: fail safely in case no links can be found
		scanner := bufio.NewScanner(resp.Body)
		for i := 0; scanner.Scan(); i++ {
			linkPattern, _ := regexp.Compile("https://cdn-.*anonfiles.*7z")
			if linkPattern.MatchString(scanner.Text()) {
				downloadLink = linkPattern.FindString(scanner.Text())
				break
			}
		}
		defer resp.Body.Close()
	}
	downloadFile(downloadLink)
}

func downloadFile(link string) {
	client := grab.NewClient()
	req, _ := grab.NewRequest(installPath, link)
	resp := client.Do(req)
	downloadUI(resp)

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

}
