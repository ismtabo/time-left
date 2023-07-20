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

func main() {
	defaultConfPath := path.Join(xdg.ConfigHome, "time-left", "config.yaml")
	var confOpt string
	flag.StringVar(&confOpt, "config", defaultConfPath, "Path to the config file")
	flag.Parse()
	conf = config.NewConfig(confOpt)
	systray.Run(onReady, onExit)
}

func onReady() {
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
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mRest.ClickedCh:
				rest = !rest
				if rest {
					mRest.SetTitle("Rest: enabled")
				} else {
					mRest.SetTitle("Rest: disabled")
				}
				updateTimeLeftMenuItem(mTimeLeft, conf.GetWorkDayEnd(rest))
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-time.Tick(1 * time.Minute):
				if t.After(conf.GetWorkDayEnd(rest)) {
					mTimeLeft.SetTitle("Time Left: none. See you tomorrow!")
					done <- true
					continue
				}
				updateTimeLeftMenuItem(mTimeLeft, conf.GetWorkDayEnd(rest))
			}
		}
	}()

	updateTimeLeftMenuItem(mTimeLeft, conf.GetWorkDayEnd(rest))
}

func updateTimeLeftMenuItem(item *systray.MenuItem, end time.Time) {
	timeLeft := time.Until(end)
	timeLeftString := ""
	if timeLeft.Hours() > 0 {
		timeLeftString += fmt.Sprintf("%dh", int(timeLeft.Hours()))
	}
	if timeLeft.Minutes() > 0 {
		timeLeftString += fmt.Sprintf("%dm", int(timeLeft.Minutes())%60)
	}
	item.SetTitle(fmt.Sprintf("Time Left: %s", timeLeftString))
}

func onExit() {
	// clean up here
	done <- true
}
