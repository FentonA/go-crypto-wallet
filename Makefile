goget:
	go get -u -d -v ./...

bld:
	go build -o wallet ./cmd/wallet/main.go

run: bld
	./wallet

run1: bld
	./wallet -f 1

run2: bld
	./wallet -f 2

run3: bld
	./wallet -f 3

run9: bld
	./wallet -f 9

.PHONY: clean
clean:
	rm -rf detect