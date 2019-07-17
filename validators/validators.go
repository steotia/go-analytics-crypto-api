package validators

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang/glog"
)

const (
	timeFormat = "2006-01-02T15:04:05"
)

var (
	errInvalidFrom     = errors.New("Can't process 'from'. not in format: '2006-01-02T15:04:05'")
	errInvalidTo       = errors.New("Can't process 'to'. not in format: '2006-01-02T15:04:05'")
	errFormatNotJSON   = errors.New("Can't process: Unsupported format. Supported format(s): json")
	errFromLaterThenTo = errors.New("Can't process: 'from' is after 'to'")
)

type InputValidation interface {
	Validate(r *http.Request) error
}

type GetMetricsInput struct {
	From     string
	FromTime time.Time
	To       string
	ToTime   time.Time
	Format   string
}

func (i *GetMetricsInput) Validate() error {
	fromTime, err := CheckTimeFormat("from", i.From)
	i.FromTime = fromTime
	if err != nil {
		return err
	}
	toTime, err := CheckTimeFormat("to", i.To)
	i.ToTime = toTime
	if err != nil {
		return err
	}
	if i.Format != "json" {
		err := errors.New("Can't process: Unsupported format. Supported format(s): json")
		glog.Error(err)
		return err
	}
	if !i.ToTime.After(i.FromTime) {
		err := errors.New("Can't process: 'to' is not after 'from'")
		glog.Error(err)
		return err
	}
	return nil
}

func NewGetMetricsInput(r *http.Request) (GetMetricsInput, error) {
	getMetricsInput := GetMetricsInput{}

	from, err := CheckParam(r, "from")
	if err != nil {
		return GetMetricsInput{}, err
	}
	getMetricsInput.From = from
	to, err := CheckParam(r, "to")
	if err != nil {
		return GetMetricsInput{}, err
	}
	getMetricsInput.To = to
	format, err := CheckParam(r, "format")
	if err != nil {
		return GetMetricsInput{}, err
	}
	getMetricsInput.Format = format
	return getMetricsInput, nil
}
