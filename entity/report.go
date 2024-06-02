package entity

type Report struct {
	Success bool        `json:"success"`
	Data    Report_Data `json:"data"`
}

type Report_Data struct {
	SummaryReport   []Report_Data_SummaryReport   `json:"summaryReport"`
	CustomersReport []Report_Data_CustomersReport `json:"customersReport"`
}

type Report_Data_SummaryReport struct {
	Sold        uint64 `json:"sold"`
	Modal       uint64 `json:"modal"`
	Profit      uint64 `json:"profit"`
	IncomeMyptm uint64 `json:"incomeMyptm"`
}

type Report_Data_CustomersReport struct {
	CustomerReportID string   `json:"customerReportId"`
	NationalityID    string   `json:"nationalityId"`
	Name             string   `json:"name"`
	Categories       []string `json:"categories"`
	Total            int      `json:"total"`
}
