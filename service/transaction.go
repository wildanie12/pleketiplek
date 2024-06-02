package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"pertamina-pleketiplek/entity"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/xuri/excelize/v2"
)

type TransactionService struct {
	client http.Client
	token  string
	url    string
}

func NewTransaction() *TransactionService {
	return &TransactionService{
		client: http.Client{},
		url:    "https://api-map.my-pertamina.id/general/v1/transactions/report",
	}
}

func (s *TransactionService) PerDay(token string, dateStart time.Time, dateEnd time.Time) (string, []string, error) {
	s.token = token
	// nerima inputan tanggal awal dan tanggal akhir
	dateEnd = dateEnd.Add(24 * time.Hour)

	// define new excel file
	f := excelize.NewFile()
	defer f.Close()
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:      true,
			Underline: "single",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0EBF5"},
			Pattern: 1,
		},
	})

	files := []string{}

	// bikinperulangan di setiap hari dalam rentang sesuai inputan
	date := dateStart
	previousMonth := dateStart.Format("01-2006")
	fmt.Println("iterating...")
	for !date.Equal(dateEnd) {
		func() {

			defer func() {
				date = date.Add(24 * time.Hour)
			}()

			if date.Format("01-2006") != previousMonth {
				_ = f.DeleteSheet("Sheet1")
				fileName := fmt.Sprintf("Bulan %s.xlsx", previousMonth)
				fmt.Printf("Building file... [%s]\n", fileName)
				files = append(files, fileName)
				if err := f.SaveAs(fileName); err != nil {
					log.Println("SaveAs err:", err)
				}
				f = excelize.NewFile()
				previousMonth = date.Format("01-2006")
			}

			fmt.Println("Processing date:", date.Format("02-01-2006"))

			sheetName := date.Format("02-01-2006")
			_, err := f.NewSheet(sheetName)
			if err != nil {
				log.Println("NewSheet err:", err)
			}
			_ = f.SetCellValue(sheetName, "A1", "NAMA")
			_ = f.SetCellValue(sheetName, "B1", "NIK")
			_ = f.SetCellValue(sheetName, "C1", "QTY BELI")
			_ = f.SetCellValue(sheetName, "D1", "KATEGORI")
			_ = f.SetCellStyle(sheetName, "A1", "D1", headerStyle)

			// nembak  API https://api-map.my-pertamina.id/general/v1/transactions/report
			// kalo kena rate limit, kasih delay semenit lalu coba lagi

			// prepare request
			httpReq, err := http.NewRequest(http.MethodGet, s.url, nil)
			if err != nil {
				log.Println("err http request:", err)
			}

			httpReq.Header.Add("Authorization", strings.TrimSpace("Bearer "+s.token))
			q := httpReq.URL.Query()
			q.Add("startDate", date.Format("2006-01-02"))
			q.Add("endDate", date.Format("2006-01-02"))
			httpReq.URL.RawQuery = q.Encode()

			// execute request
			client := http.Client{}
			httpRes, err := client.Do(httpReq)
			if err != nil {
				log.Println("err http response:", err)
			}

			// decode data
			res := entity.Report{}
			err = json.NewDecoder(httpRes.Body).Decode(&res)
			if err != nil {
				log.Println("err data decoding:", err)
			}

			if !res.Success || httpRes.StatusCode >= 300 {
				fmt.Println("[x] Request Error: ")
				b, _ := json.MarshalIndent(res, "", "  ")
				fmt.Println(string(b))
				return
			}

			type DuplicateData struct {
				Count int
				Cells []string
			}
			duplicates := map[string]DuplicateData{}
			for i, report := range res.Data.CustomersReport {
				row := strconv.Itoa(i + 2)

				// find and map duplicate
				_, exist := duplicates[report.Name+"|"+report.NationalityID]
				if !exist {
					duplicates[report.Name+"|"+report.NationalityID] = DuplicateData{
						Count: 1,
						Cells: []string{"A" + row},
					}
				} else {
					d := duplicates[report.Name+"|"+report.NationalityID]
					duplicates[report.Name+"|"+report.NationalityID] = DuplicateData{
						Count: d.Count + 1,
						Cells: append(d.Cells, "A"+row),
					}
				}

				// write to cells
				category := ""
				if len(report.Categories) > 0 {
					category = report.Categories[0]
				}
				_ = f.SetColWidth(sheetName, "A", "A", 30)
				_ = f.SetColWidth(sheetName, "B", "B", 20)
				_ = f.SetColWidth(sheetName, "D", "D", 20)
				_ = f.SetCellValue(sheetName, "A"+row, report.Name)
				_ = f.SetCellValue(sheetName, "B"+row, report.NationalityID)
				_ = f.SetCellValue(sheetName, "C"+row, report.Total)
				_ = f.SetCellValue(sheetName, "D"+row, category)
			}

			// highlight duplicates
			for _, dd := range duplicates {
				if dd.Count < 2 {
					continue
				}
				style, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{randomcolor.GetRandomColorInHex()},
						Pattern: 1,
					},
				})
				for _, cell := range dd.Cells {
					_ = f.SetCellStyle(sheetName, cell, cell, style)
					_ = f.SetCellStyle(sheetName, "B"+cell[1:], "B"+cell[1:], style)
				}
			}

			// ketika udah dapet datanya, taruh ke file excel.
		}()

	}
	date = date.Add(-24 * time.Hour)

	_ = f.DeleteSheet("Sheet1")
	fileName := fmt.Sprintf("Bulan %s.xlsx", date.Format("01-2006"))
	fmt.Printf("Building file... [%s]\n", fileName)
	if err := f.SaveAs(fileName); err != nil {
		log.Println("SaveAs err:", err)
	}
	files = append(files, fileName)

	// zip
	zipFilename := "laporan_" + dateStart.Format("02_01") + "_to_" + dateEnd.Format("02_01_2006") + ".zip"
	fmt.Println("Creating", zipFilename)
	zipfile, err := os.Create(strings.TrimSpace(zipFilename))
	if err != nil {
		return "", files, err
	}
	defer zipfile.Close()
	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	for _, file := range files {
		defer func(file string) {
			f, err := os.Open(file)
			if err != nil {
				return
			}
			defer f.Close()

			w, err := zipWriter.Create(file)
			if err != nil {
				return
			}
			_, err = io.Copy(w, f)
			if err != nil {
				return
			}

		}(file)
	}

	return zipFilename, files, nil
}
