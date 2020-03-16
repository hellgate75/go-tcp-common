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
		format(9, state) + "|" + format(30, lastRun) + "|"
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
	if cj._running {
		return
	}
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
	if ! cj._data.Since.IsZero() && time.Now().Sub(cj._data.Since) < 0 {
		cj._timer = time.NewTicker(cj._data.Since.Sub(time.Now()))
		cj._done = make(chan bool)
		BeforeTimeAwake:
		for {
			select {
			case <-cj._done:
				break BeforeTimeAwake
			case <-cj._timer.C:
				break BeforeTimeAwake
			}
		}
	}
	time.Sleep(cj._data.Delay)
	cj._timer = time.NewTicker(cj._data.Interval)
	cj._done = make(chan bool)
	cj._execute()
	BeforeTimer:
	for cj._running {
		select {
		case <-cj._done:
			break BeforeTimer
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
		message := fmt.Sprintf("time.CronJob.Stop - Job #%s %v -> No timer to Stop!!", cj._id, cj._name)
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

type cronTab struct {
	sync.Mutex
	_id			CronUUID
	_name		string
	jobs		[]CronJob
	_running	bool
	_logger		log.Logger
	_timer		*time.Ticker
	_done		chan bool
}
func (ct *cronTab) contains(job CronJob) bool {
	for _, jobX := range ct.jobs {
		if jobX.Equals(job) {
			return true
		}
	}
	return false
}

func (ct *cronTab) AddJob(job CronJob) bool {
	var state bool = true
	if job != nil && !ct.contains(job) {
		var locked bool = false
		defer func(){
			if r := recover(); r != nil {
				message := fmt.Sprintf("time.CronTab.AddJob - Error adding Job %s Error: %v", job.String(), r)
				if ct._logger != nil {
					ct._logger.Error(message)
				} else {
					color.LightRed.Println(message)
				}
				state = false
			}
			if locked {
				ct.Unlock()
			}

		}()
		ct.Lock()
		locked = true
		if ct.IsRunning() && !job.IsRunning() {
			job.Start()
		}
		ct.jobs = append(ct.jobs, job)
		ct.Unlock()
		locked = false
	} else {
		message := fmt.Sprintf("time.CronTab.AddJob - Job %v is Nil or duplicte!!", job)
		if ct._logger != nil {
			ct._logger.Warn(message)
		} else {
			color.LightYellow.Println(message)
		}
		state = false
	}
	return state
}
func (ct *cronTab) RemoveJob(uuid CronUUID) bool  {
	var locked bool = false
	var state bool = false
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronTab.RemoveJob - Error removing Job by Uuid %v Error: %v", uuid, r)
			if ct._logger != nil {
				ct._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
			state = false
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	var jobs []CronJob = make([]CronJob, 0)
	for _, job := range ct.jobs {
		if string(job.Id()) == string(uuid) {
			state = true
		} else {
			jobs = append(jobs, job)
		}
	}
	if state {
		ct.jobs = jobs
	}
	ct.Unlock()
	locked = false
	return state
}
func (ct *cronTab) ListJobs() []CronJob {
	return ct.jobs
}

func (ct *cronTab) _run() {
	if ct._running {
		return
	}
	ct._running = true
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronTab.<process> - Error Starting all Jobs, Details: %v", r)
			if ct._logger != nil {
				ct._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
			ct.Stop()
		}
	}()
	ct.StartAllJobs()
	ct._timer = time.NewTicker(CronTabDefaultPurgeTimeout)
	ct._done = make(chan bool)
	BeforeTimer:
	for ct._running {
		select {
		case <-ct._done:
			break BeforeTimer
		case <-ct._timer.C:
			ct.Purge()
		}
	}
}

func (ct *cronTab) Start() error {
	if ct.IsRunning() {
		return errors.New(fmt.Sprintf("time.CronTab.Stop - CronTab #%v %s is already running!!", ct._id, ct._name))
	}
	var locked bool = false
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("time.CronTab.Start - Unable to start CronTab, Details: %v", r))
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	go ct._run()
	go func(){
		message := fmt.Sprintf("time.CronTab.Start - CronTab #%v %s Started!!", ct._id, ct._name)
		if ct._logger != nil {
			ct._logger.Info(message)
		} else {
			color.LightWhite.Println(message)
		}
	}()
	ct.Unlock()
	locked = false
	return err
}
func (ct *cronTab) Stop() error {
	if ! ct.IsRunning() {
		return errors.New(fmt.Sprintf("time.CronTab.Stop - CronTab #%v %s is not running!!", ct._id, ct._name))
	}
	var locked bool = false
	var err error
	defer func(){
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("time.CronTab.Stop - Unable to stop CronTab, Details: %v", r))
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	ct._running = false
	if ct._timer != nil {
		ct._timer.Stop()
		ct._done <- true
		go func() {
			message := fmt.Sprintf("time.CronTab.Stop - CronTab #%v %s Stopped!!", ct._id, ct._name)
			if ct._logger != nil {
				ct._logger.Info(message)
			} else {
				color.LightWhite.Println(message)
			}
		}()
		ct._timer = nil
		close(ct._done)
	} else {
		message := fmt.Sprintf("time.CronTab.Stop - CronTab #%s %v -> No timer to Stop!!", ct._id, ct._name)
		if ct._logger != nil {
			ct._logger.Warn(message)
		} else {
			color.LightYellow.Println(message)
		}
	}
	ct.Unlock()
	locked = false
	return err
}
func (ct *cronTab) IsRunning() bool {
	return ct._running
}
func (ct *cronTab) StartAllJobs() bool {
	var locked bool = false
	var state bool = false
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronTab.StartAllJobs - Error Starting all Jobs, Details: %v", r)
			if ct._logger != nil {
				ct._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
			state = false
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	for _, job := range ct.jobs {
		if !job.IsRunning() {
			job.Start()
			state = true
		}
	}
	ct.Unlock()
	locked = false
	return state
}
func (ct *cronTab) KillAllJobs() bool {
	var locked bool = false
	var state bool = false
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronTab.KillAllJobs - Error Killing all Jobs, Details: %v", r)
			if ct._logger != nil {
				ct._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
			state = false
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	for _, job := range ct.jobs {
		if job.IsRunning() {
			job.Stop()
			state = true
		}
	}
	ct.Unlock()
	locked = false
	return state
}
func (ct *cronTab) Purge() bool {
	var locked bool = false
	var state bool = false
	defer func(){
		if r := recover(); r != nil {
			message := fmt.Sprintf("time.CronTab.Purge - Error Purging all Jobs, Details: %v", r)
			if ct._logger != nil {
				ct._logger.Error(message)
			} else {
				color.LightRed.Println(message)
			}
			state = false
		}
		if locked {
			ct.Unlock()
		}
	}()
	ct.Lock()
	locked = true
	var jobs []CronJob = make([]CronJob, 0)
	for _, job := range ct.jobs {
		if ! job.IsRunning() {
			state = true
		} else {
			jobs = append(jobs, job)
		}
	}
	if state {
		ct.jobs = jobs
	}
	ct.Unlock()
	locked = false
	return state
}
func (ct *cronTab) Equals(tab CronTab) bool {
	if tab != nil {
		return string(ct._id) == string(tab.Id())
	}
	return false
}
func (ct *cronTab) Id() CronUUID {
	return ct._id
}
func (ct *cronTab) String() string {
	return fmt.Sprintf("CronTab{Name: \"%s\", Running: %v, Uuid: \"%v\", Jobs: %v}",
		ct._name, ct._running, ct._id, ct.jobs)
}
func (ct *cronTab) Report() string {
	header := "| " + format(30, "Label") + "| " + format(CronUuidStandardLength + 6, "Cron UUID") + "| " +
		format(9, "State") + "|" + format(30, "Last Run") + "|"
	body := ""
	jobsLen := len(ct.jobs)
	running := 0
	stopped := 0
	for _, job := range ct.jobs {
		body += job.Report() + "\n"
		if job.IsRunning() {
			running ++
		} else {
			stopped ++
		}
	}
	if jobsLen == 0 {
		body += "No Job Scheduled\n"
	}
	footer := fmt.Sprintf("------------\n Total Jobs: %v   Running: %v   Stopped: %v  Running: %v", jobsLen, running, stopped, ct._running)
	return header + body + footer
}

