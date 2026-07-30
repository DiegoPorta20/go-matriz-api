// Package awsauth firma peticiones salientes con SigV4.
//
// Existe porque node-api se expone tras una Lambda Function URL con
// autenticacion IAM: sin firma, AWS rechaza la peticion antes de que llegue al
// servicio. La alternativa era abrir la Function URL a internet y dejar el JWT
// como unica barrera, que convierte un servicio interno en uno publico.
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

// El servicio al que pertenece una Function URL a efectos de firma.
const lambdaServiceName = "lambda"

// RequestSigner firma peticiones HTTP con las credenciales del entorno.
//
// SigV4 no se implementa a mano: es un protocolo criptografico con reglas de
// canonicalizacion muy precisas, y equivocarse produce fallos que solo se ven en
// produccion.
type RequestSigner struct {
	credentials aws.CredentialsProvider
	region      string
	signer      *v4.Signer
}

// NewRequestSigner toma las credenciales de la cadena estandar de AWS: variables
// de entorno, rol de la tarea o rol de la funcion Lambda.
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

// Sign añade la cabecera Authorization con la firma de la peticion.
//
// El cuerpo se pasa aparte y no se lee del request: SigV4 necesita el hash del
// payload, y consumir request.Body aqui lo dejaria vacio para el envio.
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
