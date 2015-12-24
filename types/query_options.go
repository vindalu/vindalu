package types

import (
	"fmt"
	"strconv"
	"strings"
)

type QueryOptions struct {
	From      int64               // starting point
	Size      int64               // dataset size (from starting point)
	Sort      []map[string]string // <property>:asc, <property>:desc
	Aggregate string              // property
}

func NewQueryOptions(req map[string][]string) (qo QueryOptions, err error) {
	for k, v := range req {

		switch k {
		case "from":
			qo.From, err = strconv.ParseInt(v[0], 10, 64)
		case "size":
			qo.Size, err = strconv.ParseInt(v[0], 10, 64)
		case "sort":
			qo.Sort, err = parseSortOptions(v)
		case "aggregate":
			qo.Aggregate = strings.TrimSpace(v[0])
		}

		if err != nil {
			break
		}
	}

	return
}

// Data as map
func (qo *QueryOptions) Map() map[string]interface{} {
	m := make(map[string]interface{})
	m["from"] = qo.From
	m["size"] = qo.Size
	if len(qo.Sort) > 0 {
		m["sort"] = qo.Sort
	}
	if len(qo.Aggregate) > 0 {
		m["aggregate"] = qo.Aggregate
	}
	return m
}

func parseSortOptions(sortOpts []string) (sopts []map[string]string, err error) {

	sopts = make([]map[string]string, len(sortOpts))

	for i, sval := range sortOpts {
		keyOrder := strings.Split(sval, ":")
		switch len(keyOrder) {
		case 1:
			//Case 1: no sorting order specified, do ascending by default
			sopts[i] = map[string]string{keyOrder[0]: "asc"}
			break
		case 2:
			//Case 2: sorting order specified, do what it says
			if keyOrder[1] != "asc" && keyOrder[1] != "desc" {
				err = fmt.Errorf("Sort must be in `key:[asc desc]` format")
				return
			}
			sopts[i] = map[string]string{keyOrder[0]: keyOrder[1]}
			break
		default:
			err = fmt.Errorf("Sort must be in `key:[asc desc]` format")
			return
		}
	}
	return
}
