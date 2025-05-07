package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/barasher/go-exiftool"
)

type PhotoIndex map[string][]string  // Change structure to map[file]personnames
type PeopleIndex map[string][]string // Change structure to map[personname]files

// Extracts tags related to people from image files and returns a PhotoIndex with person names as keys and associated files as values.
func extractPeopleTags(inputDir string) (PhotoIndex, PeopleIndex, error) {
	log.Printf("Extracting XMP tags from: %s", inputDir)
	files, err := filepath.Glob(filepath.Join(inputDir, "*.*"))
	if err != nil {
		return nil, nil, err
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
				}
			}
		}
		if len(peopleTags) == 0 {
			peopleIndex["no one"] = append(peopleIndex["no one"], fileMetadata.File)
			log.Printf("No people tags found in file: %s", fileMetadata.File)
			continue
		}
		// If tags were found, add them to the PhotoIndex
		for _, person := range peopleTags {
			// Add file to the list of files for each person
			peopleIndex[person] = append(peopleIndex[person], fileMetadata.File)
		}
		photoIndex[fileMetadata.File] = peopleTags
	}

	// Log photoIndex for debugging
	for person, files := range peopleIndex {
		log.Printf("Person: %s, Files: %v", person, files)
	}

	return photoIndex, peopleIndex, nil
}
