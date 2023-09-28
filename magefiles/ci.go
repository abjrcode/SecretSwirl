//go:build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Build(appVersion, buildTimestamp, commitSha, buildLink string) error {
	var ldFlags = fmt.Sprintf("-X 'main.Version=%s' -X 'main.BuildTimestamp=%s' -X 'main.CommitSha=%s' -X 'main.BuildLink=%s'", appVersion, buildTimestamp, commitSha, buildLink)
	var ouputFilename = fmt.Sprintf("swervo-%s-%s-%s", runtime.GOOS, runtime.GOARCH, appVersion)

	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		mg.Deps(mg.F(configureWailsProject, appVersion))

		fmt.Println("Building Wails App")

		var windowsOutputFilename = fmt.Sprintf("swervo-windows-amd64-%s.exe", appVersion)
		var buildWindows = sh.RunV("wails", "build", "-m", "-nosyncgomod", "-ldflags", ldFlags, "-nsis", "-platform", "windows/amd64", "-o", windowsOutputFilename)

		if buildWindows != nil {
			fmt.Println("Error building Wails App for Windows", buildWindows)
			return buildWindows
		}

		return sh.RunV("wails", "build", "-m", "-nosyncgomod", "-ldflags", ldFlags, "-platform", "linux/amd64", "-o", ouputFilename)
	} else if runtime.GOOS == "windows" || (runtime.GOOS == "linux" && runtime.GOARCH == "arm64") {
		mg.Deps(mg.F(configureWailsProject, appVersion))

		fmt.Println("Building Wails App")
		return sh.RunV("wails", "build", "-m", "-nosyncgomod", "-ldflags", ldFlags, "-o", ouputFilename)
	} else if runtime.GOOS == "darwin" {
		mg.Deps(mg.F(configureWailsProject, appVersion))

		fmt.Println("Building Wails App for Darwin")
		var buildDarwin = sh.RunV("wails", "build", "-m", "-nosyncgomod", "-ldflags", ldFlags, "-platform", "darwin/arm64,darwin/amd64")

		if buildDarwin != nil {
			fmt.Println("Error building Darwin Wails App", buildDarwin)
			return buildDarwin
		}

		fmt.Println("Building DMG")
		var amdDmgOutputPath = fmt.Sprintf("./build/bin/Swervo-amd64-%s.dmg", appVersion)
		var createAmdDmgError = sh.RunV("create-dmg", "--no-internet-enable", "--hide-extension", "Swervo-amd64.app", "--app-drop-link", "600", "200", amdDmgOutputPath, "./build/bin/Swervo-amd64.app")

		if createAmdDmgError != nil {
			fmt.Println("Error building DMG", createAmdDmgError)
			return createAmdDmgError
		}

		var armDmgOutputPath = fmt.Sprintf("./build/bin/Swervo-arm64-%s.dmg", appVersion)
		var createArmDmgError = sh.RunV("create-dmg", "--window-size", "800", "300", "--no-internet-enable", "--hide-extension", "Swervo-arm64.app", "--app-drop-link", "600", "40", armDmgOutputPath, "./build/bin/Swervo-arm64.app")

		if createArmDmgError != nil {
			fmt.Println("Error building DMG", createArmDmgError)
			return createArmDmgError
		}

		fmt.Println("Compiling seticon.swift")
		var swiftcError = sh.Run("swiftc", "./build/darwin/seticon.swift")
		if swiftcError != nil {
			fmt.Println("Error compiling seticon.swift", swiftcError)
			return swiftcError
		}

		var chmodError = sh.Run("chmod", "+x", "./seticon")
		if chmodError != nil {
			fmt.Println("Error setting permissions on seticon", chmodError)
			return chmodError
		}

		fmt.Println("Setting DMG icons")
		var setAmdIconError = sh.RunV("./seticon", "./build/bin/Swervo-amd64.app/Contents/Resources/iconfile.icns", amdDmgOutputPath)

		if setAmdIconError != nil {
			fmt.Println("Error setting AMD64 DMG icon", setAmdIconError)
			return setAmdIconError
		}

		return sh.RunV("./seticon", "./build/bin/Swervo-arm64.app/Contents/Resources/iconfile.icns", armDmgOutputPath)
	} else {
		return fmt.Errorf("Unsupported OS/architecture: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func configureWailsProject(releaseVersion string) error {
	var r, error = regexp.Compile("^v([^-]+)-(.+)$")

	if error != nil {
		fmt.Println("Error compiling regex", error)
		return error
	}

	var nsisCompliantVersion = r.ReplaceAllString(releaseVersion, "$1.$2")
	fmt.Printf("NSIS compatible version: [%s]\n", nsisCompliantVersion)

	type WailsProjectConfigAuthor struct {
		Name string `json:"name"`
	}

	type WailsProjectConfigInfo struct {
		CompanyName    string `json:"companyName"`
		ProductVersion string `json:"productVersion"`
		Copyright      string `json:"copyright"`
		Comments       string `json:"comments"`
	}

	type WailsProjectConfig struct {
		Schema               string                   `json:"$schema"`
		Name                 string                   `json:"name"`
		OutputFilename       string                   `json:"outputfilename"`
		FrontendInstall      string                   `json:"frontend:install"`
		FrontendBuild        string                   `json:"frontend:build"`
		FrontendDevWatcher   string                   `json:"frontend:dev:watcher"`
		FrontendDevServerUrl string                   `json:"frontend:dev:serverUrl"`
		Author               WailsProjectConfigAuthor `json:"author"`
		Info                 WailsProjectConfigInfo   `json:"info"`
	}

	fmt.Println("Reading Wails Config")
	var wailsConfigJson, read_error = os.ReadFile("wails.json")

	if read_error != nil {
		fmt.Println("Error reading wails.json", read_error)
		return read_error
	}

	var wailsConfig WailsProjectConfig

	var parse_error = json.Unmarshal(wailsConfigJson, &wailsConfig)

	if parse_error != nil {
		fmt.Println("Error parsing wails.json", parse_error)
		return parse_error
	}

	fmt.Println("Setting Wails Product Version")
	wailsConfig.Info.ProductVersion = nsisCompliantVersion

	var updatedWailsConfig, marshal_error = json.MarshalIndent(wailsConfig, "", "  ")

	if marshal_error != nil {
		fmt.Println("Error marshalling wails.json", marshal_error)
		return marshal_error
	}

	fmt.Println("Writing Wails Config")
	return os.WriteFile("wails.json", updatedWailsConfig, os.ModePerm)
}
