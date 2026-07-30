package awsauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

const lambdaServiceName = "lambda"

type RequestSigner struct {
	credentials aws.CredentialsProvider
	region      string
	signer      *v4.Signer
}

func NewRequestSigner(ctx context.Context, region string) (*RequestSigner, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws configuration: %w", err)
	}

	return &RequestSigner{
		credentials: awsConfig.Credentials,
		region:      awsConfig.Region,
		signer:      v4.NewSigner(),
	}, nil
}

func (s *RequestSigner) Sign(ctx context.Context, request *http.Request, payload []byte) error {
	credentials, err := s.credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("retrieve aws credentials: %w", err)
	}

	digest := sha256.Sum256(payload)

	err = s.signer.SignHTTP(
		ctx,
		credentials,
		request,
		hex.EncodeToString(digest[:]),
		lambdaServiceName,
		s.region,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	return nil
}
