package configutil

import (
	"fmt"
	"sort"
	"strings"
)

func DumpConfigsInDotEnvFormat(params ConfigsInfoIn) {
	configs := make([]ConfigInfo, len(params.ConfigsInfo))
	copy(configs, params.ConfigsInfo)

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].ConfigType.String() > configs[j].ConfigType.String()
	})

	for _, configInfo := range configs {
		fmt.Println("#", configInfo.ConfigType.String())
		for _, fieldParams := range configInfo.Fields {
			var comments []string
			if fieldParams.Required {
				comments = append(comments, "required")
			}
			if fieldParams.NotEmpty {
				comments = append(comments, "not empty")
			}
			if fieldParams.Expand {
				comments = append(comments, "expands")
			}
			if fieldParams.LoadFile {
				comments = append(comments, "loaded from file")
			}

			commentsStr := ""
			if len(comments) > 0 {
				commentsStr = " # " + strings.Join(comments, ", ")
			}

			fmt.Printf("%s=\"%s\" %s\n", fieldParams.Key, fieldParams.DefaultValue, commentsStr)
		}
	}
}
