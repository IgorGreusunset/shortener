package handlers

import (
	//"bytes"
	"context"
	"encoding/json"
	//"io"
	"net/http"
	"net/http/httptest"
	//"strings"
	"testing"

	model "github.com/IgorGreusunset/shortener/internal/app"
	"github.com/IgorGreusunset/shortener/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
)

func TestPostHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)
	
	m.EXPECT().Create(gomock.Any()).Return(nil)



	PostHandlerWrapper := func(res http.ResponseWriter, req *http.Request) {
		PostHandler(m, res, req)
	}

	handler := http.HandlerFunc(PostHandlerWrapper)

	srv := httptest.NewServer(handler)

	defer srv.Close()

	tests := []struct {
		name            string
		method          string
		reqBody         string
		expectedCode    int
		expectedContent string
	}{
		{
			name:            "normal_case",
			method:          http.MethodPost,
			reqBody:         "https://mail.ru/",
			expectedCode:    http.StatusCreated,
			expectedContent: "text/plain",
		},
		{
			name:            "not_url_case",
			method:          http.MethodPost,
			reqBody:         "some text not url",
			expectedCode:    http.StatusBadRequest,
			expectedContent: "",
		},
		{
			name:            "get_case",
			method:          http.MethodGet,
			reqBody:         "https://mail.ru/",
			expectedCode:    http.StatusBadRequest,
			expectedContent: "",
		},
	}
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = test.method
			req.URL = srv.URL
			req.Body = test.reqBody

			resp, err := req.Send()

			if err != nil {
				t.Errorf("error making HTTP request: %v", err)
			}

			if resp.StatusCode() != test.expectedCode {
				t.Errorf("Response code didn't match expected: got %d want %d", resp.StatusCode(), test.expectedCode)
			}

			if resp.Header().Get("Content-Type") != test.expectedContent {
				t.Errorf("Response content-type didn't match expected: got %v want %v", resp.Header().Get("Content-Type"), test.expectedContent)
			}
		})
	}
}

func TestGetByIDHandler(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)
	gomock.InOrder(
		m.EXPECT().GetByID("U8rtGB25").Return(model.URL{ID: "U8rtGB25", FullURL: "https://practicum.yandex.ru/"}, true),
		m.EXPECT().GetByID("g7RETf01").Return(model.URL{ID: "g7RETf01", FullURL: "https://mail.ru/"}, true),
		m.EXPECT().GetByID("yyokley").Return(model.URL{}, false),
	)
	GetHandlerWrapper := func(res http.ResponseWriter, req *http.Request) {
		GetByIDHandler(m, res, req)
	}

	handler := http.HandlerFunc(GetHandlerWrapper)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name             string
		method           string
		requestID        string
		expectedCode     int
		expectedLocation string
	}{
		{
			name:             "normal_practicum",
			method:           http.MethodGet,
			requestID:        "U8rtGB25",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:             "normal_mail",
			method:           http.MethodGet,
			requestID:        "g7RETf01",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://mail.ru/",
		},
		{
			name:         "id_not_in_storage",
			method:       http.MethodGet,
			requestID:    "yyokley",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			req := httptest.NewRequest(test.method, srv.URL+"/"+test.requestID, nil)

			cntx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, cntx))
			cntx.URLParams.Add("id", test.requestID)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetHandlerWrapper)
			h(w, req)

			res := w.Result()

			defer res.Body.Close()

			if res.StatusCode != test.expectedCode {
				t.Errorf("Response code didn't match expected: got %d want %d", res.StatusCode, test.expectedCode)
			}

			if res.Header.Get("Location") != test.expectedLocation {
				t.Errorf("Response Location didn't match expected: got %v want %v", res.Header.Get("Location"), test.expectedLocation)
			}
		})
	}
}

func TestAPIPostHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)

	m.EXPECT().Create(gomock.Any()).Return(nil)

	PostHandlerWrapper := func(res http.ResponseWriter, req *http.Request) {
		APIPostHandler(m, res, req)
	}

	handler := http.HandlerFunc(PostHandlerWrapper)

	srv := httptest.NewServer(handler)

	defer srv.Close()


	tests := []struct {
		name string
		method string
		reqBody model.APIPostRequest
		expectedCode int
		expectedContent string
	}{
		{
			name: "normal_case",
			method: http.MethodPost,
			reqBody: model.APIPostRequest{URL: "https://mail.ru/"},
			expectedCode: http.StatusCreated,
			expectedContent: "application/json",
		},
		{
			name: "not_url_case",
			method: http.MethodPost,
			reqBody: model.APIPostRequest{URL: "just text not url"},
			expectedCode: http.StatusBadRequest,
			expectedContent: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL
			req.Body, _ = json.Marshal(tt.reqBody)

			resp, err := req.Send()

			if err != nil {
				t.Errorf("error making HTTP request: %v", err)
			}

			if resp.StatusCode() != tt.expectedCode {
				t.Errorf("Response code didn't match expected: got %d want %d", resp.StatusCode(), tt.expectedCode)
			}

			if resp.Header().Get("Content-Type") != tt.expectedContent {
				t.Errorf("Response content-type didn't match expected: got %v want %v", resp.Header().Get("Content-Type"), tt.expectedContent)
			}
		})
	}
}

//Не смог одалеть написание тестов на этот хэндлер - падают тесты с ошибкой "Error during attemp to read response: http: read on closed response body"
/*func TestBatchHandler(t *testing.T)  {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)
	gomock.InOrder(
		m.EXPECT().Create(gomock.Any()).Return(nil).MaxTimes(2),
		m.EXPECT().CreateBatch(gomock.Any()).Return(nil).Times(1),
	)
	body := []model.APIBatchRequest{
		model.APIBatchRequest{ID: "1", URL: "https://mail.ru/"},
		model.APIBatchRequest{ID: "2", URL: "https://practicum.yandex.ru/"},
	}

	BatchHandlerWrapper :=  func (res http.ResponseWriter, req *http.Request){
		BathcHandler(m, res, req)
	}

	handler := http.HandlerFunc(BatchHandlerWrapper)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct{
		name string
		reqBody []model.APIBatchRequest
		expectedCode int
		expectedContent string
		expectedLen int
	}{
		{
			name: "nornal_case",
			reqBody: body,
			expectedCode: http.StatusCreated,
			expectedContent: "application/json",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL
			var b strings.Builder
			if err := json.NewEncoder(&b).Encode(tt.reqBody); err != nil {
				t.Errorf("Error during encode request body: %v", err)
			}

			req.Body = io.NopCloser(strings.NewReader(b.String()))

			resp, err := req.Send()

			if err != nil {
				t.Errorf("Error during http request: %v", err)
			}

			bodyBytes, err := io.ReadAll(resp.RawResponse.Body)
			if err != nil {
				t.Errorf("Error reading response body: %v", err)
				return	
			}

			defer resp.RawResponse.Body.Close()

			reader := bytes.NewReader(bodyBytes)

			if resp.StatusCode() != tt.expectedCode {
				t.Errorf("Response code didn't match expected: got %d want %d", resp.StatusCode(), tt.expectedCode)
			}

			if resp.Header().Get("Content-Type") != tt.expectedContent {
				t.Errorf("Response content-type didn't match expected: got %v want %v", resp.Header().Get("Content-Type"), tt.expectedContent)
			}

			var respBody []model.APIBatchResponse
			if err := json.NewDecoder(reader).Decode(&respBody); err != nil {
				t.Errorf("Error during attemp to read response: %s", err)
			} else if len(respBody) != tt.expectedLen {
				t.Errorf("Number of results in response didn't match expected: got %v want %v", len(respBody), tt.expectedLen)
			}
		})
	}

}*/