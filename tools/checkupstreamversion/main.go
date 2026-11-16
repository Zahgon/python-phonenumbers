// Copyright (C) 2009-2026 The Libphonenumber Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command checkupstreamversion compares this module's version with the latest
// google/libphonenumber GitHub release.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

const latestReleaseURL = "https://api.github.com/repos/google/libphonenumber/releases/latest"

var releaseTag = regexp.MustCompile(`^v([0-9.]+)$`)

func main() {
	releaseName, err := latestRelease()
	if err != nil {
		failConnection()
	}
	match := releaseTag.FindStringSubmatch(releaseName)
	if match == nil {
		writeAndExit(2, fmt.Sprintf("Unexpected upstream tag %s", releaseName))
	}
	upstreamVersion := match[1]
	if upstreamVersion == phonenumbers.Version {
		if _, err := fmt.Printf("Local version %s matches upstream version %s\n", phonenumbers.Version, upstreamVersion); err != nil {
			os.Exit(2)
		}
		return
	}
	if _, err := fmt.Printf("Local version %s does not match upstream version %s\n", phonenumbers.Version, upstreamVersion); err != nil {
		os.Exit(2)
	}
	os.Exit(1)
}

func latestRelease() (tag string, err error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "python-phonenumbers-upstream-check")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := response.Body.Close(); err == nil && closeErr != nil {
			err = closeErr
			tag = ""
		}
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", errors.New(response.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func failConnection() {
	writeAndExit(2, "Failed to connect to the GitHub API!")
}

func writeAndExit(code int, message string) {
	if _, err := fmt.Fprintln(os.Stderr, message); err != nil {
		os.Exit(2)
	}
	os.Exit(code)
}
