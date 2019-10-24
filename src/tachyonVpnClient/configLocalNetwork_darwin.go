//+build darwin

package tachyonVpnClient

import "github.com/tachyon-protocol/udw/udwNet"

func configLocalNetwork () {
	udwNet.MustSetDnsServerAddr("8.8.8.8")
}

func recoverLocalNetwork () {
	udwNet.MustSetDnsServerToDefault()
}
