package attest

import (
	"crypto/sha512"
	"os"
)

const (
	GramineUserReportData = "/dev/attestation/user_report_data"
	GramineQuote          = "/dev/attestation/quote"
)

func NewGramineIssuer() *GramineIssuer {
	return &GramineIssuer{}
}

type GramineIssuer struct{}

func (i *GramineIssuer) Issue(data []byte) ([]byte, error) {
	hash := sha512.Sum512(data)

	if err := os.WriteFile(GramineUserReportData, hash[:], 0600); err != nil {
		return nil, err
	}

	rawQuote, err := os.ReadFile(GramineQuote)
	if err != nil {
		return nil, err
	}

	return rawQuote, nil
}
