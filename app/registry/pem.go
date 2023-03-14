package registry

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

type PEMInfo struct {
	CACert       string
	Cert         string
	PrivateKey   string
	X509CaCACert []*x509.Certificate
	X509Cert     []*x509.Certificate
}

func DecodePEM(buf []byte) (*PEMInfo, error) {
	var (
		block     *pem.Block
		caBlock   bytes.Buffer
		certBlock bytes.Buffer
		keyBlock  bytes.Buffer
		pemInfo   PEMInfo
	)

	for {
		block, buf = pem.Decode(buf)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			x509Cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("unable to parse pem block, %w", err)
			}

			if x509Cert.IsCA {
				if err := pem.Encode(&caBlock, block); err != nil {
					return nil, fmt.Errorf("unable to encode block, %w", err)
				}

				pemInfo.X509CaCACert = append(pemInfo.X509CaCACert, x509Cert)
			} else {
				if err := pem.Encode(&certBlock, block); err != nil {
					return nil, fmt.Errorf("unable to encode block, %w", err)
				}

				pemInfo.X509Cert = append(pemInfo.X509Cert, x509Cert)
			}
		} else {
			if err := pem.Encode(&keyBlock, block); err != nil {
				return nil, fmt.Errorf("unable to encode block, %w", err)
			}
		}
	}

	pemInfo.CACert = caBlock.String()
	pemInfo.Cert = certBlock.String()
	pemInfo.PrivateKey = keyBlock.String()

	if pemInfo.PrivateKey == "" {
		return nil, errors.New("no key found in PEM file")
	} else if pemInfo.CACert == "" {
		return nil, errors.New("no CA certs found in PEM file")
	} else if pemInfo.Cert == "" {
		return nil, errors.New("no cert found in PEM file")
	}

	return &pemInfo, nil
}
