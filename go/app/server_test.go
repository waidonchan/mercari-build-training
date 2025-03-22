package app

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	// STEP 6-1: define test cases
	// 画像データを実際のファイルから読み込む  \\wsl.localhost\Ubuntu\home\choco\mercari-build-training\go\images
	dummyImageData, err := loadTestImage()
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	emptyImageData := []byte{}

	// 256文字の文字列を作成
	longString := strings.Repeat("a", 256)

	cases := map[string]struct {
		args      map[string]string
		imageData []byte
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "jacket",  // fill here
				"category": "fashion", // fill here
			},
			imageData: dummyImageData,
			wants: wants{
				req: &AddItemRequest{
					Name:     "jacket",  // fill here
					Category: "fashion", // fill here
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
			},
			imageData: dummyImageData,
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing category": {
			args: map[string]string{
				"name": "jacket",
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

			// prepare request body
			// `multipart/form-data` のリクエストを作成
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// フォームデータを追加(画像以外)
			for key, value := range tt.args {
				_ = writer.WriteField(key, value)
			}

			// prepare HTTP request
			// 画像データを送信する場合
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

			writer.Close()

			// HTTPリクエストを作成
			req, err := http.NewRequest("POST", "http://localhost:9000/items", body)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// execute test target
			got, err := parseAddItemRequest(req)

			// confirm the result
			if err != nil {
				if !tt.wants.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wants.err {
				t.Errorf("expected error but got nil")
				return
			}

			// `Image` フィールドの比較を無視して `cmp.Diff` で比較
			if diff := cmp.Diff(tt.wants.req, got, cmpopts.IgnoreFields(AddItemRequest{}, "Image")); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}

			// 画像データの内容を厳密に比較
			if !bytes.Equal(got.Image, tt.wants.req.Image) {
				t.Errorf("image data mismatch")
			}

			// MIMEタイプのバリデーション（JPEG/PNG であることを確認）
			contentType := http.DetectContentType(got.Image[:512])
			validMimeTypes := map[string]bool{"image/jpeg": true, "image/png": true}
			if _, valid := validMimeTypes[contentType]; !valid {
				t.Errorf("invalid image format: got %s, expected JPEG or PNG", contentType)
			}
		})
	}
}

