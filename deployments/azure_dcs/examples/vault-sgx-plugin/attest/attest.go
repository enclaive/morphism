package attest

import (
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
)

const (
	Debug = false
)

var (
	logger *log.Logger
)

func init() {
	if Debug {
		file, err := os.Create("plugin.log.txt")
		if err != nil {
			panic(err)
		}
		logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	}
}

type Request struct {
	Name  string
	Quote []byte
}

type Response struct {
	Quote       []byte
	Certificate *x509.Certificate
}

type Secrets struct {
	Environment map[string]string `json:"environment,omitempty"`
	Files       map[string][]byte `json:"files,omitempty"`
	Argv        []string          `json:"argv,omitempty"`
}

// Verify
// TODO add nonce to quote generation and verification
func Verify(hash [64]byte, quote []byte, reference string) error {
	if Debug {
		logger.Println("Measurement:", reference)
		logger.Println("Quote:", base64.StdEncoding.EncodeToString(quote))
		logger.Println("Hash:", hex.EncodeToString(hash[:]))
	}

	attested, err := verifyQuote(quote)
	if err != nil {
		if Debug {
			logger.Println("Error:", err)
		}
		return err
	}

	if subtle.ConstantTimeCompare(attested.ReportData, hash[:]) == 0 {
		return errors.New("report data did not match expected hash")
	}

	rawReference, err := hex.DecodeString(reference)
	if err != nil {
		if Debug {
			logger.Println("Error:", err)
		}
		return err
	}

	if subtle.ConstantTimeCompare(attested.Identity.MrEnclave[:], rawReference) == 0 {
		return errors.New("measurement did not match expected mrenclave")
	}

	if Debug {
		rawAttested, _ := json.Marshal(attested)
		logger.Println("Attested:", string(rawAttested))
		logger.Println("MRENCLAVE:", hex.EncodeToString(attested.Identity.MrEnclave[:]))
		logger.Println("MRSIGNER:", hex.EncodeToString(attested.Identity.MrSigner[:]))
		logger.Println("DATA:", hex.EncodeToString(hash[:]))
	}

	return nil
}
