gosudoku: solver.go
	go build -o $@ solver.go

main.test: solver.go solver_test.go
	go test -o $@ solver_test.go solver.go

test: main.test
	./main.test -test.bench .
