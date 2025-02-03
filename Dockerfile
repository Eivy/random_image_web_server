FROM golang:alpine as build
# RUN go install golang.org/x/tools/cmd/goimports@latest
# RUN go install honnef.co/go/tools/cmd/staticcheck@latest
# RUN go install github.com/gordonklaus/ineffassign@latest
#RUN go install github.com/gostaticanalysis/nilerr/cmd/nilerr@latest
COPY . /build
WORKDIR /build
RUN go build -o random_image_web_server
FROM alpine
RUN apk update && apk add chromium nss freetype harfbuzz ca-certificates ttf-freefont curl fontconfig \
		&& curl -O https://noto-website-2.storage.googleapis.com/pkgs/NotoSansCJKjp-hinted.zip \
		&& mkdir -p /usr/share/fonts/NotoSansCJKjp \
		&& unzip NotoSansCJKjp-hinted.zip -d /usr/share/fonts/NotoSansCJKjp/ \
		&& rm NotoSansCJKjp-hinted.zip \
		&& fc-cache -fv
COPY --from=build /build/random_image_web_server .
CMD ["/random_image_web_server"]
