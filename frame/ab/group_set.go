package ab

import (
	"fmt"

	"github.com/spf13/viper"
)

type GroupSet struct {
	groups []*UserIDModGroup
}

func (s GroupSet) FindGroup(userID int64) *UserIDModGroup {
	for _, group := range s.groups {
		if group.In(userID) {
			return group
		}
	}
	return nil
}

func GetGroupSetFromConf(path string) (GroupSet, error) {
	maps := viper.GetStringMap(path)
	groups := make([]*UserIDModGroup, 0, len(maps))
	for key, _ := range maps {
		group, err := GetGroupFromConf(fmt.Sprintf("%s.%s", path, key))
		if err != nil {
			return GroupSet{}, err
		}
		groups = append(groups, group)
	}

	return GroupSet{
		groups: groups,
	}, nil
}
