package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"os"
	"bytes"
	"mime/multipart"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

var testImageData []byte

func TestMain(m *testing.M) {
	// 画像データのモックをセットアップ
	var err error
	testImageData, err = os.ReadFile("../images/default.jpg")
	if err != nil {
		testImageData = []byte("test image data")
	}

	os.Exit(m.Run())
}


func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	// ✅ 画像データのモックを用意
	testImageData, err := os.ReadFile("../images/default.jpg")
	if err != nil {
		testImageData = []byte("test image data")
	}

	// STEP 6-1: define test cases
	cases := map[string]struct {
		args 	  map[string]string
		imageData []byte
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "jacket", // fill here
				"category": "fashion", // fill here
			},
			imageData: testImageData,
			wants: wants{
				req: &AddItemRequest{
					Name: "jacket", // fill here
					Category: "fashion", // fill here
					Image: testImageData,
				},
				err: false,
			},
		},
		"ng: missing name": {
			args: map[string]string{
				"category": "fashion",
			},
			imageData: testImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing category": {
			args: map[string]string{
				"name": "jacket",
			},
			imageData: testImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
	
			req, err := createMultipartRequest("http://localhost:9000/items", tt.args, tt.imageData)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
	
			// execute test target
			got, err := parseAddItemRequest(req)
	
			// confirm the result
			if err != nil {
				if !tt.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if diff := cmp.Diff(tt.wants.req, got); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}
		})
	}
}

func createMultipartRequest(url string, params map[string]string, imageData []byte) (*http.Request, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 文字データの追加
    for key, value := range params {
        _ = writer.WriteField(key, value)
    }

    // 画像データの追加
    filePart, err := writer.CreateFormFile("image", "default.jpg")
    if err != nil {
        return nil, err
    }
    _, err = filePart.Write(imageData)
    if err != nil {
        return nil, err
    }

    writer.Close()

    req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType()) // Content-Type を設定

    return req, nil
}


func parseAddItemRequest(req *http.Request) (*AddItemRequest, error) {
    err := req.ParseMultipartForm(10 << 20) // 10MB
    if err != nil {
        return nil, fmt.Errorf("failed to parse multipart form: %w", err)
    }

    name := req.FormValue("name")
    category := req.FormValue("category")
    file, _, err := req.FormFile("image")
    if err != nil {
        return nil, fmt.Errorf("image is required")
    }
    defer file.Close()

    imageData, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read image data: %w", err)
    }

    return &AddItemRequest{
        Name:     name,
        Category: category,
        Image:    imageData,
    }, nil
}


func TestHelloHandler(t *testing.T) {
	t.Parallel()

	// Please comment out for STEP 6-2
	// predefine what we want
	// type wants struct {
	// 	code int               // desired HTTP status code
	// 	body map[string]string // desired body
	// }
	// want := wants{
	// 	code: http.StatusOK,
	// 	body: map[string]string{"message": "Hello, world!"},
	// }

	// set up test
	req := httptest.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()

	h := &Handlers{}
	h.Hello(res, req)

	// STEP 6-2: confirm the status code

	// STEP 6-2: confirm response body
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	type wants struct {
		code int
	}
	cases := map[string]struct {
		args     map[string]string
		injector func(m *MockItemRepository)
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				// STEP 6-3: define mock expectation
				// succeeded to insert
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusOK,
			},
		},
		"ng: failed to name": {
			args: map[string]string{
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				// STEP 6-3: define mock expectation
				// failed to insert
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
		"ng: failed to category": {
			args: map[string]string{
				"name": "used iPhone 16e",
			},
			injector: func(m *MockItemRepository) {
				// STEP 6-3: define mock expectation
				// failed to insert
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				// ✅ Insert が失敗するモックを設定
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errItemNotFound)
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIR := NewMockItemRepository(ctrl)
			tt.injector(mockIR)
			h := &Handlers{itemRepo: mockIR}

			// テストケース内で使用する
			req, err := createMultipartRequest("/items", tt.args, testImageData)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}
			req, err = createMultipartRequest("/items", tt.args, testImageData)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}

			for _, v := range tt.args {
				if !strings.Contains(rr.Body.String(), v) {
					t.Errorf("response body does not contain %s, got: %s", v, rr.Body.String())
				}
			}
		})
	}
}

// STEP 6-4: uncomment this test
// func TestAddItemE2e(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping e2e test")
// 	}

// 	db, closers, err := setupDB(t)
// 	if err != nil {
// 		t.Fatalf("failed to set up database: %v", err)
// 	}
// 	t.Cleanup(func() {
// 		for _, c := range closers {
// 			c()
// 		}
// 	})

// 	type wants struct {
// 		code int
// 	}
// 	cases := map[string]struct {
// 		args map[string]string
// 		wants
// 	}{
// 		"ok: correctly inserted": {
// 			args: map[string]string{
// 				"name":     "used iPhone 16e",
// 				"category": "phone",
// 			},
// 			wants: wants{
// 				code: http.StatusOK,
// 			},
// 		},
// 		"ng: failed to insert": {
// 			args: map[string]string{
// 				"name":     "",
// 				"category": "phone",
// 			},
// 			wants: wants{
// 				code: http.StatusBadRequest,
// 			},
// 		},
// 	}

// 	for name, tt := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			h := &Handlers{itemRepo: &itemRepository{db: db}}

// 			values := url.Values{}
// 			for k, v := range tt.args {
// 				values.Set(k, v)
// 			}
// 			req := httptest.NewRequest("POST", "/items", strings.NewReader(values.Encode()))
// 			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 			rr := httptest.NewRecorder()
// 			h.AddItem(rr, req)

// 			// check response
// 			if tt.wants.code != rr.Code {
// 				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
// 			}
// 			if tt.wants.code >= 400 {
// 				return
// 			}
// 			for _, v := range tt.args {
// 				if !strings.Contains(rr.Body.String(), v) {
// 					t.Errorf("response body does not contain %s, got: %s", v, rr.Body.String())
// 				}
// 			}

// 			// STEP 6-4: check inserted data
// 		})
// 	}
// }

// func setupDB(t *testing.T) (db *sql.DB, closers []func(), e error) {
// 	t.Helper()

// 	defer func() {
// 		if e != nil {
// 			for _, c := range closers {
// 				c()
// 			}
// 		}
// 	}()

// 	// create a temporary file for e2e testing
// 	f, err := os.CreateTemp(".", "*.sqlite3")
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	closers = append(closers, func() {
// 		f.Close()
// 		os.Remove(f.Name())
// 	})

// 	// set up tables
// 	db, err = sql.Open("sqlite3", f.Name())
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	closers = append(closers, func() {
// 		db.Close()
// 	})

// 	// TODO: replace it with real SQL statements.
// 	cmd := `CREATE TABLE IF NOT EXISTS items (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		name VARCHAR(255),
// 		category VARCHAR(255)
// 	)`
// 	_, err = db.Exec(cmd)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	return db, closers, nil
// }
