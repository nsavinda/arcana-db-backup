.PHONY: keygen all run build

keygen:
	openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:4096
	openssl rsa -in private.pem -pubout -out public.pem

# all: 

# make run
run: build
	./backup-service

# make build
build:
	go build -o backup-service main.go

