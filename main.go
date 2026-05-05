package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"

	d "github.com/jacksonzamorano/strata-search/definitions"
	"github.com/jacksonzamorano/strata/component"
)

type ripgrepMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
type ripgrepText struct {
	Text string `json:"text"`
}
type ripgrepMatchMessage struct {
	Path       ripgrepText `json:"path"`
	Lines      ripgrepText `json:"lines"`
	LineNumber int         `json:"line_number"`
}
type ripgrepSummaryMessage struct {
	ElapsedTotal ripgrepElapsedTotal `json:"elapsed_total"`
}
type ripgrepElapsedTotal struct {
	Nanos int `json:"nanos"`
}

func search(
	input d.SearchInput,
	ctx *component.ComponentContainer,
) (*d.SearchOutput, error) {
	cString := strconv.Itoa(input.Context)

	res := ctx.RunInDirectory(input.Path, "rg", "-i", "-C", cString, "--json", input.Pattern)
	if !res.Ok {
		return nil, fmt.Errorf("Search could not be completed: %s.", res.Error)
	}

	out := d.SearchOutput{
		Matches: []d.SearchMatch{},
	}
	var tempMatch d.SearchMatch

	for l := range bytes.Lines([]byte(res.Output)) {
		var msg ripgrepMessage
		err := json.Unmarshal(l, &msg)
		if err != nil {
			ctx.Logger.Log("Received invalid message from ripgrep: %s", err.Error())
			continue
		}
		switch msg.Type {
		case "context":
			var matchData ripgrepMatchMessage
			err := json.Unmarshal(msg.Data, &matchData)
			if err != nil {
				continue
			}
			if len(tempMatch.Match) == 0 {
				// Before context
				tempMatch.Before += matchData.Lines.Text
			} else {
				// After context
				tempMatch.After += matchData.Lines.Text
			}
		case "match":
			if len(tempMatch.Match) > 0 {
				out.Matches = append(out.Matches, tempMatch)
				tempMatch = d.SearchMatch{}
			}
			var matchData ripgrepMatchMessage
			err := json.Unmarshal(msg.Data, &matchData)
			if err != nil {
				continue
			}
			tempMatch.File = matchData.Path.Text
			tempMatch.AbsolutePath = path.Join(input.Path, matchData.Path.Text)
			tempMatch.LineNumber = matchData.LineNumber
			tempMatch.Match = matchData.Lines.Text
		case "separator", "end":
			if len(tempMatch.Match) > 0 {
				out.Matches = append(out.Matches, tempMatch)
			}
			tempMatch = d.SearchMatch{}
		case "summary":
			var summaryData ripgrepSummaryMessage
			err := json.Unmarshal(msg.Data, &summaryData)
			if err != nil {
				continue
			}
			out.TimeSeconds = (time.Nanosecond * time.Duration(summaryData.ElapsedTotal.Nanos)).Seconds()
		}
	}

	return &out, nil
}

func main() {
	component.CreateComponent(
		d.Manifest,
		component.Mount(d.Search, search),
	).Start()
}
