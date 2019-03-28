go get -u github.com/Masterminds/glide
go get -u golang.org/x/lint/golint
go get github.com/axw/gocov/gocov
go get github.com/mattn/goveralls
go get golang.org/x/tools/cmd/cover
ccm create test -v 3.9 -n 1 -s

make cover_ci ; else echo 'skipping unit tests'; fi
