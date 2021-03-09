FROM golang

ENV WORKDIR /app
WORKDIR ${WORKDIR}

COPY ./ ${WORKDIR}

RUN GOOS=linux go test -v ./...
RUN GOOS=linux go build -o ${TARGETDIR}/main .

