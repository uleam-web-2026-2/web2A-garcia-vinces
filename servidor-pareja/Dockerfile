# Etapa 1: compilar la API con Go.
FROM golang:1.27.1 AS compilacion
WORKDIR /src

# Aprovecha la caché de Docker mientras las dependencias no cambien.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/latent .

# Etapa 2: imagen mínima con solo el binario estático.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=compilacion /out/latent /latent

EXPOSE 8081
ENTRYPOINT ["/latent"]
