// Package version provides version information for the application.
package version

// Version is the application version.
// Update this for each release.
const Version = "0.4.0"

// UserAgent returns the HTTP User-Agent string for the crawler.
func UserAgent() string {
	return "RoguePlanet/" + Version + " (+https://github.com/adewale/rogue_planet)"
}

// Generator returns the generator string for HTML output.
func Generator() string {
	return "Rogue Planet v" + Version
}
