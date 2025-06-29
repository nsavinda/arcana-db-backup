.PHONY: keygen run build

keygen:
	openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:4096
	openssl rsa -in private.pem -pubout -out public.pem

run: build
	./arcanadbbackup

build:
	go build -o arcanadbbackup main.go
