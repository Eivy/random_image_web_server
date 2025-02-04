FROM golang:alpine as build
# RUN go install golang.org/x/tools/cmd/goimports@latest
# RUN go install honnef.co/go/tools/cmd/staticcheck@latest
# RUN go install github.com/gordonklaus/ineffassign@latest
#RUN go install github.com/gostaticanalysis/nilerr/cmd/nilerr@latest
COPY . /build
WORKDIR /build
RUN go build -o random_image_web_server
FROM alpine as font
RUN apk update && apk add curl && curl -L -o Cica.zip "https://github.com/miiton/Cica/releases/download/v5.0.3/Cica_v5.0.3_without_emoji.zip" && unzip Cica.zip
FROM alpine
COPY --from=build /build/random_image_web_server .
COPY --from=font Cica-Regular.ttf .
CMD ["/random_image_web_server"]
