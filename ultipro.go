package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

type Person struct {
	ID           string
	Name         string
	ReportsCount int
	SupervisorID string
}

type PersonWithReports struct {
	ID                string              `json:"employeeId"`
	Name              string              `json:"employeeName"`
	SupervisorID      string              `json:"supervisorId"`
	DirectReportCount int                 `json:"directReportCount"`
	Reports           []PersonWithReports `json:"children"`
	// Supervisors                       []interface{}       `json:"supervisors"`
	// Collapsed                         bool                `json:"collapsed"`
	// CompanyID                         string              `json:"companyId"`
	// DirectSupervisorsLocalizedTooltip string              `json:"directSupervisorsLocalizedTooltip"`
	// Focused                           bool                `json:"focused"`
	// FullJobDescription                bool                `json:"fullJobDescription"`
	// IDandCompany                      string              `json:"id"`
	// ImageURL                          string              `json:"imageUrl"`
	// IndirectReportCount               int                 `json:"indirectReportCount"`
	// IndirectReportCountVisible        bool                `json:"indirectReportCountVisible"`
	// IndirectReportCSSStyle            string              `json:"indirectReportCssStyle"`
	// Left                              int                 `json:"left"`
	// Level                             int                 `json:"level"`
	// PhotoVisible                      bool                `json:"photoVisible"`
	// Selected                          bool                `json:"selected"`
	// SupervisorCount                   int                 `json:"supervisorCount"`
	// SupervisorInfo                    interface{}         `json:"supervisorInfo"`
	// ThumbnailFields                   []struct {
	// 	BaseID         string      `json:"baseId"`
	// 	DropdownValues interface{} `json:"dropdownValues"`
	// 	Section        interface{} `json:"section"`
	// 	TemplateName   string      `json:"templateName"`
	// 	Title          string      `json:"title"`
	// 	ValueSelected  interface{} `json:"valueSelected"`
	// 	Values         []string    `json:"values"`
	// } `json:"thumbnailFields"`
	// Top int `json:"top"`
}

func GetAllReports(browser *http.Client, person *PersonWithReports, people []*Person, indexes []int) []*Person {
	if verbose {
		fmt.Println(person.Name, person.Title, person.DirectReportCount, indexes, len(people))
	}

	people = append(people, &Person{person.ID, person.Name, person.DirectReportCount, person.SupervisorID})

	if person.DirectReportCount == 0 {
		return people
	}

	root, err := GetDirectReports(browser, person.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Listing a users reports still returns the root
	// So we keep track of index (where we are in the tree) and go back there
	for _, n := range indexes {
		root = &root.Reports[n]
	}

	for i, p := range root.Reports {
		people = GetAllReports(browser, &p, people, append(indexes, i))
	}

	return people
}

func Login() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		log.Fatal(err)
	}

	browser := &http.Client{
		Jar: jar,
	}

	resp, err := browser.Get("https://nw11.ultipro.com/Login.aspx")

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	tokens := map[string]string{}

	doc.Find(`input[id^="__"]`).Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		value, _ := s.Attr("value")

		tokens[id] = value
	})

	if err != nil {
		return nil, err
	}

	form := url.Values{}
	for key, value := range tokens {
		form.Add(key, value)
	}
	form.Add("ctl00$Content$Login1$UserName", os.Getenv("ULTIPRO_USERNAME"))
	form.Add("ctl00$Content$Login1$Password", os.Getenv("ULTIPRO_PASSWORD"))
	form.Add("ctl00$Content$Login1$LoginButton", "Log in")
	form.Add("ctl00$Content$languageSelection", "0")

	data := form.Encode()

	req, err := http.NewRequest("POST", "https://nw11.ultipro.com/Login.aspx", strings.NewReader(data))
	// req, err := http.NewRequest("POST", "https://nw11.ultipro.com/Login.aspx", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = browser.Do(req)

	if err != nil {
		return nil, err
	}

	return browser, nil
}

func GetDirectReports(browser *http.Client, employeeID string) (*PersonWithReports, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://nw11.ultipro.com/services/OrganizationWebService.svc/OrgHierarchy?coid=ZGFMI&eeid=%s&_=1594948614950", employeeID), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json, text/javascript, */*")

	resp, err := browser.Do(req)

	if err != nil {
		return nil, err
	}

	data := struct {
		OrgChart PersonWithReports `json:"orgChart"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return &data.OrgChart, nil
}
