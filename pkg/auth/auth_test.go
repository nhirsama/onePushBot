package auth

import (
	"testing"
)

func TestAuth_Authenticate(t *testing.T) {
	auth := NewAuth()
	//正确的签名
	s := "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\n123123123\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQQRwGo4vvZxEEkeupu98uJkGyYLKAUCaPzCfgAKCRC98uJkGyYL\nKDchAP9fluP1ITQFuEGJVik+5Y9GVGEFB9lkjaafnWvZH/hWPwD6A1IOBkdr2Li/\ne8Lcyv1/+pQCnkxQmCSI3M2BT79cjQM=\n=O158\n-----END PGP SIGNATURE-----\nQkKhpQ7uYTAhUAagM=\n=Fii4\n- -----END PGP SIGNATURE-----\n\n/92VBvmKUG+RCy5eE9oxHw1H62VGB9cqd853Tr9xwEA+wWokVtQ2aZk\nQpn1acUusrD3dtQkKhpQ7uYTAhUAagM=\n=Fii4\n-----END PGP SIGNATURE-----\n"
	if auth.Authenticate([]byte(s)) {
		t.Error("过期签名验证为真")
	}
	//签名人错误
	s = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\neqweqwecszxfas\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQTAqmuCZXka96AtN9KTBVxAaFiKZwUCaPzGgwAKCRCTBVxAaFiK\nZ1aPAP4uJVyNeX7LUrs1YyBeqohhMCxEaZ2Kshb2iwivwBYt+wEAn0RoNDLXYi/9\nXsaIUMjLT1wFKjL6hpl8TqNl9b3p8gI=\n=l4oj\n-----END PGP SIGNATURE-----\n"
	if auth.Authenticate([]byte(s)) {
		t.Error("错误签名验证为真")
	}

	//被篡改
	s = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\n12313123123\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQQRwGo4vvZxEEkeupu98uJkGyYLKAUCaPzCfgAKCRC98uJkGyYL\nKDchAP9fluP1ITQFuEGJVik+5Y9GVGEFB9lkjaafnWvZH/hWPwD6A1IOBkdr2Li/\ne8Lcyv1/+pQCnkxQmCSI3M2BT79cjQM=\n=O158\n-----END PGP SIGNATURE-----\nQkKhpQ7uYTAhUAagM=\n=Fii4\n- -----END PGP SIGNATURE-----\n\n/92VBvmKUG+RCy5eE9oxHw1H62VGB9cqd853Tr9xwEA+wWokVtQ2aZk\nQpn1acUusrD3dtQkKhpQ7uYTAhUAagM=\n=Fii4\n-----END PGP SIGNATURE-----\n"
	if auth.Authenticate([]byte(s)) {
		t.Error("被篡改签名验证为真")
	}
	//若失败请签名一个五分钟内的信息用以验证
	s = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\n213123123123\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQQRwGo4vvZxEEkeupu98uJkGyYLKAUCaPzgdgAKCRC98uJkGyYL\nKEpwAP9R/c458412VV7Tqky0y4KEQXoulVBHnc1rApa29LSPiQEA4pDmaYXAz1ju\nWgInr3db6KuE+9d7PFB0oeL+7N31wAc=\n=0NdB\n-----END PGP SIGNATURE-----\n"
	if !auth.Authenticate([]byte(s)) {
		t.Error("正确签名验证为否")
	}
	if auth.Authenticate([]byte(s)) {
		t.Error("重放签名验证为真")
	}
}
