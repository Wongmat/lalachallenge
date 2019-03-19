package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	semver.Sort(releases)
	var versionSlice []*semver.Version
	var major int64 = 0 //keeps track of current sequence major
	var minor int64 = 0 //keeps track of current sequence minor
	for i := len(releases) - 1; i >= 0 && releases[i].Compare(*minVersion) > -1; i-- {
		fmt.Println(releases[i])
		if releases[i].Major != major || releases[i].Minor != minor {
			releases[i].PreRelease = ""
			versionSlice = append(versionSlice, releases[i])
			major = releases[i].Major
			minor = releases[i].Minor
		}
	}
	return versionSlice
}

//Reads from the provided file until scanner.Scan() == false
func Reader(location string) map[string]string {
	var output = make(map[string]string)
	file, err := os.Open(location)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	scanner.Scan()
	for scanner.Scan() == true {
		line := strings.Split(scanner.Text(), ",")
		names := line[0]
		version := line[1]
		fmt.Println("READ: " + names + " Min. Version: " + version)
		output[names] = version

	}
	file.Close()
	return output
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	//get filepath input + IO
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter file path: ")
	path, _ := reader.ReadString('\n')
	path = path[:len(path)-1]
	fmt.Println("Reading from " + path + "...")
	list := Reader(path)

	for names, version := range list {
		// Github
		owner := strings.Split(names, "/")[0]
		software := strings.Split(names, "/")[1]
		client := github.NewClient(nil)
		ctx := context.Background()
		opt := &github.ListOptions{PerPage: 10}
		releases, _, err := client.Repositories.ListReleases(ctx, owner, software, opt)
		if err != nil {
			log.Output(0, names+": "+err.Error())
			continue
		}
		//Process releases
		minVersion := semver.New(version)
		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}
		//Check if releases were found before latestVersions() + stdout. If not, stdout that software probably doesn't follow semver
		if len(allReleases) != 0 {
			versionSlice := LatestVersions(allReleases, minVersion)

			fmt.Printf("Latest versions of %s/%s: %s \n", owner, software, versionSlice)
		} else {
			fmt.Printf("No releases found for %s. Please ensure that this software follows Semantic Versioning format.\n", names)
		}
	}
}
