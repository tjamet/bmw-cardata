package bmwcardata

//go:generate go run ./cmd/generate/
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0 -generate client,types -o ./auth/auth.go -package auth swagger/auth.json
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0 -generate client,types -o ./cardataapi/cardataapi.go -package cardataapi swagger/cardataapi.json
