package random

import (
	"github.com/rs/xid"
)

func GenerateXid() string {
	id := xid.New()
	return id.String()
	//fmt.Printf("github.com/rs/xid:           %s\n", id.String())
}
