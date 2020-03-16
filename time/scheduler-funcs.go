package time

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/hellgate75/go-tcp-common/log"
	"strings"
	"sync"
	"time"
)

type cronJob struct{
	sync.Mutex
	_running	bool
	_data		CronData
	_lastRun	time.Time
	_id			CronUUID
	_name		string
	_func		func()()
	_counter	int64
	_logger		log.Logger
	_timer		*time.Ticker
	_done 		chan bool

}

func format(size int, value interface{}) string{
	text := fmt.Sprintf("%v", value)
	length := len(text)
	if length > size {
		text = text[:size-3] + "..."
	} else if length < size {
		text += strings.Repeat(" ", size - length)
	}
	return text
}
func (cj *cronJob) String() string {
	lastRun := "<N.D.>"
	if ! cj._lastRun.IsZero() {
		lastRun = cj._lastRun.String()
	}
	return fmt.Sprintf("CronJob{Name: \"%s\", Uuid: \"%v\", Running: %v, Last Run %s}",
						cj._name, cj._id, cj._running, lastRun)
}

func (cj *cronJob) Report() string {
	lastRun := "<N.D.>"
	if ! cj._lastRun.IsZero() {
		lastRun = cj._lastRun.String()
	}
	state := " Stopped"
	if cj._running {
		state=" Running"
	}
	return "| " + format(30, cj._name) + "| " + format(CronUuidStandardLength + 6, cj._id) + "| " +
		format(9, state) + "|" + format(30, lastRun)
}

func (cj *cronJob) Label() string {
	return cj._name
}

func durationToNumber(duration time.Duration, unit time.Duration) int64 {
	return int64(duration/unit)
}

func clockUnit(duration time.Duration) (time.Duration, time.Duration) {
	if durationToNumber(duration, time.Millisecond) > 1000 {
		if durationSeconds := durationToNumber(duration, time.Second); durationSeconds > 2 {
			mod := durationSeconds % 2
			return 2 * time.Second, time.Duration(mod) * time.Second
		}
	}
	return duration, 0 * time.Second
}

func durationDiff(d1 time.Duration, d2 time.Duration) int64 {
	return durationToNumber(d1, time.Nanosecond) - durationToNumber(d2, time.Nanosecond)
}
func (cj *cronJob) _execute() {
	cj._lastRun = time.Now()
	go func(cron *cronJob) {
		if cron._data.NumExecutions > 0 {
			cron._counter++
			if cron._counter >= cron._data.NumExecutions {
				cron.Stop()
			}
		}
	}(cj)
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronJob.<execution> - Errors running Job #%v %s , Details: %v", cj._id, cj._name, r)
			if cj._logger != nil {
				cj._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
		}
	}()
	cj._func()
}

func (cj *cronJob) _run() {
	cj._running = true
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronJob.<process> - Errors running Job #%v %s , Details: %v", cj._id, cj._name, r)
			if cj._logger != nil {
				cj._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
		}
	}()
	time.Sleep(cj._data.Delay)
	cj._timer = time.NewTicker(cj._data.Interval)
	cj._done = make(chan bool)
	for cj._running {
		select {
		case <-cj._done:
			break
		case <-cj._timer.C:
			cj._execute()
		}
	}
}

func (cj *cronJob) Start() error {
	if cj.IsRunning() {
		return errors.New(fmt.Sprintf("time.CronJob.Start - Job #%v %s is already running!!", cj._id, cj._name))
	}
	if "" == cj._id {
		cj._id = GenerateCronId()
	}
	var locked bool = false
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("time.CronJob.Start - Unable to discover nodes, Details: %v", r))
		}
		if locked {
			cj.Unlock()
		}
	}()
	cj.Lock()
	locked = true

	go cj._run()
	go func() {
		message := fmt.Sprintf("time.CronJob.Start - Job #%v %s Started!!", cj._id, cj._name)
		if cj._logger != nil {
			cj._logger.Info(message)
		} else {
			color.LightWhite.Println(message)
		}
	}()
	cj.Unlock()
	locked = false
	return err
}

func (cj *cronJob) Stop() error {
	if ! cj.IsRunning() {
		return errors.New(fmt.Sprintf("time.CronJob.Stop - Job #%v %s is not running!!", cj._id, cj._name))
	}
	var locked bool = false
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("time.CronJob.Stop - Unable to discover nodes, Details: %v", r))
		}
		if locked {
			cj.Unlock()
		}
	}()
	cj.Lock()
	locked = true
	cj._running = false
	if cj._timer != nil {
		cj._timer.Stop()
		cj._done <- true
		go func() {
			message := fmt.Sprintf("time.CronJob.Stop - Job #%v %s Stopped!!", cj._id, cj._name)
			if cj._logger != nil {
				cj._logger.Info(message)
			} else {
				color.LightWhite.Println(message)
			}
		}()
		cj._timer = nil
		close(cj._done)
	} else {
		message := fmt.Sprintf("time.CronJob.Stop - Job #%s %v No timer to Stop!!", cj._id, cj._name)
		if cj._logger != nil {
			cj._logger.Warn(message)
		} else {
			color.LightYellow.Println(message)
		}
	}
	cj.Unlock()
	locked = false
	return err
}
func (cj *cronJob) IsRunning() bool {
	return cj._running
}
func (cj *cronJob) Update(data CronData) error {
	return nil
}
func (cj *cronJob) LastRun() time.Time {
	return cj._lastRun
}
func (cj *cronJob) Equals(job CronJob) bool {
	if job != nil {
		return string(cj._id) == string(job.Id())
	}
	return false
}

func (cj *cronJob) Id() CronUUID {
	return cj._id
}

