package debug

import (
	// "bytes"
	// "crypto/rand"
	// "encoding/json"
	// "fmt"
	// "io"
	// "net/http"
	"runtime"
	"time"
)

type context struct {
	// uuid    string
	os      string
	started time.Time
	debug   bool
	email   string
}

func NewContext(debug bool) (*context, error) {
	// s, err := newUUID()
	// if err != nil {
	// 	return nil, err
	// }
	return &context{
		started: time.Now(),
		os:      runtime.GOOS,
		// uuid:    s,
		debug:   debug,
	}, nil
}

func (d *context) Action(text string) {
	if d.debug {
	}
}

func (d *context) Error(text string, err error) {
	if d.debug {
	}
}

// func (d *context) UUID() string {
// 	return d.uuid
// }
// func (d *context) SetEmail(email string) {
// 	d.email = email
// }
// func (d *context) IsDiagnostics() bool {
// 	return d.debug
// }

// func newUUID() (string, error) {
// 	uuid := make([]byte, 16)
// 	n, err := io.ReadFull(rand.Reader, uuid)
// 	if n != len(uuid) || err != nil {
// 		return "", err
// 	}
// 	// variant bits; see section 4.1.1
// 	uuid[8] = uuid[8]&^0xc0 | 0x80
// 	// version 4 (pseudo-random); see section 4.1.3
// 	uuid[6] = uuid[6]&^0xf0 | 0x40
// 	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil

// }
