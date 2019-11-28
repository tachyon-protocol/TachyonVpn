package tyTls

import (
	"os"
	"sync"
)

var gAllowTlsVersion13Once sync.Once
func AllowTlsVersion13(){
	gAllowTlsVersion13Once.Do(func(){
		os.Setenv("GODEBUG","tls13=1")
	})
}