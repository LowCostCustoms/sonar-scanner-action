package sonarscanner

import "fmt"

type AnalysisStatus int

const (
	AnalysisStatusUndefined AnalysisStatus = iota
	AnalysisStatusOk        AnalysisStatus = iota
	AnalysisStatusWarning   AnalysisStatus = iota
	AnalysisStatusError     AnalysisStatus = iota
	AnalysisStatusNone      AnalysisStatus = iota
)

const (
	analysisStatusUndefinedStr = "UNDEFINED"
	analysisStatusOkStr        = "OK"
	analysisStatusWarningStr   = "WARN"
	analysisStatusErrorStr     = "ERROR"
	analysisStatusNoneStr      = "NONE"
)

func (status AnalysisStatus) String() string {
	switch status {
	case AnalysisStatusOk:
		return analysisStatusOkStr
	case AnalysisStatusWarning:
		return analysisStatusWarningStr
	case AnalysisStatusError:
		return analysisStatusErrorStr
	case AnalysisStatusNone:
		return analysisStatusNoneStr
	default:
		return analysisStatusUndefinedStr
	}
}

func parseAnalysisStatus(value string) (AnalysisStatus, error) {
	switch value {
	case analysisStatusOkStr:
		return AnalysisStatusOk, nil
	case analysisStatusWarningStr:
		return AnalysisStatusWarning, nil
	case analysisStatusErrorStr:
		return AnalysisStatusError, nil
	case analysisStatusNoneStr:
		return AnalysisStatusNone, nil
	default:
		return AnalysisStatusUndefined, fmt.Errorf("unexpected analysis status '%s'", value)
	}
}
