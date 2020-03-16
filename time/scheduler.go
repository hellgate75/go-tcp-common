package time

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/hellgate75/go-tcp-common/log"
	uuid2 "github.com/satori/go.uuid"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	CronUuidReparatorLeg int = 8
	CronUuidStandardLength int = 40
)


//
func salt(t time.Time) int64 {
	time.Sleep(100 * time.Millisecond)
	return 7 * durationToNumber(time.Now().Sub(t), time.Nanosecond)
}

// Generate a Cron eleemnt unique id of given length
func GenerateCronId() string {
	var floatNum float64  = (float64(CronUuidStandardLength)/float64(CronUuidReparatorLeg))
	var tokens int = int(math.Floor(floatNum))
	if CronUuidStandardLength < 10 {
		tokens = 0
	}
	b := make([]byte, CronUuidStandardLength-tokens)
	if _, err := rand.Read(b); err != nil {
		uuid, err :=  uuid2.NewV4()
		if err != nil {
			return fmt.Sprintf("cron-%v-%v-cron", time.Now().Unix(), salt(time.Now()))
		}
		return uuid.String()
	}
	if tokens > 0 {
		hexList := strings.Split(hex.EncodeToString(b), "")
		uuid := "cron-"
		for i, digit := range hexList {
			if i % CronUuidReparatorLeg == 0 {
				uuid += "-"
			}
			uuid += digit
		}
		return uuid + "-cron"
	}
	return "cron-" + hex.EncodeToString(b) + "-cron"
}


type CronData struct {
	Delay			time.Duration
	Interval		time.Duration
	NumExecutions	int64
}

type CronJob interface {
	Name() string
	Start() error
	Stop() error
	IsRunning() bool
	Update(data CronData) error
	LastRun() time.Time
	Equals(job CronJob) bool
	Id() string
}

type cronJob struct{
	sync.Mutex
	_running	bool
	_data		CronData
	_lastRun	time.Time
	_id			string
	_name		string
	_func		func()()
	_counter	int64
	_logger		log.Logger
	_timer		*time.Ticker
	_done 		chan bool

}

func (cj *cronJob) Name() string {
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
			message := fmt.Sprintf("time.CronJob.<execution> - Errors running Job #%s %s , Details: %v", cj._id, cj._name, r)
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
			message := fmt.Sprintf("time.CronJob.<process> - Errors running Job #%s %s , Details: %v", cj._id, cj._name, r)
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
		return errors.New(fmt.Sprintf("time.CronJob.Start - Job #%s %s is already running!!", cj._id, cj._name))
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
		message := fmt.Sprintf("time.CronJob.Start - Job #%s %s Started!!", cj._id, cj._name)
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
		return errors.New(fmt.Sprintf("time.CronJob.Stop - Job #%s %s is not running!!", cj._id, cj._name))
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
			message := fmt.Sprintf("time.CronJob.Stop - Job #%s %s Stopped!!", cj._id, cj._name)
			if cj._logger != nil {
				cj._logger.Info(message)
			} else {
				color.LightWhite.Println(message)
			}
		}()
		cj._timer = nil
		close(cj._done)
	} else {
		message := fmt.Sprintf("time.CronJob.Stop - Job #%s %s No timer to Stop!!", cj._id, cj._name)
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
		return cj._id == job.Id()
	}
	return false
}

func (cj *cronJob) Id() string {
	return cj._id
}

func NewCronJob(label string, task func()(), time CronData, logger log.Logger) CronJob{
	return &cronJob{
		_id: GenerateCronId(),
		_name: label,
		_running: false,
		_data: time,
		_func: task,
		_counter: 0,
		_logger: logger,
	}
}

//func main() {
//	job := NewCronJob("Sample", func(){
//		fmt.Print("Toc..")
//	}, CronData{
//		Delay: 0 * time.Second,
//		Interval: 3 * time.Second,
//		NumExecutions: 4,
//	}, nil)
//	job.Start()
//	fmt.Println("\nStarted ...")
//	time.Sleep(20 * time.Second)
//	job.Stop()
//	fmt.Println("\nFinish!!")
//	time.Sleep(5 * time.Second)
//	fmt.Println("\nExit!!")
//
//}