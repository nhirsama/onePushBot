package auth

import (
	"fmt"
	"testing"
)

func TestAuth_VerifyPGPSignature(t *testing.T) {
	auth := NewAuth()
	//正确的签名
	s := "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\n123123123\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQQRwGo4vvZxEEkeupu98uJkGyYLKAUCaPzCfgAKCRC98uJkGyYL\nKDchAP9fluP1ITQFuEGJVik+5Y9GVGEFB9lkjaafnWvZH/hWPwD6A1IOBkdr2Li/\ne8Lcyv1/+pQCnkxQmCSI3M2BT79cjQM=\n=O158\n-----END PGP SIGNATURE-----\nQkKhpQ7uYTAhUAagM=\n=Fii4\n- -----END PGP SIGNATURE-----\n\n/92VBvmKUG+RCy5eE9oxHw1H62VGB9cqd853Tr9xwEA+wWokVtQ2aZk\nQpn1acUusrD3dtQkKhpQ7uYTAhUAagM=\n=Fii4\n-----END PGP SIGNATURE-----\n"
	st := auth.VerifyPGPSignature([]byte(s))
	if st.Err != nil {
		fmt.Println(st)
		fmt.Println(string(st.Message))
		t.Error(st.Err)
	}
	//签名人错误
	s = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\neqweqwecszxfas\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQTAqmuCZXka96AtN9KTBVxAaFiKZwUCaPzGgwAKCRCTBVxAaFiK\nZ1aPAP4uJVyNeX7LUrs1YyBeqohhMCxEaZ2Kshb2iwivwBYt+wEAn0RoNDLXYi/9\nXsaIUMjLT1wFKjL6hpl8TqNl9b3p8gI=\n=l4oj\n-----END PGP SIGNATURE-----\n"
	st = auth.VerifyPGPSignature([]byte(s))
	if st.Valid == true {
		fmt.Println(st)
		t.Error(st.Err)
	}

	//被篡改
	s = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA512\n\n12313123123\n-----BEGIN PGP SIGNATURE-----\n\niHUEARYKAB0WIQQRwGo4vvZxEEkeupu98uJkGyYLKAUCaPzCfgAKCRC98uJkGyYL\nKDchAP9fluP1ITQFuEGJVik+5Y9GVGEFB9lkjaafnWvZH/hWPwD6A1IOBkdr2Li/\ne8Lcyv1/+pQCnkxQmCSI3M2BT79cjQM=\n=O158\n-----END PGP SIGNATURE-----\nQkKhpQ7uYTAhUAagM=\n=Fii4\n- -----END PGP SIGNATURE-----\n\n/92VBvmKUG+RCy5eE9oxHw1H62VGB9cqd853Tr9xwEA+wWokVtQ2aZk\nQpn1acUusrD3dtQkKhpQ7uYTAhUAagM=\n=Fii4\n-----END PGP SIGNATURE-----\n"
	st = auth.VerifyPGPSignature([]byte(s))
	if st.Valid == true {
		fmt.Println(st)
		t.Error(st.Err)
	}
}
