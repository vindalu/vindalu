package auth

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func LoadLocalAuthGroups(groupsfile string) (localAuthGroups LocalAuthGroups, err error) {
	if !filepath.IsAbs(groupsfile) {
		groupsfile, _ = filepath.Abs(groupsfile)
	}

	var b []byte

	if b, err = ioutil.ReadFile(groupsfile); err != nil {
		return
	}
	err = json.Unmarshal(b, &localAuthGroups)
	return
}

func (l *LocalAuthGroups) UserHasGroupMembership(user, group string) bool {
	for k, v := range *l {
		if k == group {
			for _, u := range v {
				if u == user {
					return true
				}
			}
			return false
		}
	}
	return false
}
