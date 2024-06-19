package rpcserver

import (
	"github.com/latifrons/latigo/berror"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

const OrderDirectionASC = "ASC"
const OrderDirectionDESC = "DESC"

var (
	ErrBadRequest = "ErrBadRequest"
	ErrInternal   = "ErrInternal"
	ErrNotFound   = "ErrNotFound"
)

type PagingParams struct {
	Offset    int  `json:"offset"`     // Default to 0
	Limit     int  `json:"limit"`      // Default to 10
	NeedTotal bool `json:"need_total"` // Default to false
}

func (p PagingParams) ToPageNumSize() (pageNum int, pageSize int) {
	pageNum = int(p.Offset/p.Limit) + 1
	pageSize = p.Limit
	return
}

type PagingResult struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

type OrderParams struct {
	OrderBy        string `json:"order_by"`        // DB column name. Empty string means no sorting.
	OrderDirection string `json:"order_direction"` // ASC or DESC
}

func (o *OrderParams) ToSqlOrderBy() (c interface{}, err error) {
	if o.OrderBy == "" || o.OrderDirection == "" {
		c = ""
		return
	}
	dir := strings.ToUpper(o.OrderDirection)
	if dir != OrderDirectionASC && dir != OrderDirectionDESC {
		err = berror.NewBusinessFail(nil, ErrBadRequest, "bad order params")
		return
	}
	return clause.OrderByColumn{Column: clause.Column{Name: o.OrderBy}, Desc: dir == OrderDirectionDESC}, nil
}

type TimeRangeParams struct {
	StartTime *time.Time
	EndTime   *time.Time
}

func NewTimeRange(startTimestamp *int64, endTimestamp *int64) (v TimeRangeParams) {

	if startTimestamp != nil {
		secs := *startTimestamp  // Get seconds
		tp := time.Unix(secs, 0) // Convert to time.Time struct
		v.StartTime = &tp
	}
	if endTimestamp != nil {
		secs := *endTimestamp    // Get seconds
		tp := time.Unix(secs, 0) // Convert to time.Time struct
		v.EndTime = &tp
	}
	return v
}
