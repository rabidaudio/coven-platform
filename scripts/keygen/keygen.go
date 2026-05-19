package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/atlantacoven/coven-platform/tlvwt"
)

const DOMAIN = "thecoven.space"

// The AID is an application ID as defined by ISO 7816-4. They can be up to 16 bytes.
// Values starting with 0xF* are proprietary. Here we use our domain name
// surrounded by 0xFF bytes.
var AID []byte

func init() {
	AID = append(AID, 0xFF)
	AID = append(AID, DOMAIN...)
	AID = append(AID, 0xFF)
}

// INFO is the value of the optional info field used in the HPKE algorithm. Here we use
// our domain.
var INFO = DOMAIN

// paths
const serverSigningPath = "member-site/signing.pem"
const firmwareIncludePath = "firmware/include/_keys.h"
const androidResPath = "android/app/src/main/res/values/keys.xml"

func main() {
	log.Println("Generating new keys...")
	ServerSignPub, ServerKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	DoorSignPub, DoorKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	log.Printf("Saving server key to: %v\n", serverSigningPath)
	ssk := must(x509.MarshalPKCS8PrivateKey(ServerKey))
	f0 := must(os.Create(serverSigningPath))
	defer f0.Close()
	err = pem.Encode(f0, &pem.Block{Type: "ED25519 PRIVATE KEY", Bytes: ssk})
	if err != nil {
		panic(err)
	}

	log.Printf("Saving includes file for door lock code: %v\n", firmwareIncludePath)
	f1 := must(os.Create(firmwareIncludePath))
	defer f1.Close()
	f1.WriteString("#include <Arduino.h>\n\n")
	// TODO: put these in PROGMEM instead
	fmt.Fprintf(f1, "#define EXCHANGE_INFO_LEN %v\nbyte EXCHANGE_INFO[] = \"%v\";\n\n", len(INFO), INFO)
	fmt.Fprintf(f1, "#define AID_LEN %v\n", len(AID))
	writeCArray(f1, "AID", AID)
	writeCArray(f1, "DOOR_SIGN_PRIV_KEY", DoorKey)
	writeCArray(f1, "SERVER_SIGNING_PUB_KEY", ServerSignPub)

	log.Printf("Saving public keys for Android app: %v\n", androidResPath)
	f2 := must(os.Create(androidResPath))
	defer f2.Close()
	f2.WriteString("<resources>\n")
	fmt.Fprintf(f2, "\t<string name=\"aid\">%v</string>\n", hex.EncodeToString(AID))
	fmt.Fprintf(f2, "\t<string name=\"info\">%v</string>\n", INFO)
	spk := must(x509.MarshalPKIXPublicKey(ServerSignPub))
	fmt.Fprintf(f2, "\t<string name=\"server_signing_pubkey\">%v</string>\n", hex.EncodeToString(spk))
	dpk := must(x509.MarshalPKIXPublicKey(DoorSignPub))
	fmt.Fprintf(f2, "\t<string name=\"door_signing_pubkey\">%v</string>\n", hex.EncodeToString(dpk))
	f2.WriteString("</resources>\n")

	// Test key (expired)
	userId := 1234
	token := generateExpiredToken(userId, ServerKey)
	fmt.Println("Test Token:")
	fmt.Printf("%x\n", token)
}

func generateExpiredToken(uid int, serverkey ed25519.PrivateKey) []byte {
	token := tlvwt.Token{}
	token.SetType("TWT")
	token.SetIssuer(DOMAIN)
	expires := time.Now().Add(-24 * time.Hour)
	token.SetExpiration(expires)
	tlvwt.TLVMessage(token).SetUInt64("uid", uint64(uid))
	return must(token.SignAndEncodeString(tlvwt.Ed25519Algorithm{Key: serverkey}))
}

func writeCArray(w io.Writer, name string, data []byte) (err error) {
	_, err = fmt.Fprintf(w, "/* %v */\n", hex.EncodeToString(data))
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(w, "const uint8_t %v[%v] = {\n\t", name, len(data))
	if err != nil {
		return
	}
	for i, b := range data {
		_, err = fmt.Fprintf(w, "0x%x, ", b)
		if err != nil {
			return
		}
		if (i+1)%16 == 0 && (i+1) != len(data) {
			_, err = w.Write([]byte("\n\t"))
			if err != nil {
				return
			}
		}
	}
	_, err = w.Write([]byte("\n};\n\n"))
	return
}

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
