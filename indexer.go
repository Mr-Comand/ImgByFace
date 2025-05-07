package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/barasher/go-exiftool"
)

type PhotoIndex map[string][]string  // Change structure to map[file]personnames
type PeopleIndex map[string][]string // Change structure to map[personname]files

// Extracts tags related to people from image files and returns a PhotoIndex with person names as keys and associated files as values.
func extractPeopleTags(inputDir string) (PhotoIndex, PeopleIndex, error) {
	log.Printf("Extracting XMP tags from: %s", inputDir)

	var files []string
	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walking inputDir failed: %v", err)
	}

	// Initialize ExifTool
	exif, err := exiftool.NewExiftool()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize ExifTool: %v", err)
	}
	defer exif.Close()

	// Extract metadata from files
	data := exif.ExtractMetadata(files...)

	peopleIndex := make(PeopleIndex)
	photoIndex := make(PhotoIndex)
	for _, fileMetadata := range data {
		// Initialize a list of people associated with the file
		var peopleTags []string
		for name, value := range fileMetadata.Fields {
			// Check for XMP tags related to person names (RegionPersonDisplayName, RegionName)
			if name == "RegionPersonDisplayName" || name == "RegionName" {
				switch v := value.(type) {
				case []interface{}:
					for _, person := range v {
						peopleTags = append(peopleTags, person.(string))
						peopleIndex[person.(string)] = append(peopleIndex[person.(string)], fileMetadata.File)
					}
				case string:
					peopleTags = append(peopleTags, v)
					peopleIndex[v] = append(peopleIndex[v], fileMetadata.File)
				}
			}
		}
		if len(peopleTags) == 0 {
			peopleIndex["no one"] = append(peopleIndex["no one"], fileMetadata.File)
			photoIndex[fileMetadata.File] = []string{"no one"}
			log.Printf("No people tags found in file: %s", fileMetadata.File)
			continue
		}

		photoIndex[fileMetadata.File] = peopleTags
	}

	// Log photoIndex for debugging
	for person, files := range peopleIndex {
		log.Printf("Person: %s, Files: %v", person, files)
	}

	return photoIndex, peopleIndex, nil
}
