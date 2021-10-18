package progress_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/dude333/rapinav2/pkg/progress"
)

func TestProgress(t *testing.T) {

	progress.Cursor(false)
	defer progress.Cursor(true)
	progress.Status("a status msg")

	progress.Error(bad())
	progress.ErrorStack(bad())
	progress.ErrorStack(fmt.Errorf("stackless error"))
	progress.ErrorStack(badInt())

	progress.Running("start process")
	progress.Error(errors.New("some error"))
	time.Sleep(time.Second)
	progress.RunOK()

	progress.Running("start another process")
	time.Sleep(time.Second)
	progress.Status("middle")
	time.Sleep(time.Second)
	progress.RunFail()

	f1()

	progress.Running("start spinner")
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		progress.Spinner()
		if i == 20 {
			progress.Status("spinner interrupt")
		}
	}
	progress.RunOK()

	progress.Status("end.")
}

func f1() {
	progress.Running("Running *f1*")
	time.Sleep(time.Second)
	progress.Warning("f1 warning")
	time.Sleep(time.Second)
	progress.RunOK()
}

func bad() error {
	return errors.New("a bad error")
}

func badInt() error {
	_, err := strconv.Atoi("nonnumber")
	return errors.WithStack(err)
}
