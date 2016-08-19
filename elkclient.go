package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ELKClient implements the elasticsearch REST API
type ELKClient struct {
	Index  string
	Age    int
	Client *http.Client
}

// ELKResult instances contain information about the result of a
// single operation on the elasticsearch server.
type ELKResult struct {
	Msg string `json:"data,omitempty"`
	Err error  `json:"error,omitempty"`
}

// GetIndexes retrieves the list of all indexes from the elasticsearch server.
// It returns a channel of ELKResults containing the index listings.
// If any problems are encountered, the ELKResult will contain an error.
func (ec *ELKClient) GetIndexes() <-chan ELKResult {
	c := make(chan ELKResult)

	go func() {
		defer close(c)
		url := "http://localhost:9200/_cat/indices?v"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c <- ELKResult{Err: err}
			return
		}
		res, err := ec.Client.Do(req)
		if err != nil {
			c <- ELKResult{Err: err}
			return
		}

		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c <- ELKResult{Err: err}
			return
		}

		if res.StatusCode != http.StatusOK {
			err := ELKResult{
				Err: errors.New(fmt.Sprintf("%d: %s",
					res.StatusCode, string(data))),
			}
			c <- err
			return
		}

		old := time.Now().Add(time.Hour * time.Duration(-24*ec.Age))
		old = time.Date(old.Year(), old.Month(), old.Day(),
			0, 0, 0, 0, time.Local)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if pos := strings.Index(line, ec.Index); pos > 0 {
				if len(line) > pos+len(ec.Index)+11 {
					index := line[pos : pos+len(ec.Index)+11]
					tmpstr := line[pos+len(ec.Index)+1 : pos+len(ec.Index)+11]
					it, err := time.ParseInLocation("2006.01.02", tmpstr,
						time.Local)
					if err != nil {
						c <- ELKResult{Err: err}
						continue
					}

					if it.Unix() < old.Unix() {
						c <- ELKResult{Msg: index}
					}
				}
			}
		}
	}()

	return c
}

// DeleteIndex deletes the specified index from the elasticsearch server.
// It returns an ELKResult value containing the name of the deleted index.
// If any problems are encountered, the ELKResult will contain an error.
func (ec *ELKClient) DeleteIndex(index string) ELKResult {
	url := "http://localhost:9200/" + index
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return ELKResult{Err: err}
	}
	res, err := ec.Client.Do(req)
	if err != nil {
		return ELKResult{Err: err}
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return ELKResult{Err: err}
		}

		return ELKResult{
			Err: errors.New(fmt.Sprintf("%d: %s",
				res.StatusCode, string(data))),
		}
	}

	return ELKResult{Msg: index}
}
