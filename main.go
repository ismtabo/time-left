package main

import (
	"encoding/json"
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
	"github.com/zserge/lorca"
)

const (
	TimeLeftFormat = "⏳ Time Left: %s"
	RestTimeFormat = "🍽 Rest: %s"
	EndTimeFormat  = "%s End: %s"
	ReloadText     = "🔄 Reload"
	ConfigText     = "⚙️ Config"
	QuitText       = "❌ Quit"
)

var (
	conf         config.Config
	rest         bool
	done         chan bool
	quit         chan bool
	running      bool
	activeConfig lorca.UI
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
	mConfig := systray.AddMenuItem(ConfigText, "Open config")
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
			case <-mConfig.ClickedCh:
				go openConfig(func() {
					reloadConfig(mEnd, mTimeLeft)
				})
			}
		}
	}()

	reloadConfig(mEnd, mTimeLeft)
}

type onConfigSavedCallback func()

func openConfig(callback onConfigSavedCallback) {
	if activeConfig != nil {
		log.Println("Config UI is already open")
		return
	}
	ui, err := lorca.New("", "", 860, 480)
	if err != nil {
		log.Printf("Failed to open config UI: %v", err)
		return
	}
	activeConfig = ui
	defer func() {
		activeConfig = nil
		ui.Close()
	}()
	ui.Bind("handleSaveConfig", func(configData string) {
		var data struct {
			WorkdayStart    string `json:"workdayStart"`
			WorkdayDuration string `json:"workdayDuration"`
			RestDuration    string `json:"restDuration"`
		}
		if err := json.Unmarshal([]byte(configData), &data); err != nil {
			log.Printf("Failed to parse config data: %v", err)
			return
		}
		workdayStart, err := time.Parse("15:04", data.WorkdayStart)
		if err != nil {
			log.Printf("Failed to parse workday start time: %v", err)
			return
		}
		workdayDuration, err := time.ParseDuration(data.WorkdayDuration)
		if err != nil {
			log.Printf("Failed to parse workday duration: %v", err)
			return
		}
		restDuration, err := time.ParseDuration(data.RestDuration)
		if err != nil {
			log.Printf("Failed to parse rest duration: %v", err)
			return
		}
		conf.SetWorkDayStart(workdayStart)
		conf.SetWorkDayDuration(workdayDuration)
		conf.SetRestDuration(restDuration)
		if err := conf.Save(); err != nil {
			log.Printf("Failed to save config: %v", err)
			return
		}
		if callback != nil {
			callback()
		}
	})
	ui.Load("data:text/html," + (`
<!DOCTYPE html>
<html lang="en">
<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Time Left Config</title>
		<style>
				body {
						font-family: Arial, sans-serif;
						padding: 20px;
				}
				label {
						display: block;
						margin-top: 10px;
				}
				input[type="text"] {
						width: 100%;
						padding: 8px;
						box-sizing: border-box;
				}
				button {
						margin-top: 20px;
						padding: 10px 20px;
						font-size: 16px;
				}
		</style>
</head>
<body>
		<h1>Time Left Config</h1>
		<form id="configForm" action="javascript:void(0);" onsubmit="saveConfig()">
		<label for="workdayStart">Workday Start Time (HH:mm):</label>
		<input type="text" id="workdayStart" value="` + conf.GetWorkDayStart().Format("15:04") + `">
		<label for="workdayDuration">Workday Duration (e.g. 8h):</label>
		<input type="text" id="workdayDuration" value="` + conf.GetWorkDayDuration().String() + `">
		<label for="restDuration">Rest Duration (e.g. 1m, 30s):</label>
		<input type="text" id="restDuration" value="` + conf.GetRestDuration().String() + `">
		<button type="submit">Save</button>
		</form>
		<script>
				function saveConfig() {
						const data = document.getElementById('configForm');
						const workdayStart = data.workdayStart.value;
						const workdayDuration = data.workdayDuration.value;
						const restDuration = data.restDuration.value;
						handleSaveConfig(JSON.stringify({ workdayStart, workdayDuration, restDuration }));
				}
		</script>
</body>
</html>
	`))
	<-ui.Done()
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
