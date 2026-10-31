package utilities

import (
	"log"
	"strconv"
	"strings"

	"solana/classes/context"
)

// Parse extracts the build ID and build string from the provided user agent string.
// If both the build ID and build string are successfully extracted, it processes
// the user agent information and returns a pointer to a SeasonInfo object.
// If extraction fails, it logs an error message and returns nil.
//
// Parameters:
//   - userAgent: A string representing the user agent to be parsed.
//
// Returns:
//   - A pointer to a context.SeasonInfo object if parsing is successful, otherwise nil.
func Parse(userAgent string) *context.SeasonInfo {
	buildID := getBuildID(userAgent)
	buildString := getBuildString(userAgent)

	if buildID != "" && buildString != "" {
		return handleValidBuild(&context.UserAgent{
			VersionId:  buildID,
			VersionStr: buildString,
		}, getPlatform(userAgent))
	}

	log.Printf("Failed to parse Build ID and Build String from user agent: %s", userAgent)
	return nil
}

// getBuildID extracts the build ID from a given user agent string.
// It looks for the presence of "CL-" or "-" in the user agent string
// and splits the string accordingly. The build ID is assumed to be
// the first part after the split, further split by spaces and "+".
// If no build ID is found, an empty string is returned.
//
// Parameters:
//
//	userAgent (string): The user agent string to parse.
//
// Returns:
//
//	string: The extracted build ID or an empty string if not found.
func getBuildID(userAgent string) string {
	var parts []string
	if strings.Contains(userAgent, "CL-") {
		parts = strings.Split(userAgent, "CL-")
	} else if strings.Contains(userAgent, "-") {
		parts = strings.Split(userAgent, "-")
	}

	if len(parts) > 1 {
		subParts := strings.Split(parts[1], " ")
		return strings.Split(subParts[0], "+")[0]
	}

	return ""
}

// getPlatform extracts the platform information from a given user agent string.
// It assumes that the platform information is the last word in the user agent string.
// If the user agent string is empty or does not contain any spaces, it returns an empty string.
//
// Parameters:
//   userAgent - the user agent string from which to extract the platform information.
//
// Returns:
//   A string representing the platform information extracted from the user agent string.

func getPlatform(userAgent string) string {
	lastSpaceIndex := strings.LastIndex(userAgent, " ")
	if lastSpaceIndex != -1 && lastSpaceIndex < len(userAgent)-1 {
		return userAgent[lastSpaceIndex+1:]
	}
	return ""
}

// getBuildString extracts the build string from a user agent string.
// The user agent string is expected to contain a "Release-" substring followed by the build string.
// If the "Release-" substring is not found or the build string is empty, an empty string is returned.
//
// Parameters:
//
//	userAgent - the user agent string to parse
//
// Returns:
//
//	the extracted build string or an empty string if not found
func getBuildString(userAgent string) string {
	parts := strings.Split(userAgent, "Release-")
	if len(parts) <= 1 || len(parts[1]) == 0 {
		return ""
	}

	subParts := strings.Split(parts[1], "-")
	if len(subParts) == 0 {
		return ""
	}

	version := strings.TrimSpace(subParts[0])
	return version
}

func parseNetCL(versionId string) float64 {
	netcl, err := strconv.ParseFloat(versionId, 64)
	if err != nil {
		netcl = 0
	}
	return netcl
}

// handleValidBuild processes the user agent information and operating system
// to generate a SeasonInfo object. It parses the net client version, determines
// the season, and identifies the appropriate lobby.
//
// Parameters:
//   - userAgentInfo: A pointer to a context.UserAgent struct containing user agent details.
//   - os: A string representing the operating system.
//
// Returns:
//   - A pointer to a context.SeasonInfo struct containing the parsed net client version,
//     operating system, and determined lobby.
func handleValidBuild(userAgentInfo *context.UserAgent, os string) *context.SeasonInfo {
	netcl := parseNetCL(userAgentInfo.VersionId)
	season, buildVersion := parseVersion(userAgentInfo.VersionStr)

	lobby := getLobby(netcl, season)

	result := &context.SeasonInfo{
		Season: season,
		Build:  buildVersion,
		CL:     strconv.FormatFloat(netcl, 'f', -1, 64),
		OS:     os,
		Lobby:  lobby,
	}
	return result
}

func parseVersion(versionStr string) (int, string) {
	if versionStr == "" {
		return 0, ""
	}

	versionStr = strings.TrimPrefix(versionStr, "Release-")

	versionParts := strings.Split(versionStr, ".")
	season := 0
	buildVersion := versionStr

	if len(versionParts) > 0 {
		majorVersion, err := strconv.Atoi(versionParts[0])
		if err == nil {
			season = majorVersion
			buildVersion = versionStr
		}
	}

	return season, buildVersion
}

func getLobby(netcl float64, season int) string {
	switch {
	case netcl == 0:
		return "LobbySeason0"
	case netcl < 3724489:
		return "Season0"
	case netcl <= 3790078:
		return "LobbySeason1"
	case season == 6:
		return "Lobby6"
	case season == 10:
		return "Lobby10"
	default:
		return "Lobby" + strconv.Itoa(season)
	}
}
