package tyTls

import (
	"os"
	"sync"
)

var gEnableTlsVersion13Once sync.Once
func EnableTlsVersion13(){
	gEnableTlsVersion13Once.Do(func(){
		os.Setenv("GODEBUG","tls13=1")
	})
}