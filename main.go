package main

import (
	"flag"
	"fmt"
	"path"
	"time"

	"github.com/adrg/xdg"
	"github.com/getlantern/systray"
	"github.com/ismtabo/time-left/config"
	"github.com/ismtabo/time-left/icon"
)

var conf config.Config
var rest bool
var done chan bool
var quit chan bool

func main() {
	defaultConfPath := path.Join(xdg.ConfigHome, "time-left", "config.yaml")
	var confOpt string
	flag.StringVar(&confOpt, "config", defaultConfPath, "Path to the config file")
	flag.Parse()
	conf = config.NewConfig(confOpt)
	systray.Run(onReady, onExit)
}

func onReady() {
	quit = make(chan bool)
	done = make(chan bool)

	systray.SetIcon(icon.Data)
	systray.SetTooltip("Time left until the end of the work day")
	mTimeLeft := systray.AddMenuItem("Time left", "Time left")
	mRest := systray.AddMenuItem("Rest: disabled", "Rest: click to toggle")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

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
					mRest.SetTitle("Rest: enabled")
				} else {
					mRest.SetTitle("Rest: disabled")
				}
				if updateTimeLeftMenuItem(mTimeLeft, time.Now(), conf.GetWorkDayEnd(rest)) {
					done <- true
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-time.Tick(conf.GetRefreshInterval()):
				if updateTimeLeftMenuItem(mTimeLeft, t, conf.GetWorkDayEnd(rest)) {
					done <- true
				}
			}
		}
	}()

	if updateTimeLeftMenuItem(mTimeLeft, time.Now(), conf.GetWorkDayEnd(rest)) {
		done <- true
	}
}

func updateTimeLeftMenuItem(item *systray.MenuItem, t time.Time, end time.Time) bool {
	if t.After(conf.GetWorkDayEnd(rest)) {
		item.SetTitle("Time Left: none. See you tomorrow!")
		return true
	}
	timeLeft := time.Until(end)
	item.SetTitle(fmt.Sprintf("Time Left: %s", timeLeft.Truncate(conf.GetTruncateDuration()).String()))
	return false
}

func onExit() {
	// clean up here
	done <- true
	quit <- true
}
