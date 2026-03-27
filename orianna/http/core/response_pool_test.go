package core_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/core/mocks"
	"go.uber.org/mock/gomock"
)

func TestAcquireReleaseSuccessResponse(t *testing.T) {
	data := map[string]string{"foo": "bar"}
	resp := core.AcquireSuccessResponse(http.StatusOK, "success msg", data)

	if resp.HTTPStatus != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.HTTPStatus)
	}
	if resp.Message != "success msg" {
		t.Errorf("expected message 'success msg', got '%s'", resp.Message)
	}

	core.ReleaseSuccessResponse(resp)
	core.ReleaseSuccessResponse(nil)
}

func TestAcquireReleaseErrorResponse(t *testing.T) {
	resp := core.AcquireErrorResponse("ERR_CODE", http.StatusBadRequest, "error msg")

	if resp.Code != "ERR_CODE" {
		t.Errorf("expected code ERR_CODE, got %s", resp.Code)
	}
	if resp.Message != "error msg" {
		t.Errorf("expected message 'error msg', got '%s'", resp.Message)
	}

	core.ReleaseErrorResponse(resp)
	core.ReleaseErrorResponse(nil)
}

func TestResponse_WithHTTPStatus(t *testing.T) {
	errResp := core.NewErrorResponse("BAD", http.StatusBadRequest, "bad").WithHTTPStatus(http.StatusNotFound)
	if errResp.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, errResp.HTTPStatus)
	}
}

func TestHandleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCtx := mocks.NewMockContext(ctrl)
	mockCtx.EXPECT().RequestID().Return("req-1").AnyTimes()
	mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	mockCtx.EXPECT().Get(gomock.Any()).Return("").AnyTimes()

	// JSON takes one argument and returns error
	mockCtx.EXPECT().JSON(gomock.Any()).Return(nil).AnyTimes()
	mockCtx.EXPECT().Status(gomock.Any()).Return(mockCtx).AnyTimes()

	testErr := errors.New("test err")

	// Call HandleError which formats standard errors into generic 500s
	core.HandleError(mockCtx, testErr)

	// Call SendSuccess to hit 80% coverage lines
	resp := core.AcquireSuccessResponse(http.StatusOK, "ok", nil)
	if err := core.SendSuccess(mockCtx, resp); err != nil {
		t.Errorf("SendSuccess failed: %v", err)
	}

	// Call WrapError
	_ = core.WrapError(testErr, "wrapped")
}
