package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/tebeka/selenium"
)

func main() {

	browserPath := GetBrowserPath("chromedriver")
	fmt.Println("--")
	fmt.Println(browserPath)
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
	err = loginTxt.SendKeys(os.Getenv("username"))
	if err != nil {
		panic(err)
	}

	passTxt, err := wd.FindElement(selenium.ByID, "Input_Password")
	if err != nil {
		panic(err)
	}

	err = passTxt.SendKeys(os.Getenv("password"))
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

	if err = dateTxt.SendKeys(os.Getenv("date")); err != nil {
		panic(err)
	}

	if err = dateTxt.SendKeys(selenium.EnterKey); err != nil {
		panic(err)
	}
	fmt.Println("waiting")
	wd.SetImplicitWaitTimeout(4 * time.Second)
	fmt.Println("clicking")

	// document.querySelector("#AthleteTheme_wt6_block_wtMainContent_wt9_wtClassTable")
	calendarTableInitial, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
	if err != nil {
		panic(err)
	}
	initialSize, _ := calendarTableInitial.Size()
	tableSizeChanged := func(wd selenium.WebDriver) (bool, error) {
		t, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
		s, _ := t.Size()
		if err != nil {
			return false, err
		}

		return s != initialSize, nil
	}

	wd.Wait(tableSizeChanged)

	calendarTable, err := wd.FindElement(selenium.ByCSSSelector, ".TableRecords")
	if err != nil {
		panic(err)
	}
	text, err := calendarTable.FindElements(selenium.ByTagName, "tr")
	if err != nil {
		panic(err)
	}

	for _, v := range text {
		span, err := v.FindElement(selenium.ByTagName, "span")
		if err != nil {
			panic(err)
		}

		fmt.Println(span.Text())
	}

	// tableTr, err := calendarTable.FindElements(selenium.ByTagName, "tr")
	// if err != nil {
	// 	panic(err)
	// }
	//*[@id="AthleteTheme_wt6_block_wtMainContent_wt9_wtClassTable"]/tbody/tr[4]/td[1]/div/span
	// for _, v := range tableTr {
	//*[@id="WebForm1"]/div[5]/div[2]/div[2]/div/span/table/tbody/tr[3]/td[1]/div/span
	// tr, err := v.FindElement(selenium.ByCSSSelector, "td:nth-child(1) > div > span")
	// if err != nil {
	// 	panic(err)
	// }

	// // text, err := tr.Text()
	// // if err != nil {
	// // 	panic(err)
	// // }

	// text, _ := tr.Text()
	// fmt.Println(text)

	// }

	time.Sleep(10 * time.Second)

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
