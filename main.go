package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/adrg/xdg"
	"github.com/getlantern/systray"
	"github.com/ismtabo/time-left/config"
	"github.com/ismtabo/time-left/icon"
)

const (
	TimeLeftFormat = "⏳ Time Left: %s"
	RestTimeFormat = "🍽 Rest: %s"
	EndTimeFormat  = "%s End: %s"
	ReloadText     = "🔄 Reload"
	QuitText       = "❌ Quit"
)

var (
	conf    config.Config
	rest    bool
	done    chan bool
	quit    chan bool
	running bool
)

func main() {
	defaultConfPath := path.Join(xdg.ConfigHome, "time-left", "config.yaml")
	var confOpt string
	flag.StringVar(&confOpt, "config", defaultConfPath, "Path to the config file")
	flag.Parse()
	conf = config.NewConfig(confOpt)
	log.Printf("Version: %s", config.Version)
	log.Printf("build.Time: %s", config.BuildTime)
	log.Printf("build.OS: %s", config.OS)
	log.Printf("Config path: %s", confOpt)
	log.Println("Starting...")
	systray.Run(onReady, onExit)
}

func onExit() {
	// clean up here
	log.Println("Exiting...")
	endTimer()
	close(done)
	log.Println("Waiting for goroutines to finish...")
	quit <- true
	close(quit)
	log.Println("Bye!")
	os.Exit(0)
}

func onReady() {
	done = make(chan bool, 1)
	quit = make(chan bool, 1)

	endOfWorkDay := conf.GetWorkDayEnd(rest)
	systray.SetIcon(icon.Data)
	systray.SetTooltip("Time left until the end of the work day")
	mTimeLeft := systray.AddMenuItem("Time left", "Time left")
	mRest := systray.AddMenuItemCheckbox(fmt.Sprintf(RestTimeFormat, "disabled"), "Rest: click to toggle", false)
	mEnd := systray.AddMenuItem(getEndOfWorkdayTitle(endOfWorkDay), "End of the work day")
	mReload := systray.AddMenuItem(ReloadText, "Reload config")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem(QuitText, "Quit the whole app")

	mQuit.SetIcon(icon.Data)

	go func() {
		for {
			select {
			case <-quit:
				return
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mRest.ClickedCh:
				rest = !rest
				if rest {
					mRest.SetTitle(fmt.Sprintf(RestTimeFormat, "enabled"))
				} else {
					mRest.SetTitle(fmt.Sprintf(RestTimeFormat, "disabled"))
				}
				mEnd.SetTitle(getEndOfWorkdayTitle(conf.GetWorkDayEnd(rest)))
				if updateTimeLeftMenuItem(mTimeLeft, time.Now(), conf.GetWorkDayEnd(rest)) {
					done <- true
				}
			case <-mReload.ClickedCh:
				reloadConfig(mEnd, mTimeLeft)
			}
		}
	}()

	reloadConfig(mEnd, mTimeLeft)
}

func reloadConfig(mEnd *systray.MenuItem, mTimeLeft *systray.MenuItem) {
	if err := conf.Reload(); err != nil {
		panic(err)
	}
	endOfWorkday := conf.GetWorkDayEnd(rest)
	mEnd.SetTitle(getEndOfWorkdayTitle(endOfWorkday))
	startTimer(endOfWorkday, mTimeLeft)
	running = true
}

func startTimer(wordayEnd time.Time, mTimeLeft *systray.MenuItem) {
	if running {
		done <- true
	}
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if updateTimeLeftMenuItem(mTimeLeft, time.Now(), conf.GetWorkDayEnd(rest)) {
					endTimer()
				}
			}
		}
	}()
	if updateTimeLeftMenuItem(mTimeLeft, time.Now(), wordayEnd) {
		endTimer()
	}
}

func endTimer() {
	if !running {
		return
	}
	done <- true
	running = false
}

func updateTimeLeftMenuItem(item *systray.MenuItem, t time.Time, end time.Time) bool {
	if t.After(conf.GetWorkDayEnd(rest)) {
		item.SetTitle(fmt.Sprintf(TimeLeftFormat, "none. See you tomorrow!"))
		return true
	}
	timeLeft := time.Until(end)
	item.SetTitle(fmt.Sprintf(TimeLeftFormat, timeLeft.Truncate(conf.GetTruncateDuration()).String()))
	return false
}

func getEndOfWorkdayTitle(end time.Time) string {
	return fmt.Sprintf(EndTimeFormat, getEndOfWorkdayClockEmoji(end), end.Format("15:04"))
}

func getEndOfWorkdayClockEmoji(end time.Time) string {
	nextHalfHour := end.Round(30 * time.Minute).Format("03:04")
	switch nextHalfHour {
	case "00:00":
		return "🕛"
	case "00:30":
		return "🕧"
	case "01:00":
		return "🕐"
	case "01:30":
		return "🕜"
	case "02:00":
		return "🕑"
	case "02:30":
		return "🕝"
	case "03:00":
		return "🕒"
	case "03:30":
		return "🕞"
	case "04:00":
		return "🕓"
	case "04:30":
		return "🕟"
	case "05:00":
		return "🕔"
	case "05:30":
		return "🕠"
	case "06:00":
		return "🕕"
	case "06:30":
		return "🕡"
	case "07:00":
		return "🕖"
	case "07:30":
		return "🕢"
	case "08:00":
		return "🕗"
	case "08:30":
		return "🕣"
	case "09:00":
		return "🕘"
	case "09:30":
		return "🕤"
	case "10:00":
		return "🕙"
	case "10:30":
		return "🕥"
	case "11:00":
		return "🕚"
	case "11:30":
		return "🕦"
	}
	return "🛑"
}
