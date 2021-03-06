/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package crud

import (
	"connectordb/authoperator"
	"connectordb/datastream"
	"errors"
	"fmt"
	"net/http"
	"server/restapi/restcore"
	"server/webcore"
	"strconv"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrRangeArgs is thrown when invalid arguments are given to trange
	ErrRangeArgs = errors.New(`A range needs [both "i1" and "i2" int] or ["t1" and ["t2" decimal and/or "limit" int]]`)
	//ErrTime2IndexArgs is the error when args are incorrectly given to t2i
	ErrTime2IndexArgs = errors.New(`time2index requires an argument of "t" which is a decimal timestamp`)
)

//StreamLength gets the stream length
func StreamLength(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	l, err := o.LengthStream(streampath)

	return restcore.IntWriter(writer, l, logger, err)
}

//WriteStream writes the given stream
func WriteStream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	var datapoints []datastream.Datapoint
	err := restcore.UnmarshalRequest(request, &datapoints)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	restamp := request.Method == "PUT"

	querylog := fmt.Sprintf("Insert %d", len(datapoints))
	if restamp {
		querylog += " (restamp)"
	}

	err = o.InsertStream(streampath, datapoints, restamp)
	if err != nil {
		lvl, _ := restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return lvl, querylog
	}
	atomic.AddUint32(&webcore.StatsInserts, uint32(len(datapoints)))
	restcore.OK(writer)
	return webcore.DEBUG, querylog
}

//StreamRange gets a range of data from a stream
func StreamRange(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)
	q := request.URL.Query()
	transform := q.Get("transform")

	i1, i2, err := restcore.ParseIRange(q)
	if err == nil {
		querylog := fmt.Sprintf("irange [%d,%d)", i1, i2)
		dr, err := o.GetStreamIndexRange(streampath, i1, i2, transform)
		lvl, _ := restcore.WriteJSONResult(writer, dr, logger, err)
		return lvl, querylog
	} else if err != restcore.ErrCantParse {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	//The error is ErrCantParse - meaning that i1 and i2 are not present in query

	t1, t2, lim, err := restcore.ParseTRange(q)
	if err == nil {
		querylog := fmt.Sprintf("trange [%.1f,%.1f) limit=%d", t1, t2, lim)
		dr, err := o.GetStreamTimeRange(streampath, t1, t2, lim, transform)
		lvl, _ := restcore.WriteJSONResult(writer, dr, logger, err)
		return lvl, querylog
	}

	//None of the limits were recognized. Rather than exploding, return bad request
	return restcore.WriteError(writer, logger, http.StatusBadRequest, ErrRangeArgs, false)
}

//StreamTime2Index gets the time associated with the index
func StreamTime2Index(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)
	logger = logger.WithField("op", "Time2Index")

	ts := request.URL.Query().Get("t")
	t, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, ErrTime2IndexArgs, false)
	}
	logger.Debugln("t=", ts)

	i, err := o.TimeToIndexStream(streampath, t)

	lvl, _ := restcore.JSONWriter(writer, i, logger, err)
	return lvl, "t=" + ts
}
