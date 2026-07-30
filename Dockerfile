# Fiber v3.4.0 exige Go 1.25 como minimo.
FROM golang:1.25-alpine AS build

WORKDIR /src

# Las dependencias se copian antes que el codigo para que la capa de descarga
# sobreviva a cualquier cambio en los fuentes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO desactivado para obtener un binario estatico que corra en una imagen sin
# libc. -trimpath quita las rutas de compilacion, que de otro modo aparecerian en
# los mensajes de panico.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api ./cmd/api

FROM alpine:3.22 AS runtime

# No se instala nada: el healthcheck usa el wget de busybox, que alpine ya trae.
RUN adduser --disabled-password --gecos "" --uid 10001 appuser

# AWS Lambda Web Adapter. Traduce el evento de Lambda a una peticion HTTP contra
# el servidor que ya tenemos, asi que Fiber no necesita un handler especial.
#
# /opt/extensions solo lo lee el runtime de Lambda: fuera de Lambda nadie arranca
# este binario, asi que LA MISMA IMAGEN sigue funcionando en docker compose.
#
# Con Fiber importa: usa fasthttp y no net/http, asi que los adaptadores clasicos
# tipo aws-lambda-go-api-proxy encajan mal. Al adaptador, que habla HTTP contra
# localhost, le da igual el framework.
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 \
     /lambda-adapter /opt/extensions/lambda-adapter

# El adaptador necesita saber en que puerto escucha la aplicacion y cuando esta
# lista. Ambas variables las ignora cualquier entorno que no sea Lambda.
ENV AWS_LWA_PORT=8080
ENV AWS_LWA_READINESS_CHECK_PATH=/health

COPY --from=build /out/api /usr/local/bin/api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/api"]
