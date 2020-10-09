package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

func main() {
	classTime := os.Getenv("time")
	if len(classTime) == 0 {
		fmt.Println("Invalid class time")
		os.Exit(1)
	}
	username := os.Getenv("username")
	if len(username) == 0 {
		fmt.Println("Invalid username")
		os.Exit(1)
	}
	password := os.Getenv("password")
	if len(password) == 0 {
		fmt.Println("Invalid password")
		os.Exit(1)
	}
	date := os.Getenv("date")
	if len(date) == 0 {
		fmt.Println("Invalid date")
		os.Exit(1)
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	parsedTime, err := time.Parse("01/02/2006 15:04", fmt.Sprintf("%s %s", date, classTime))
	if err != nil {
		panic(err)
	}
	_ = parsedTime

	// classTimer := time.NewTimer(parsedTime.Sub(time.Now()))
	classTimer := time.NewTimer(5 * time.Second)

	for {
		select {
		case <-classTimer.C:
			DoTheThing(username, password, classTime, date)
		}
	}
}
func DoTheThing(username, password, classTime, date string) {
	// time.NewTimer

	browserPath := GetBrowserPath("chromedriver")
	// browserPath := "/Users/nmartin/go/src/github.com/tebeka/selenium/vendor/chromedriver"
	port, err := pickUnusedPort()

	var opts []selenium.ServiceOption
	service, err := selenium.NewChromeDriverService(browserPath,
		port, opts...)

	if err != nil {
		fmt.Printf("Error starting the ChromeDriver server: %v", err)
	}

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	wd, err := selenium.NewRemote(caps, "http://127.0.0.1:"+strconv.Itoa(port)+"/wd/hub")
	if err != nil {
		panic(err)
	}

	wd.Refresh()

	if err := wd.Get("https://app.wodify.com/SignIn/Login?OriginalURL=&RequiresConfirm=false"); err != nil {
		panic(err)
	}
	wd.SetImplicitWaitTimeout(2 * time.Second)
	loginTxt, err := wd.FindElement(selenium.ByID, "Input_UserName")
	if err != nil {
		panic(err)
	}
	err = loginTxt.SendKeys(username)
	if err != nil {
		panic(err)
	}

	passTxt, err := wd.FindElement(selenium.ByID, "Input_Password")
	if err != nil {
		panic(err)
	}

	err = passTxt.SendKeys(password)
	if err != nil {
		panic(err)
	}

	signinBtn, err := wd.FindElement(selenium.ByXPATH, `//*[@id="FormLogin"]/div[2]/div[5]/button`)
	if err != nil {
		panic(err)
	}

	if err := signinBtn.Click(); err != nil {
		panic(err)
	}

	newTitle := "WOD - Kiosk"
	titleChangeCondition := func(wd selenium.WebDriver) (bool, error) {
		title, err := wd.Title()
		if err != nil {
			return false, err
		}

		return title == newTitle, nil
	}
	wd.Wait(titleChangeCondition)
	calendardLink, err := wd.FindElement(selenium.ByLinkText, "CALENDAR")
	if err != nil {
		panic(err)
	}
	if err := calendardLink.Click(); err != nil {
		panic(err)
	}

	newTitle = "Calendar List - Kiosk"
	wd.Wait(titleChangeCondition)

	dateTxt, err := wd.FindElement(selenium.ByXPATH, `//*[@id="WebForm1"]/div[5]/div[2]/div[2]/div/span/div[2]/table/tbody/tr/td/input[2]`)
	if err != nil {
		panic(err)
	}

	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	dateTxt.SendKeys(selenium.BackspaceKey)
	if err = dateTxt.Clear(); err != nil {
		panic(err)
	}

	if err = dateTxt.SendKeys(date); err != nil {
		panic(err)
	}

	if err = dateTxt.SendKeys(selenium.EnterKey); err != nil {
		panic(err)
	}

	calendarTableInitial, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
	if err != nil {
		panic(err)
	}

	initialText, err := calendarTableInitial.Text()
	if err != nil {
		panic(err)
	}

	tableTextChanged := func(wd selenium.WebDriver) (bool, error) {
		t, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
		if err != nil {
			return false, err
		}
		s, err := t.Text()
		if err != nil {
			return false, err
		}

		return s != initialText, nil
	}

	wd.Wait(tableTextChanged)

	calendarTable, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
	if err != nil {
		panic(err)
	}
	text, err := calendarTable.FindElements(selenium.ByTagName, "tr")
	if err != nil {
		panic(err)
	}

	for _, tr := range text {
		span, err := tr.FindElement(selenium.ByTagName, "span")
		if err != nil {
			panic(err)
		}
		spanText, err := span.Text()
		if err != nil {
			panic(err)
		}
		if strings.Contains(spanText, classTime) {
			td, err := tr.FindElement(selenium.ByXPATH, "//*/td[3]/div/a")
			if err != nil {
				panic(err)
			}

			err = td.Click()
			if err != nil {
				panic(err)
			}
			break
		}

	}

	fmt.Println("Successfully registered")
	defer service.Stop()
}

func pickUnusedPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

func GetBrowserPath(browser string) string {
	if _, err := os.Stat(browser); err != nil {
		path, err := exec.LookPath(browser)
		if err != nil {
			panic("Browser binary path not found")
		}
		return path
	}
	return browser
}
