package main

import (
	"bytes"
	"context"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Client struct {
	GUID        string
	Description string
}

type Store struct {
	GUID        string
	Description string
}

type Good struct {
	GUID        string
	Description string
}

type GoodsItem struct {
	Good  Good
	Price int
	Count int
	Sum   int
}

type Invoice struct {
	Client Client
	Store  Store
	Goods  []GoodsItem
}

// Функция для генерации HTML из шаблона и данных
func generateInvoiceHTML(invoice Invoice) (string, error) {
	// Загружаем HTML-шаблон из файла
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		return "", err
	}

	// Буфер для хранения результата
	var tpl bytes.Buffer
	// Выполняем шаблон с данными
	err = tmpl.Execute(&tpl, invoice)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func generateInvoicePDF(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Пример данных
	log.Println("Формируем данные")
	invoice := Invoice{
		Client: Client{
			GUID:        "CLIENT-GUID1",
			Description: "Ромашка",
		},
		Store: Store{
			GUID:        "STORE-GUID1",
			Description: "Магазин №3",
		},
		Goods: []GoodsItem{
			{
				Good:  Good{GUID: "GOOD-GUID1", Description: "Ручка"},
				Price: 100,
				Count: 10,
				Sum:   1000,
			},
			{
				Good:  Good{GUID: "GOOD-GUID2", Description: "Карандаш"},
				Price: 40,
				Count: 10,
				Sum:   400,
			},
			{
				Good:  Good{GUID: "GOOD-GUID3", Description: "Тетрадь"},
				Price: 140,
				Count: 10,
				Sum:   1400,
			},
		},
	}

	// Генерация HTML на основе шаблона и данных накладной
	log.Println("Создаем верстку")
	htmlContent, err := generateInvoiceHTML(invoice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем временный файл для HTML
	log.Println("Создаем временный файл для HTML")
	tmpFile, err := ioutil.TempFile("", "invoice-*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())

	// Записываем HTML в временный файл
	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpFile.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx1, cancel1 := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("single-process", false),
		chromedp.Flag("disable-dev-shm-usage", true),
	)...)
	defer cancel1()

	log.Println("Создаем контекст для chromedp")
	ctx, cancel2 := chromedp.NewContext(ctx1)
	defer cancel2()

	// Выполняем chromedp задачи для рендеринга HTML в PDF
	var pdfBuffer []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpFile.Name()),
		chromedp.ActionFunc(func(ctx context.Context) error {
			start := time.Now()
			log.Println("Формируем PDF")
			var err error
			pdfBuffer, _, err = page.PrintToPDF().Do(ctx)
			log.Printf("Формирование завершено за %v", time.Since(start))
			return err
		}),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки и отправляем PDF клиенту
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=invoice.pdf")
	log.Printf("Возвращаем результат за %v", time.Since(start))
	_, err = w.Write(pdfBuffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Проверяем наличие шаблона
	if _, err := os.Stat("template.html"); os.IsNotExist(err) {
		panic("Файл шаблона template.html не найден")
	}

	http.HandleFunc("/pdf", generateInvoicePDF)
	http.ListenAndServe(":8081", nil)
}
