package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/upstream"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type PyPiProductMetadataRaw struct {
	Releases map[string][]struct {
		PackageType string            `json:"packagetype"`
		URL         string            `json:"url"`
		UploadTime  string            `json:"upload_time_iso_8601"`
		Digests     map[string]string `json:"digests"`
	} `json:"releases"`
}

type PyPiRelease struct {
	version      *semver.Version
	SourceURL    string
	UploadTime   time.Time
	SourceSHA256 string
}

func (release PyPiRelease) Version() *semver.Version {
	return release.version
}

func getAllVersions() (versionology.VersionFetcherArray, error) {

	var pypiMetadata PyPiProductMetadataRaw
	err := upstream.GetAndUnmarshal("https://pypi.org/pypi/pip/json", &pypiMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve new versions from upstream: %w", err)
	}

	var allVersions versionology.VersionFetcherArray

	for version, releasesForVersion := range pypiMetadata.Releases {
		for _, release := range releasesForVersion {
			if release.PackageType != "sdist" {
				continue
			}

			fmt.Printf("Parsing semver version %s\n", version)

			newVersion, err := semver.NewVersion(version)
			if err != nil {
				continue
			}

			uploadTime, err := time.Parse(time.RFC3339, release.UploadTime)
			if err != nil {
				return nil, fmt.Errorf("could not parse upload time '%s' as date for version %s: %w", release.UploadTime, version, err)
			}

			allVersions = append(allVersions, PyPiRelease{
				version:      newVersion,
				SourceSHA256: release.Digests["sha256"],
				SourceURL:    release.URL,
				UploadTime:   uploadTime,
			})
		}
	}

	return allVersions, nil
}

func generateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()
	pipRelease, ok := versionFetcher.(PyPiRelease)
	if !ok {
		return nil, errors.New("expected a PyPiRelease")
	}

	configMetadataDependency := cargo.ConfigMetadataDependency{
		CPE:            fmt.Sprintf("cpe:2.3:a:pypa:pip:%s:*:*:*:*:python:*:*", version),
		ID:             "pip",
		Licenses:       retrieve.LookupLicenses(pipRelease.SourceURL, upstream.DefaultDecompress),
		PURL:           retrieve.GeneratePURL("pip", version, pipRelease.SourceSHA256, pipRelease.SourceURL),
		Source:         pipRelease.SourceURL,
		SourceChecksum: fmt.Sprintf("sha256:%s", pipRelease.SourceSHA256),
		Stacks:         []string{"*"},
		Version:        version,
	}

	return versionology.NewDependencyArray(configMetadataDependency, "noarch")
}

func main() {
	retrieve.NewMetadata("pip", getAllVersions, generateMetadata)
}
