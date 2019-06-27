package aliacm

import (
	"fmt"
)

// URL for api construction.
type URL string

const (
	acmIP            URL = "/diamond-server/diamond"
	acmConfig        URL = "/diamond-server/config.co"
	acmDeleteConfig  URL = "/diamond-server/datum.do?method=deleteAllDatums"
	acmPublishConfig URL = "/diamond-server/basestone.do?method=syncUpdateAll"
	acmAllConfig     URL = "/diamond-server/basestone.do?method=getAllConfigByTenant"
	acmLongPull      URL = "/diamond-server/config.co"
)

func (u URL) String(addr string) string {
	return fmt.Sprintf("http://%s:8080%s", addr, string(u))
}
