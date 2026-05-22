package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
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
const androidResPath = "app/android/app/src/main/res/values/keys.xml"
const iosSwiftPath = "app/ios/Runner/Secrets.swift"

func main() {
	  err := godotenv.Load()
    if err != nil {
        panic(err)
    }

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
	f1.WriteString("#ifndef _KEYS_H_\n#define _KEYS_H_\n")
	f1.WriteString("#include <Arduino.h>\n\n")
	// TODO: put these in PROGMEM instead
	fmt.Fprintf(f1, "#define EXCHANGE_INFO_LEN %v\nbyte EXCHANGE_INFO[] = \"%v\";\n\n", len(INFO), INFO)
	fmt.Fprintf(f1, "#define AID_LEN %v\n", len(AID))
	writeCArray(f1, "AID", AID)
	writeCArray(f1, "DOOR_SIGN_PRIV_KEY", DoorKey)
	writeCArray(f1, "SERVER_SIGNING_PUB_KEY", ServerSignPub)
	if os.Getenv("WIFI_SSID") == "" || os.Getenv("WIFI_PASS") == "" {
		panic(fmt.Errorf("WIFI_SSID and WIFI_PASS env vars must be set"))
	}
	f1.WriteString("\n\n")
	fmt.Fprintf(f1, "char WIFI_SSID[] = \"%v\";\n",os.Getenv("WIFI_SSID"))
	fmt.Fprintf(f1, "char WIFI_PASS[] = \"%v\";\n",os.Getenv("WIFI_PASS"))
	f1.WriteString("#endif // _KEYS_H_\n")

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

	log.Printf("Saving public keys for iOS app: %v\n", iosSwiftPath)
	f3 := must(os.Create(iosSwiftPath))
	defer f3.Close()
	f3.WriteString("import CryptoKit\n\n")
	fmt.Fprintf(f3, "// %v\n", hex.EncodeToString(AID))
	fmt.Fprintf(f3, "let AID = Data(base64Encoded: String(\"%v\").data(using: .ascii)!)!\n\n", base64.StdEncoding.EncodeToString(AID))
	fmt.Fprintf(f3, "let INFO = \"%v\".data(using: .ascii)!\n\n", INFO)
	fmt.Fprintf(f3, "// %v\n", hex.EncodeToString(ServerSignPub))
	fmt.Fprintf(f3, "let ServerSignPub = try! Curve25519.Signing.PublicKey(rawRepresentation: Data(base64Encoded: String(\"%v\").data(using: .ascii)!)!)\n\n", base64.StdEncoding.EncodeToString(ServerSignPub))
	fmt.Fprintf(f3, "// %v\n", hex.EncodeToString(DoorSignPub))
	fmt.Fprintf(f3, "let DoorSignPub = try! Curve25519.Signing.PublicKey(rawRepresentation: Data(base64Encoded: String(\"%v\").data(using: .ascii)!)!)\n", base64.StdEncoding.EncodeToString(DoorSignPub))
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
