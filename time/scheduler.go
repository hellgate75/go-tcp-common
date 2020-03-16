package time

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/hellgate75/go-tcp-common/log"
	uuid2 "github.com/satori/go.uuid"
	"math"
	"strings"
	"time"
)

var (
	CronUuidReparatorLeg int = 8
	CronUuidStandardLength int = 40
)

type CronUUID string
//
func salt(t time.Time) int64 {
	time.Sleep(100 * time.Millisecond)
	return 7 * durationToNumber(time.Now().Sub(t), time.Nanosecond)
}

// Generate a Cron eleemnt unique id of given length
func GenerateCronId() CronUUID {
	var floatNum float64  = (float64(CronUuidStandardLength)/float64(CronUuidReparatorLeg))
	var tokens int = int(math.Floor(floatNum))
	if CronUuidStandardLength < 10 {
		tokens = 0
	}
	b := make([]byte, CronUuidStandardLength-tokens)
	if _, err := rand.Read(b); err != nil {
		uuid, err :=  uuid2.NewV4()
		if err != nil {
			return CronUUID(fmt.Sprintf("cron-%v-%v-cron", time.Now().Unix(), salt(time.Now())))
		}
		return CronUUID(uuid.String())
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
		return CronUUID(uuid + "-cron")
	}
	return CronUUID("cron-" + hex.EncodeToString(b) + "-cron")
}


type CronData struct {
	Delay			time.Duration
	Interval		time.Duration
	NumExecutions	int64
}

type CronJob interface {
	Label() string
	Start() error
	Stop() error
	IsRunning() bool
	Update(data CronData) error
	LastRun() time.Time
	Equals(job CronJob) bool
	Id() CronUUID
	String() string
	Report() string
}

type CronTab interface {
	AddJob(job CronJob) error
	RemoveJob(uuid string) error
	ListJobs() []CronJob
	Start() error
	Stop() error
	IsRunning() error
	StartAllJobs()
	KillAllJobs()
	Equals(job CronTab) bool
	Id() CronUUID
	String() string
	Report() string
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