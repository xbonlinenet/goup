package ab

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var (
	ErrRangeNotValid = errors.New("range format not vaild")
)

type Range struct {
	Start int
	End   int
}

func (r Range) In(userID int64) bool {
	mod := userID % 100
	g := int(mod)
	if g >= r.Start && g < r.End {
		return true
	}

	return false
}

// UserIDModGroup 按用户 ID 取模分组
type UserIDModGroup struct {
	Name   string
	Ranges []Range
	Config map[string]interface{}
}

// In 判断用户是否在测试分组当中
func (group *UserIDModGroup) In(userID int64) bool {

	for _, r := range group.Ranges {
		if r.In(userID) {
			return true
		}
	}
	return false
}

func (group *UserIDModGroup) GetInt64(key string) int64 {
	value := group.Config[key]
	i, _ := cast.ToInt64E(value)
	return i
}

func (group *UserIDModGroup) GetBool(key string) bool {
	value := group.Config[key]
	return cast.ToBool(value)
}

func (group *UserIDModGroup) GetString(key string) string {
	value := group.Config[key]
	return cast.ToString(value)
}

func GetGroupFromConf(path string) (*UserIDModGroup, error) {

	str := viper.GetString(fmt.Sprintf("%s.range", path))
	ranges, err := ParseRange(str)
	if err != nil {
		return nil, err
	}

	return &UserIDModGroup{
		Name:   path,
		Ranges: ranges,
		Config: viper.GetStringMap(fmt.Sprintf("%s.config", path)),
	}, nil
}

// ParseRange 从字符串中解析分组范围, 10-20,50-70
func ParseRange(str string) ([]Range, error) {
	rangesStr := strings.Split(str, ",")

	ranges := make([]Range, 0, len(rangesStr))
	for _, r := range rangesStr {
		pos := strings.Split(strings.TrimSpace(r), "-")
		if len(pos) < 2 {
			return []Range{}, ErrRangeNotValid
		}

		start, err := strconv.Atoi(pos[0])
		if err != nil {
			return []Range{}, ErrRangeNotValid
		}

		end, err := strconv.Atoi(pos[1])
		if err != nil {
			return []Range{}, ErrRangeNotValid
		}

		ranges = append(ranges, Range{
			Start: start,
			End:   end,
		})
	}
	return ranges, nil
}
