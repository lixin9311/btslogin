package api

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gin-gonic/gin"
	bitshares "github.com/scorum/bitshares-go"
)

var (
	client *bitshares.Client
	data   = map[string][32]byte{}
)

type challenge struct {
	Apikey   string
	Username string
	Signed   string
}

func init() {
	var err error
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	client, err = bitshares.NewClient("wss://b.mrx.im/ws", "private", "node")
	if err != nil {
		log.Fatalln(err)
	}
}

func ChallengeGet(c *gin.Context) {
	apikey, _ := c.GetQuery("apikey")
	username, _ := c.GetQuery("username")
	// apikey processing
	challangeText := fmt.Sprintf("%s@%s+%d", username, apikey, time.Now().UnixNano())
	hashed := sha256.Sum256([]byte(challangeText))
	data[apikey+username] = hashed
	resp := base64.StdEncoding.EncodeToString(hashed[:])
	c.String(200, "%s", resp)
}

func ChallengePost(c *gin.Context) {
	req := challenge{}
	c.BindJSON(&req)
	hashed, ok := data[req.Apikey+req.Username]
	if !ok {
		c.AbortWithStatusJSON(404, gin.H{"code": 404, "token": ""})
		return
	}

	accountDetail, err := client.Database.GetAccountByName(req.Username)
	if err != nil {
		log.Println("Failed to get account detail: ", err)
		c.AbortWithStatusJSON(500, gin.H{"code": 500, "token": ""})
		return
	}
	if accountDetail == nil {
		c.AbortWithStatusJSON(404, gin.H{"code": 404, "token": ""})
		return
	}

	for _, auth := range accountDetail.Active.KeyAuths {
		pubk, err := parsePubkey(auth.Key)
		if err != nil {
			log.Println("Failed to parse public key: ", err)
			c.AbortWithStatusJSON(500, gin.H{"code": 500, "token": ""})
			return
		}
		if ok, err := verifySig(hashed, req.Signed, pubk); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"code": 400, "token": ""})
			return
		} else {
			if ok {
				token := generateToken(req.Username, req.Apikey, hashed)
				c.AbortWithStatusJSON(200, gin.H{"code": 200, "token": token})
				return
			}
		}
	}

	c.AbortWithStatusJSON(400, gin.H{"code": 400, "token": ""})
	return
}

func parsePubkey(pubkeyStr string) (*btcec.PublicKey, error) {
	pbkeyStr := pubkeyStr
	pbkeyBytes := base58.Decode(pbkeyStr[3:])
	pbkeyBytes = pbkeyBytes[:33]
	return btcec.ParsePubKey(pbkeyBytes, btcec.S256())
}

func verifySig(hashed [32]byte, signed string, pubkey *btcec.PublicKey) (bool, error) {
	signedBytes, err := base64.StdEncoding.DecodeString(signed)
	if err != nil {
		return false, err
	}
	sig, err := btcec.ParseDERSignature(signedBytes, btcec.S256())
	if err != nil {
		return false, err
	}
	return sig.Verify(hashed[:], pubkey), nil
}

func generateToken(username, apikey string, hashed [32]byte) string {
	datatohash := append([]byte(username+apikey), hashed[:]...)
	tokenHashed := sha256.Sum256(datatohash)
	return base64.StdEncoding.EncodeToString(tokenHashed[:])
}
