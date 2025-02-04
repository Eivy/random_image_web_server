package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	gq "github.com/PuerkitoBio/goquery"
	"github.com/golang/freetype/truetype"
)

// ハンドラ関数
func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// クエリパラメータから "seed" を取得
	seedStr := query.Get("seed")
	var num int64
	if seedStr != "" {
		num = randomFromString(seedStr, threshold)
	} else {
		num = time.Now().UnixNano() % threshold
	}

	url := fmt.Sprintf("https://zukan.pokemon.co.jp/detail/%04d", num)
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		log.Println("[error]", "request root", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("[error]", "parse", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonElm := doc.Find(`script#json-data`).First()
	if jsonElm == nil {
		log.Println("[error]", "getting element", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	text := jsonElm.Text()
	var d Data
	err = json.Unmarshal([]byte(text), &d)
	if err != nil {
		log.Println("[error]", "getting element", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println(seedStr, num, d.Pokemon.Name)

	res, err = http.DefaultClient.Get(d.Pokemon.ImageS)
	if err != nil {
		log.Println("[error]", "request img", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	img, err := png.Decode(res.Body)
	if err != nil {
		log.Println("[error]", "request error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	x, err := strconv.Atoi(query.Get("x"))
	if err != nil {
		x = img.Bounds().Dx()
	}
	y, err := strconv.Atoi(query.Get("y"))
	if err != nil {
		y = img.Bounds().Dy()
	}

	img2 := ResizeImage(img, x, y)
	face := truetype.NewFace(cicaFont, &truetype.Options{
		Size: 15,
	})
	bounds := img2.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img2, bounds.Min, draw.Src)

	draw := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(y)},
	}
	draw.DrawString(d.Pokemon.Name)

	var b bytes.Buffer
	png.Encode(&b, rgba)
	w.Header().Add("Content-Length", fmt.Sprint(len(b.Bytes())))
	w.Header().Add("Content-Type", "image/png")
	w.WriteHeader(res.StatusCode)
	io.Copy(w, &b)
}

type Data struct {
	Pokemon struct {
		Name   string `json:"name"`
		ImageS string `json:"image_s"`
	} `json:"pokemon"`
}

func ResizeImage(img image.Image, width, height int) image.Image {
	// 欲しいサイズの画像を新しく作る
	newImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// サイズを変更しながら画像をコピーする
	draw.BiLinear.Scale(newImage, newImage.Bounds(), img, img.Bounds(), draw.Over, nil)

	return newImage
}

func randomFromString(s string, threshold int64) int64 {
	// SHA-256ハッシュを生成
	hash := sha256.Sum256([]byte(s))

	// ハッシュの一部を数値に変換してシードとして使用
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	// ランダムジェネレータを初期化
	rnd := rand.New(rand.NewSource(seed))

	// 0から100までのランダムな数値を返す
	return rnd.Int63n(threshold)
}

var threshold int64
var cicaFont *truetype.Font

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	var err error
	threshold, err = strconv.ParseInt(os.Getenv("THRESHOLD"), 10, 64)
	if err != nil {
		panic(err)
	}

	ttf, err := os.ReadFile("./Cica-Regular.ttf")
	if err != nil {
		panic(err)
	}
	cicaFont, err = truetype.Parse(ttf)

	http.HandleFunc("/", handler)
	fmt.Println("Starting server")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
