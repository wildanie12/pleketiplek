package handler

import (
	"net/http"
	"os"
	"pertamina-pleketiplek/service"
	"time"

	"github.com/golang-module/carbon"
	"github.com/labstack/echo/v4"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
}

func NewTransaction(tSrv *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: tSrv,
	}
}

func (h *TransactionHandler) Input(c echo.Context) error {
	html := `
		<html>
			<head>
				<style>
					body {
						font-size: 1.2rem;
						font-family: sans-serif;
					}
					textarea, button {
						display: block;
						width: 100%;
					}
					textarea {
						min-height: 120px;
					}
					textarea, input {
						padding: 8px;
						font-size: 1.2rem;
					}
					button {
						font-size: 1.2rem;
						padding: 8px;
					}
					form {
						display: flex;
						flex-direction: column;
						gap: 8px;
					}
				</style>
			</head>
			<body style='display: flex; align-items: center; justify-content: center; width: 100%; height: 100vh'>
				<form method="POST" action="/perday">
					<textarea type="text" placeholder="Masukkan token disini" name="token"></textarea>
					<div>
						Mulai dari tanggal : <input type="date" name="date_start"  value="` + time.Now().Add(-7*24*time.Hour).Format("2006-01-02") + `">
					</div>
					<div>
						Sampai tanggal: <input type="date" name="date_end" value="` + time.Now().Format("2006-01-02") + `">
					</div>
					<button type="submit">Submit</submit>
				</form>
			</body>
		</html>
	`
	return c.HTML(http.StatusOK, html)
}

func (h *TransactionHandler) ProcessPerDay(c echo.Context) error {
	token := c.FormValue("token")
	dateStart := carbon.Parse(c.FormValue("date_start")).ToStdTime()
	dateEnd := carbon.Parse(c.FormValue("date_end")).ToStdTime()

	file, files, _ := h.transactionService.PerDay(token, dateStart, dateEnd)
	defer func() {
		os.Remove(file)
		for _, file := range files {
			_ = os.Remove(file)
		}
	}()
	return c.File(file)
}
