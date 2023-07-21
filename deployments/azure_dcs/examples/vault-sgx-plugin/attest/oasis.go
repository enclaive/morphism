package attest

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/oasisprotocol/oasis-core/go/common/sgx"
	"github.com/oasisprotocol/oasis-core/go/common/sgx/pcs"
	"io"
	"net/http"
	"net/url"
	"time"
)

//goland:noinspection GoUnusedConst
const (
	pccsUrlBase = "https://api.trustedservices.intel.com/sgx/certification/v4"

	pccsPathPckCrl     = "/pckcrl?ca=platform"
	pccsPathTcb        = "/tcb"
	pccsPathQeIdentity = "/qe/identity"
	pccsPathRootCaCrl  = "/rootcacrl"

	pccsHeaderCrlChain = "SGX-PCK-CRL-Issuer-Chain"
	pccsHeaderTcbChain = "TCB-Info-Issuer-Chain"
)

func requestPccs(path string, query url.Values, out interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, pccsUrlBase+path, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, out); err != nil {
		return nil, err
	}

	return res, nil
}

func getTcbBundle(fsmpc []byte) (*pcs.TCBBundle, error) {
	tcbBundle := &pcs.TCBBundle{
		TCBInfo:      pcs.SignedTCBInfo{},
		QEIdentity:   pcs.SignedQEIdentity{},
		Certificates: nil,
	}

	tcbQuery := map[string][]string{"fmspc": {hex.EncodeToString(fsmpc)}}
	tcbResponse, err := requestPccs(pccsPathTcb, tcbQuery, &tcbBundle.TCBInfo)
	if err != nil {
		return nil, err
	}

	certificates, err := url.QueryUnescape(tcbResponse.Header.Get(pccsHeaderTcbChain))
	if err != nil {
		return nil, err
	}

	tcbBundle.Certificates = []byte(certificates)

	_, err = requestPccs(pccsPathQeIdentity, url.Values{}, &tcbBundle.QEIdentity)
	if err != nil {
		return nil, err
	}

	return tcbBundle, nil
}

func verifyQuote(rawQuote []byte) (*sgx.VerifiedQuote, error) {
	var err error

	var quote pcs.Quote
	if err = quote.UnmarshalBinary(rawQuote); err != nil {
		return nil, err
	}

	quoteSignature, ok := quote.Signature.(*pcs.QuoteSignatureECDSA_P256)
	if !ok {
		return nil, errors.New("unsupported attestation key type")
	}

	switch quoteSignature.CertificationData.(type) {
	case *pcs.CertificationData_PCKCertificateChain:
	default:
		return nil, errors.New("unsupported certification data")
	}

	pckInfo, err := quoteSignature.VerifyPCK(time.Now())
	if err != nil {
		return nil, err
	}

	tcbBundle, err := getTcbBundle(pckInfo.FMSPC)
	if err != nil {
		return nil, err
	}

	quotePolicy := &pcs.QuotePolicy{
		Disabled:                   false,
		TCBValidityPeriod:          90, // in days
		MinTCBEvaluationDataNumber: pcs.DefaultMinTCBEvaluationDataNumber,
	}

	attested, err := quote.Verify(quotePolicy, time.Now(), tcbBundle)
	if err != nil {
		return nil, err
	}

	return attested, nil
}