func TestHelloHandler(t *testing.T) {
	t.Parallel()

	// Please comment out for STEP 6-2
	// predefine what we want
	type wants struct {
		code int               // desired HTTP status code
		body map[string]string // desired body
	}
	want := wants{
		code: http.StatusOK,
		body: map[string]string{"message": "Hello, world!"},
	}

	// set up test
	req := httptest.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()

	h := &Handlers{}
	h.Hello(res, req)

	// STEP 6-2: confirm the status code
	if res.Code != want.code {
		t.Errorf("expected status code %d, got %d", want.code, res.Code)
	}

	// STEP 6-2: confirm response body
	for _, v := range want.body {
		if !strings.Contains(res.Body.String(), v) {
			t.Errorf("response body does not contain %s, got: %s", v, res.Body.String())
		}
	}
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	// ダミー画像データの準備
	dummyImageData, err := loadTestImage()
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	type wants struct {
		code int
		body string
	}
	cases := map[string]struct {
		args       map[string]string
		imageData  []byte
		setupMocks func(m *MockItemRepository)
		wants
	}{
		"ok: correctly insert item with new category": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			imageData: dummyImageData,
			setupMocks: func(m *MockItemRepository) {
				// カテゴリが見つからなかったら、新しいカテゴリを作成し、その後 Insert を呼び出す
				m.EXPECT().GetCategoryByName(gomock.Any(), gomock.Eq("phone")).
					Return(nil, errors.New("category not found"))
				m.EXPECT().InsertCategory(gomock.Any(), "phone").
					Return(&Category{ID: 1, Name: "phone"}, nil)
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).
					Return(nil)
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
			setupMocks: func(m *MockItemRepository) {
				// succeeded to insert with existing category
				// カテゴリが既に存在する場合
				m.EXPECT().GetCategoryByName(gomock.Any(), gomock.Eq("laptop")).Return(&Category{ID: 2, Name: "laptop"}, nil)
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
			setupMocks: func(m *MockItemRepository) {
				// カテゴリ取得失敗（データベースエラー）
				m.EXPECT().GetCategoryByName(gomock.Any(), gomock.Eq("tablet")).Return(nil, errors.New("database error"))
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
			setupMocks: func(m *MockItemRepository) {
				// カテゴリは正常に取得
				m.EXPECT().GetCategoryByName(gomock.Any(), gomock.Eq("phone")).Return(&Category{ID: 3, Name: "phone"}, nil)
				// アイテム挿入失敗
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("failed to insert item"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
				body: "failed to insert item",
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
			mockRepo := NewMockItemRepository(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			// テスト対象のハンドラーを作成
			h := &Handlers{
				imgDirPath: tempDir,
				itemRepo:   mockRepo,
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
				part, err := writer.CreateFormFile("image", "test.jpg")
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
func TestAddItemE2e(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	db, closers, err := setupDB(t)
	if err != nil {
		t.Fatalf("failed to set up database: %v", err)
	}
	t.Cleanup(func() {
		for _, c := range closers {
			c()
		}
	})

	// Create temporary directory for image files
	tempDir, err := os.MkdirTemp("", "test-images-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Prepare dummy image data
	dummyImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46} // JPEG header

	type wants struct {
		code int
		// Add database check expectations
		shouldExistInDB bool
	}

	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			wants: wants{
				code:            http.StatusOK,
				shouldExistInDB: true,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "",
				"category": "phone",
			},
			wants: wants{
				code:            http.StatusBadRequest,
				shouldExistInDB: false,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			// Setup the handler with the temp directory
			h := &Handlers{
				itemRepo:   &itemRepository{db: db},
				imgDirPath: tempDir,
			}

			// Create a multipart form request with image
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add form fields
			for k, v := range tt.args {
				err := writer.WriteField(k, v)
				if err != nil {
					t.Fatalf("failed to write field: %v", err)
				}
			}

			// Add image file
			part, err := writer.CreateFormFile("image", "test.jpg")
			if err != nil {
				t.Fatalf("failed to create form file: %v", err)
			}
			_, err = part.Write(dummyImageData)
			if err != nil {
				t.Fatalf("failed to write image data: %v", err)
			}

			err = writer.Close()
			if err != nil {
				t.Fatalf("failed to close writer: %v", err)
			}

			// Create the request with proper multipart form content type
			req := httptest.NewRequest("POST", "/items", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			// Check response status code
			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}

			// Skip further checks if we expect an error
			if tt.wants.code >= 400 {
				return
			}

			// Check if the response contains the item name (don't check for category in response)
			if !strings.Contains(rr.Body.String(), tt.args["name"]) {
				t.Errorf("response body does not contain item name %s, got: %s", tt.args["name"], rr.Body.String())
			}

			// STEP 6-4: check inserted data
			if tt.wants.shouldExistInDB {
				// Check category was inserted
				var categoryID int
				var categoryName string
				err := db.QueryRow("SELECT id, name FROM categories WHERE name = ?", tt.args["category"]).Scan(&categoryID, &categoryName)
				if err != nil {
					t.Errorf("find category in database failed: %v", err)
					return
				}

				if categoryName != tt.args["category"] {
					t.Errorf("expected category name %s, got %s", tt.args["category"], categoryName)
				}

				// Check item was inserted with correct category_id
				var itemName string
				var itemCategoryID int
				var imageName string

				err = db.QueryRow("SELECT name, category_id, image_name FROM items WHERE name = ?", tt.args["name"]).Scan(&itemName, &itemCategoryID, &imageName)
				if err != nil {
					t.Errorf("find item in database failed: %v", err)
					return
				}

				if itemName != tt.args["name"] {
					t.Errorf("expected item name %s, got %s", tt.args["name"], itemName)
				}

				if itemCategoryID != categoryID {
					t.Errorf("expected category_id %d, got %d", categoryID, itemCategoryID)
				}

				if imageName == "" {
					t.Errorf("expected non-empty image_name, got empty string")
				}

				// Check if image file exists in the temporary directory
				if _, err := os.Stat(filepath.Join(tempDir, imageName)); os.IsNotExist(err) {
					t.Errorf("image file %s does not exist in directory %s", imageName, tempDir)
				}
			}
		})
	}
}

func setupDB(t *testing.T) (db *sql.DB, closers []func(), e error) {
	t.Helper()

	defer func() {
		if e != nil {
			for _, c := range closers {
				c()
			}
		}
	}()

	// create a temporary file for e2e testing
	f, err := os.CreateTemp(".", "*.sqlite3")
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		f.Close()
		os.Remove(f.Name())
	})

	// set up tables
	db, err = sql.Open("sqlite3", f.Name())
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		db.Close()
	})

	// TODO: replace it with real SQL statements.
	cmd := `
	CREATE TABLE IF NOT EXISTS categories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name VARCHAR(255) NOT NULL
			);
			CREATE TABLE IF NOT EXISTS items (
				id INTEGER PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				category_id INTEGER,
				image_name VARCHAR(255) NOT NULL,
				FOREIGN KEY (category_id) REFERENCES categories(id)
			);
		`
	_, err = db.Exec(cmd)
	if err != nil {
		return nil, nil, err
	}

	return db, closers, nil
}

// **テスト用の画像を読み込む**
func loadTestImage() ([]byte, error) {
	// カレントディレクトリ取得
	_, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("カレントディレクトリを取得できません: %v", err)
	}

	// テストデータのディレクトリは `go/testdata/` に配置しているので、プロジェクトルートのパスを取得
	rootPath, err := filepath.Abs("../")
	if err != nil {
		return nil, fmt.Errorf("プロジェクトルートのパスを取得できません: %v", err)
	}

	// `testdata/default.jpg` の絶対パスを構築
	imagePath := filepath.Join(rootPath, "testdata", "default.jpg")

	// ファイルの存在をチェック
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("画像ファイルが見つかりません: %s", imagePath)
	}

	// 画像を読み込む
	dummyImageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("画像ファイルを読み込めません: %v", err)
	}

	// 画像が空でないかチェック
	if len(dummyImageData) == 0 {
		return nil, fmt.Errorf("画像データが空です: %s", imagePath)
	}
	return dummyImageData, nil
}
