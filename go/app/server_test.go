package app_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"mercari-build-training/app"
	
)

var dummyImageData []byte
func init() {
	var err error
	dummyImageData, err = os.ReadFile("/home/choco/mercari-build-training/go/images/default.jpg")
	if err != nil {
		// 失敗時はデフォルトのダミーデータを使用
		dummyImageData = []byte{0xFF, 0xD8, 0xFF, 0xE0}
	}
}

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *app.AddItemRequest
		err bool
	}

	// 画像データを実際のファイルから読み込む
	dummyImageData, err := os.ReadFile("/home/choco/mercari-build-training/go/images/default.jpg")
	if err != nil {
		t.Fatalf("failed to read image file: %v", err)
	}

	emptyImageData := []byte{}

	// 256文字の文字列を作成
	longString := strings.Repeat("a", 256)

	cases := map[string]struct {
		args      map[string]string
		imageData []byte // 画像データを追加
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "jacket",
				"category": "fashion",
				"image":    "default.jpg",
			},
			imageData: dummyImageData,
			wants: wants{
				req: &app.AddItemRequest{
					Name:     "jacket",
					Category: "fashion",
					Image:    dummyImageData,
				},
				err: false,
			},
		},
		"ng: empty request": {
			args:      map[string]string{},
			imageData: nil,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing name": {
			args: map[string]string{
				"category": "fashion",
				"image":    "default.jpg",
			},
			imageData: dummyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing category": {
			args: map[string]string{
				"name":  "jacket",
				"image": "default.jpg",
			},
			imageData: dummyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing image": {
			args: map[string]string{
				"name":     "jacket",
				"category": "fashion",
			},
			imageData: nil,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: empty image file": {
			args: map[string]string{
				"name":     "jacket",
				"category": "fashion",
				"image":    "default.jpg",
			},
			imageData: emptyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: invalid image format": {
			args: map[string]string{
				"name":     "jacket",
				"category": "fashion",
				"image":    "default.jpg",
			},
			imageData: []byte("this is not an image"), // テキストデータを画像として送信
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: too long name": {
			args: map[string]string{
				"name":     longString,
				"category": "fashion",
				"image":    "default.jpg",
			},
			imageData: dummyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: too long category": {
			args: map[string]string{
				"name":     "jacket",
				"category": longString,
				"image":    "default.jpg",
			},
			imageData: dummyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// `multipart/form-data` のリクエストを作成
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// フォームデータを追加
			for key, value := range tt.args {
				_ = writer.WriteField(key, value)
			}

			// 画像を送信する場合
			if _, exists := tt.args["image"]; exists {
				part, _ := writer.CreateFormFile("image", "default.jpg")
				part.Write(tt.imageData)
			}

			writer.Close()

			// HTTPリクエストを作成
			req, err := http.NewRequest("POST", "http://localhost:9000/items", body)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// execute test target
			r := httptest.NewRequest("POST", "/items", nil) // `r` を明示的に定義
			got, err := app.ParseAddItemRequest(r)

			// confirm the result
			if err != nil {
				if !tt.wants.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wants.err && err == nil {
				t.Errorf("expected error but got nil")
				return
			}

			// ここに挿入する
			// var got *app.AddItemRequest
			got, err = app.ParseAddItemRequest(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != "jacket" {
				t.Errorf("expected 'jacket', got '%s'", got.Name)
			}

			// `Image` フィールドの比較を無視して `cmp.Diff` で比較
			if diff := cmp.Diff(tt.wants.req, got, cmpopts.IgnoreFields(&app.AddItemRequest{}, "Image")); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}

			// 画像データのサイズをチェック
			if len(got.Image) != len(tt.wants.req.Image) {
				t.Errorf("image size mismatch: got %d, want %d", len(got.Image), len(tt.wants.req.Image))
			}
		})
	}
}

func TestHelloHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	h := &app.Handlers{}
	h.Hello(res, req)

	want := http.StatusOK
	if res.Code != want {
		t.Errorf("expected status code %d, got %d", want, res.Code)
	}

	var resp app.HelloResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Message != "Hello, world!" {
		t.Errorf("expected message %q, got %q", "Hello, world!", resp.Message)
	}
}


func TestAddItemHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := app.NewMockItemRepository(ctrl)
	h := &app.Handlers{ItemRepo: mockRepo, ImgDirPath: "default.ipg"}
	// テスト用のリクエストとレスポンス
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	// Helloハンドラを実行
	h.Hello(res, req)

	tempImage := []byte{0xFF, 0xD8, 0xFF, 0xE0} // Dummy JPEG header

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "Test Item")
	_ = writer.WriteField("category", "Test Category")

	part, _ := writer.CreateFormFile("image", "test.jpg")
	_, _ = part.Write(tempImage)
	writer.Close()

	type wants struct {
		code int
		body string
	}

	cases := map[string]struct {
		args       map[string]string
		imageData  []byte
		setupMocks func(m *app.MockItemRepository)
		wants
	}{
		"ok: correctly insert item with new category": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// succeeded to insert with new category
				// カテゴリが存在しない場合
				m.EXPECT().app.GetCategoryByName(gomock.Any(), "phone").Return(nil, errors.New("category not found"))
				// 新しいカテゴリを作成
				m.EXPECT().InsertCategory(gomock.Any(), "phone").Return(&app.Category{ID: 1, Name: "phone"}, nil)
				// アイテム挿入成功 - 任意の引数を許容
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusOK,
				body: "item received: used iPhone 16",
			},
		},
		"ok: correctly insert item with existing category": {
			args: map[string]string{
				"name":     "MacBook Pro",
				"category": "laptop",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// succeeded to insert with existing category
				// カテゴリが既に存在する場合
				m.EXPECT().app.GetCategoryByName(gomock.Any(), "laptop").Return(&app.Category{ID: 2, Name: "laptop"}, nil)
				// アイテム挿入成功 - 任意の引数を許容
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusOK,
				body: "item received: MacBook Pro",
			},
		},
		"ng: failed to get category": {
			args: map[string]string{
				"name":     "iPad",
				"category": "tablet",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// カテゴリ取得失敗（データベースエラー）
				m.EXPECT().app.GetCategoryByName(gomock.Any(), "tablet").Return(nil, errors.New("database error"))
				// 新しいカテゴリを作成しようとして失敗
				m.EXPECT().InsertCategory(gomock.Any(), "tablet").Return(nil, errors.New("failed to create category"))
				// この場合はInsertは呼ばれないので期待しない
			},
			wants: wants{
				code: http.StatusInternalServerError,
				body: "Failed to create category",
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// カテゴリは正常に取得
				m.EXPECT().app.GetCategoryByName(gomock.Any(), "phone").Return(&app.Category{ID: 3, Name: "phone"}, nil)
				// アイテム挿入失敗
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("failed to insert item"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
				body: "failed to insert item",
			},
		},
		"ng: invalid request (missing name)": {
			args: map[string]string{
				"category": "phone",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// モックの呼び出しは期待されない
			},
			wants: wants{
				code: http.StatusBadRequest,
				body: "name is required",
			},
		},
		"ng: invalid request (missing category)": {
			args: map[string]string{
				"name": "iPhone",
			},
			imageData: dummyImageData,
			setupMocks: func(m *app.MockItemRepository) {
				// モックの呼び出しは期待されない
			},
			wants: wants{
				code: http.StatusBadRequest,
				body: "category is required",
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// テスト用のファイルシステムを作成
			tempDir, err := os.MkdirTemp("", "test-images-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// tempDirが実際に存在することを確認
			if _, err := os.Stat(tempDir); os.IsNotExist(err) {
				t.Fatalf("temp directory does not exist: %v", err)
			}

			// Gomockコントローラーの設定
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// モックリポジトリの作成と設定
			mockRepo := app.NewMockItemRepository(ctrl)
			tt.setupMocks(mockRepo)

			// テスト対象のハンドラーを作成
			h := &app.Handlers{
				ImgDirPath: tempDir,
				ItemRepo:   mockRepo,
			}

			// multipart/form-dataリクエストの作成
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// フォームフィールドの追加
			for key, value := range tt.args {
				err := writer.WriteField(key, value)
				if err != nil {
					t.Fatalf("failed to write field: %v", err)
				}
			}

			// 画像ファイルの追加
			if tt.imageData != nil {
				part, err := writer.CreateFormFile("image", "default.jpg")
				if err != nil {
					t.Fatalf("failed to create form file: %v", err)
				}
				_, err = part.Write(tt.imageData)
				if err != nil {
					t.Fatalf("failed to write image data: %v", err)
				}
			}

			err = writer.Close()
			if err != nil {
				t.Fatalf("failed to close writer: %v", err)
			}

			// HTTPリクエストとレスポンスレコーダーの作成
			req := httptest.NewRequest("POST", "/items", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			res := httptest.NewRecorder()

			mockRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			// テスト対象のハンドラーを実行
			h.AddItem(res, req)

			// 結果の検証
			if res.Code != tt.wants.code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, res.Code)
			}

			if tt.wants.body != "" && !strings.Contains(res.Body.String(), tt.wants.body) {
				t.Errorf("expected response body to contain %q, got %q", tt.wants.body, res.Body.String())
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
// 			h := &app.Handlers{ItemRepo: &itemRepository{db: db}}

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