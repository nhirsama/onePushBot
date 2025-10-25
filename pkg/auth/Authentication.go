package auth

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type Auth struct {
	pubs    map[string]struct{}
	queue   []*VerifyResult
	mu      sync.RWMutex
	keyRing openpgp.KeyRing
}

func NewAuth() *Auth {
	pubkey := "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmDMEaOCwHRYJKwYBBAHaRw8BAQdAMnQTvzoYCyJ4fXNO2hibYkN5oP5pcYbteqvR\n9Gg4HVq0G2xpbmcgPG5oaXJzYW1hQG91dGxvb2suY29tPoiQBBMWCgA4FiEESphO\nMjDKzhc1FsczQaagloo4zswFAmjgsB0CGwMFCwkIBwIGFQoJCAsCBBYCAwECHgEC\nF4AACgkQQaagloo4zsw4nAEA1NhY41watg+tzaBwVsjH2qimEevjR8aquBCwUZE3\nBdoA/1Po5iRnwG/v/j4CKufqBgICGLd86KvU6VAwcfnctF8MiHUEEBYKAB0WIQRL\nGZ1g46ZN5rk7simy5VpjL2GKpAUCaOxXeAAKCRCy5VpjL2GKpH54AQCem+D2xp4D\nDxav1KKYAzRVzbuZ5vFHSxmBjRqZzZ6sRwD+MTTKrxBZHTk9SUbhuc0MLbRZ3UXQ\nltXY12Z3G/Pg1QK4OARo4LAdEgorBgEEAZdVAQUBAQdAHp2+g57y8E4BfUFuBhW4\nvsY/qmzKy6v0qe1zw9AmJnADAQgHiHgEGBYKACAWIQRKmE4yMMrOFzUWxzNBpqCW\nijjOzAUCaOCwHQIbDAAKCRBBpqCWijjOzOFmAQC5bcgsN2yZX3eO28dlHNucQMX7\nWqSI4Vi1tHHx6pS1ZwEA4/R27z9N/zKOyQ2yOYI3j7v0BpUUd/vYFZWHx9ukGQ64\nMwRo6QrTFgkrBgEEAdpHDwEBB0BXwJaiEx7Gm6N7vpjqIrAo8fKfrxp4OEfAz0Pi\n7sHfWoj1BBgWCgAmFiEESphOMjDKzhc1FsczQaagloo4zswFAmjpCtMCGwIFCQlm\nAYAAgQkQQaagloo4zsx2IAQZFgoAHRYhBBHAaji+9nEQSR66m73y4mQbJgsoBQJo\n6QrTAAoJEL3y4mQbJgso3/IBAInDIr9IGzNlrYxy6nZW3TWP1deNg7RfSnvdcKTW\nZWSfAP4j9Q5EBpAsI275Ppl4QzCAXZWTFSCaMaA9D8iyQqBOCPeuAP9fWhaGY4nQ\nBevSFCLQ3Lpok5WdW612tPsae7aavpNaiwEAwEtNPfemUqMY47LoZBlZ7LZF6T6S\nDc5C+ga9RMcR8w+5AY0EaOmrLgEMAKL8JCtFFeyMhyzq2NPdY/MGRalnuD89FnGP\n3e/K3ZJSubFCyWk3Eq5o69Zbvzq5WDdn98g6JuPXk+Mkui0mPReRy5ITs6m5fBKu\n1I6gcnp3LNLAYClAxZfnUjyXJerUVCmoODdsA1P+nOwaWsbhSvzyZqydqshRdKuL\nxk60b6WCkWNjpvmsWGNhuuFBbODsh5VeZylMRcTq8jy4vJyFHIw7FZdMCLmK+d1q\nSdK4mHB0DvaH7aBXsQjCf0gp0j6CuVgELbkLXk1NMHsF8pcxXV9j1trTCL4S4jGy\neIY940V+CPVQKP5G/HLvNoeOdtmL4sJ1569Vx2YH3Wqv8YdVFStoCOcQQvLDGMkL\nkAfvb5FDmzBAV5BCoDOCl+nZ1L+lifJyDlN5P+OUmAFQ31ggYb7wEvBkJkTtmum7\n1ma1AV2d1s1W1iBClI1vt/6eR/PUuYONZiUIdNA+ksj/BOwfMK0Sfffh550eyZQK\n8y8T4TRHpjPTvpmyFtMrBI862fAYawARAQABiH4EGBYKACYWIQRKmE4yMMrOFzUW\nxzNBpqCWijjOzAUCaOmrLgIbIAUJCWYBgAAKCRBBpqCWijjOzILFAP91zt/A59//\n/XNjtFtZVR4rLwxbd96Hs1YFeEF3/1VV6wD+Kl0715A4eJqRtxWiLekhc9IXIVim\nQAqSgHJuFq1MqAQ=\n=jbYy\n-----END PGP PUBLIC KEY BLOCK-----\n"
	keyRing, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pubkey))
	if err != nil {
		log.Fatalf("failed to read keyring from hardcoded data: %s", err)
	}
	Auth := &Auth{
		pubs:    make(map[string]struct{}),
		keyRing: keyRing,
	}
	return Auth
}

func (a *Auth) Authenticate(message []byte) bool {

	re := a.VerifyPGPSignature(message)
	if !re.Valid {
		log.Println("invalid signature", re.Err)
		return false
	} else {
		if time.Now().Unix()-re.SignedAt > 300 {
			log.Println("signature expired")
			return false
		}
		hashkey := re.Hash + strconv.FormatInt(re.SignedAt, 10)

		a.mu.Lock()
		for len(a.queue) > 0 && time.Now().Unix()-a.queue[0].SignedAt > 300 {
			newHashKey := a.queue[0].Hash + strconv.FormatInt(a.queue[0].SignedAt, 10)
			delete(a.pubs, newHashKey)
			a.queue = a.queue[1:]
		}
		if _, ok := a.pubs[hashkey]; ok {
			return false
		}
		a.queue = append(a.queue, &re)
		a.pubs[hashkey] = struct{}{}
		defer a.mu.Unlock()
		return true
	}
}
