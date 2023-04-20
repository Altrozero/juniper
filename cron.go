package juniper

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/1set/cronrange"
	"github.com/go-playground/validator"
	"gopkg.in/yaml.v3"
)

const (
	cronLopEnv         = "CRON_LOG"
	defaultCronLogPath = "./configs/cron.yml"
)

type CronTask struct {
	Command  string   ` yaml:"command" validate:"required"`
	Schedule string   ` yaml:"schedule" validate:"required"`
	Args     []string `yaml:"args"`
}

type cronTasks struct {
	Tasks []CronTask `validate:"dive"`
}

// shouldRun checks if the given time matches the schedule set on the task
func (task CronTask) ShouldRun(now time.Time) bool {
	// tz, _ := now.Zone()
	schedule, err := cronrange.ParseString(fmt.Sprintf("DR=1; TZ=UTC; %s", task.Schedule))

	if err != nil {
		return false
	}
	
	now = now.Truncate(time.Second)

	return schedule.IsWithin(now)
}

// ParseCronSchedule file by its file path
func ParseCronSchedule(configPath string) ([]CronTask, error) {
	var tasks []CronTask

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	if err := validateCronSchedule(tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func validateCronSchedule(tasks []CronTask) error {
	v := validator.New()
	return v.Struct(cronTasks{Tasks: tasks})
}

var _ error = (*CronErrors)(nil)

type CronErrors map[string]error

// Error implementation keeps CronErrors adheering to the error interface
func (ce CronErrors) Error() string {
	errString := strings.Builder{}

	for key, err := range ce {
		errString.WriteString(fmt.Sprintf("%s: %s", key, err))
	}

	return errString.String()
}

// RunCranTasks that have hit their trigger
func RunCronTasks(tasks []CronTask, register CliCommandEntries) CronErrors {
	var (
		wg      sync.WaitGroup
		errors  = make(CronErrors)
		now     = time.Now()
		cronLog *os.File
		mux     sync.Mutex
		logPath string
		ok      bool
	)

	if logPath, ok = os.LookupEnv(cronLopEnv); !ok {
		logPath = defaultCronLogPath
	}

	cronLog, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	for _, task := range tasks {
		wg.Add(1)

		go func(task CronTask) {
			defer wg.Done()

			if !task.ShouldRun(now) {
				return
			}

			entry := register.Find(task.Command)
			if entry == nil {
				log.Printf("Command '%s' not found", task.Command)
				return
			}

			err := entry.Run(task.Args)
			if err != nil {
				errors[task.Command] = err
			}

			if cronLog != nil {
				mux.Lock()
				cronLog.WriteString(fmt.Sprintf("[%s] %s:\n%v\n\n", time.Now(), task.Command, err))
				mux.Unlock()
			}
		}(task)
	}

	wg.Wait()

	if len(errors) > 0 {
		return errors
	}

	return nil
}
