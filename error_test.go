package gobinance

import "testing"

func TestHttpError_Error(t *testing.T) {
	uut := HttpError{
		HttpStatus: 400,
		errorDTO:   errorDTO{
			Code: -1234,
			Msg:  "test message",
		},
	}

	expected := "got status 400 from binance. error code was -1234: test message"
	if got := uut.Error(); got != expected {
		t.Errorf("unexpected error message. expected\n\t%v\ngot\n\t%v", expected, got)
	}
}

func TestHttpError_ErrorCode(t *testing.T) {
	uut := HttpError{
		HttpStatus: 400,
		errorDTO:   errorDTO{
			Code: -1234,
			Msg:  "test message",
		},
	}
	if got := uut.ErrorCode(); got != -1234 {
		t.Errorf("unexpected error code. expected %v but got %v", -1234, got)
	}
}

func TestHttpError_StatusCode(t *testing.T) {
	uut := HttpError{
		HttpStatus: 400,
		errorDTO:   errorDTO{
			Code: -1234,
			Msg:  "test message",
		},
	}
	if got := uut.StatusCode(); got != 400 {
		t.Errorf("unexpected status code. expected %v but got %v", 400, got)
	}
}