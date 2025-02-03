package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/image/draw"

	"github.com/chromedp/chromedp"
)

// ハンドラ関数
func handler(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータから "seed" を取得
	seedStr := r.URL.Query().Get("seed")
	var num int64

	if seedStr != "" {
		num = randomFromString(seedStr, threshold)
	} else {
		num = time.Now().UnixNano()
	}

	url := fmt.Sprintf("https://zukan.pokemon.co.jp/detail/%04d", num)
	var imgUrl string
	var name string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(`document.querySelector("main header img").getAttribute("src")`, &imgUrl),
		chromedp.Evaluate(`document.querySelector("main header img").getAttribute("alt")`, &name),
	)
	log.Println(seedStr, num, name)
	if err != nil {
		log.Println("[error]", "request error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := http.DefaultClient.Get(imgUrl)
	defer res.Body.Close()
	img, err := png.Decode(res.Body)
	if err != nil {
		log.Println("[error]", "request error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	img2 := ResizeImage(img, 128, 128)

	w.Header().Add("Content-Type", res.Header.Get("Content-Type"))
	w.Header().Add("Content-Length", res.Header.Get("Content-Length"))
	w.WriteHeader(res.StatusCode)
	png.Encode(w, img2)
}

func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
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
var ctx context.Context

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
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // headless=false に変更
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	http.HandleFunc("/", handler)

	fmt.Println("Starting server")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
