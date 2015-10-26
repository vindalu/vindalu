package versioning

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/vindalu/vindalu/store"
)

type VersionDiff struct {
	Version        int64       `json:"version"`
	UpdatedBy      interface{} `json:"updated_by"`
	Timestamp      interface{} `json:"timestamp"`
	AgainstVersion int64       `json:"against_version"`
	Diff           string      `json:"diff"`
}

func GenerateDiff(prevName, prevStr, currName, currStr string) (string, error) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(prevStr),
		B:        difflib.SplitLines(currStr),
		FromFile: prevName,
		ToFile:   currName,
		Context:  0,
	}
	return difflib.GetUnifiedDiffString(diff)
}

/* Convert a given number to an int64 */
func parseVersion(ver interface{}) (verInt int64, err error) {
	switch ver.(type) {
	case float64:
		verF, _ := ver.(float64)
		verInt = int64(verF)
		break
	case int64:
		verInt, _ = ver.(int64)
		break
	case int:
		verI, _ := ver.(int)
		verInt = int64(verI)
		break
	case string:
		verStr, _ := ver.(string)
		verInt, err = strconv.ParseInt(verStr, 10, 64)
		break
	default:
		err = fmt.Errorf("Could not parse version: %v", ver)
		break
	}

	return
}

func GenerateVersionDiffs(versions ...store.BaseAsset) (list []VersionDiff, err error) {
	list = make([]VersionDiff, len(versions)-1)

	for i, version := range versions {
		if i+1 >= len(versions) {
			break
		}

		var (
			bi, bi1 []byte
			text    string
		)

		// Store and remove version before diff'ing
		var verInt int64
		if verInt, err = parseVersion(version.Data["version"]); err != nil {
			return
		}
		delete(versions[i].Data, "version")

		var verInt1 int64
		if verInt1, err = parseVersion(versions[i+1].Data["version"]); err != nil {
			return
		}
		delete(versions[i+1].Data, "version")

		if bi, err = json.MarshalIndent(version.Data, "", " "); err != nil {
			return
		}
		if bi1, err = json.MarshalIndent(versions[i+1].Data, "", " "); err != nil {
			return
		}

		if text, err = GenerateDiff(
			fmt.Sprintf("v%d", verInt1), fmt.Sprintf("%s", bi1),
			fmt.Sprintf("v%d", verInt), fmt.Sprintf("%s", bi)); err != nil {
			return
		}
		//list[fmt.Sprintf("v%d", i+1)] = text
		list[i] = VersionDiff{
			Version:        verInt,
			UpdatedBy:      version.Data["updated_by"],
			Timestamp:      version.Timestamp,
			AgainstVersion: verInt1,
			Diff:           text,
		}
		// put next version back after the diff is calculated
		versions[i+1].Data["version"] = verInt1
	}
	return
}
